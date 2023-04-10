package slogsyslog

import (
	"context"
	"encoding/json"
	"fmt"
	"log/syslog"

	"golang.org/x/exp/slog"
)

type Option struct {
	// log level (default: debug)
	Level slog.Leveler

	// connection to syslog server
	Writer *syslog.Writer

	// optional: customize json payload builder
	Converter Converter
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

	message := converter(h.attrs, &record)

	bytes, err := json.Marshal(message)
	if err != nil {
		return err
	}

	switch record.Level {
	case slog.LevelDebug:
		return h.option.Writer.Debug(string(bytes))
	case slog.LevelInfo:
		return h.option.Writer.Info(string(bytes))
	case slog.LevelWarn:
		return h.option.Writer.Warning(string(bytes))
	case slog.LevelError:
		return h.option.Writer.Err(string(bytes))
	}

	return fmt.Errorf("slog-syslog: unexpected log level")
}

func (h *SyslogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &SyslogHandler{
		option: h.option,
		attrs:  appendAttrsToGroup(h.groups, h.attrs, attrs),
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
