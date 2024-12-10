package slogsyslog

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	slogcommon "github.com/samber/slog-common"
)

type Option struct {
	Level           slog.Leveler
	Writer          io.Writer
	Converter       Converter
	Marshaler       func(v Message) ([]byte, error)
	ReplaceAttr     func(groups []string, a slog.Attr) slog.Attr
	AttrFromContext []func(ctx context.Context) []slog.Attr
	AddSource       bool
}

func (o Option) NewSyslogHandler() slog.Handler {
	if o.Level == nil {
		o.Level = slog.LevelDebug
	}

	if o.Writer == nil {
		panic("missing syslog server connections")
	}

	if o.Converter == nil {
		o.Converter = DefaultConverter
	}

	if o.Marshaler == nil {
		o.Marshaler = marshalBinary
	}

	if o.AttrFromContext == nil {
		o.AttrFromContext = []func(ctx context.Context) []slog.Attr{}
	}

	return &SyslogHandler{
		option: o,
		attrs:  []slog.Attr{},
		groups: []string{},
	}
}

var _ slog.Handler = (*SyslogHandler)(nil)

type SyslogHandler struct {
	attrs  []slog.Attr
	groups []string
	option Option
}

func (h *SyslogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.option.Level.Level()
}

func (h *SyslogHandler) Handle(ctx context.Context, record slog.Record) error {
	fromContext := slogcommon.ContextExtractor(ctx, h.option.AttrFromContext)
	message := h.option.Converter(h.option.AddSource, h.option.ReplaceAttr, append(h.attrs, fromContext...), h.groups, &record)
	bytes, err := h.option.Marshaler(message)
	if err != nil {
		bytes = []byte{}
	}
	// non-blocking
	go func() {
		_, _ = fmt.Fprintf(h.option.Writer, "%d %s", len(bytes), bytes)
	}()

	return nil
}

func (h *SyslogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &SyslogHandler{
		option: h.option,
		attrs:  slogcommon.AppendAttrsToGroup(h.groups, h.attrs, attrs...),
		groups: h.groups,
	}
}

func (h *SyslogHandler) WithGroup(name string) slog.Handler {
	// https://cs.opensource.google/go/x/exp/+/46b07846:slog/handler.go;l=247
	if name == "" {
		return h
	}

	return &SyslogHandler{
		option: h.option,
		attrs:  h.attrs,
		groups: append(h.groups, name),
	}
}
