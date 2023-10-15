package slogsyslog

import (
	"context"
	"encoding/json"
	"io"

	"log/slog"

	slogcommon "github.com/samber/slog-common"
)

const ceePrefix = "@cee: "

type Option struct {
	// log level (default: debug)
	Level slog.Leveler

	// connection to syslog server
	Writer io.Writer

	// optional: customize json payload builder
	Converter Converter

	// optional: see slog.HandlerOptions
	AddSource   bool
	ReplaceAttr func(groups []string, a slog.Attr) slog.Attr
}

func (o Option) NewSyslogHandler() slog.Handler {
	if o.Level == nil {
		o.Level = slog.LevelDebug
	}

	if o.Writer == nil {
		panic("missing syslog server connections")
	}

	return &SyslogHandler{
		option: o,
		attrs:  []slog.Attr{},
		groups: []string{},
	}
}

var _ slog.Handler = (*SyslogHandler)(nil)

type SyslogHandler struct {
	option Option
	attrs  []slog.Attr
	groups []string
}

func (h *SyslogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.option.Level.Level()
}

func (h *SyslogHandler) Handle(ctx context.Context, record slog.Record) error {
	converter := DefaultConverter
	if h.option.Converter != nil {
		converter = h.option.Converter
	}

	message := converter(h.option.AddSource, h.option.ReplaceAttr, h.attrs, h.groups, &record)

	bytes, err := json.Marshal(message)
	if err != nil {
		return err
	}

	_, err = h.option.Writer.Write(append([]byte(ceePrefix), bytes...))
	return err
}

func (h *SyslogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &SyslogHandler{
		option: h.option,
		attrs:  slogcommon.AppendAttrsToGroup(h.groups, h.attrs, attrs...),
		groups: h.groups,
	}
}

func (h *SyslogHandler) WithGroup(name string) slog.Handler {
	return &SyslogHandler{
		option: h.option,
		attrs:  h.attrs,
		groups: append(h.groups, name),
	}
}
