package slogsyslog

import (
	"encoding"
	"fmt"
	"reflect"

	"golang.org/x/exp/slog"
)

type Converter func(loggerAttr []slog.Attr, record *slog.Record) map[string]any

func DefaultConverter(loggerAttr []slog.Attr, record *slog.Record) map[string]any {
	log := map[string]any{
		"timestamp": record.Time.UTC(),
		"level":     record.Level.String(),
		"message":   record.Message,
	}

	extra := attrsToValue(loggerAttr)

	record.Attrs(func(attr slog.Attr) {
		for k, v := range attrsToValue([]slog.Attr{attr}) {
			extra[k] = v
		}
	})

	if err, ok := extra["error"]; ok {
		log["error"] = err
		delete(extra, "error")
	}

	log["extra"] = extra

	return log
}

func attrsToValue(attrs []slog.Attr) map[string]any {
	log := map[string]any{}

	for i := range attrs {
		k, v := attrToValue(attrs[i])
		log[k] = v
	}

	return log
}

func attrToValue(attr slog.Attr) (string, any) {
	k := attr.Key
	v := attr.Value
	kind := v.Kind()

	switch kind {
	case slog.KindAny:
		if k == "error" {
			if err, ok := v.Any().(error); ok {
				return k, buildExceptions(err)
			}
		}

		return k, v.Any()
	case slog.KindLogValuer:
		return k, v.Any()
	case slog.KindGroup:
		return k, attrsToValue(v.Group())
	case slog.KindInt64:
		return k, v.Int64()
	case slog.KindUint64:
		return k, v.Uint64()
	case slog.KindFloat64:
		return k, v.Float64()
	case slog.KindString:
		return k, v.String()
	case slog.KindBool:
		return k, v.Bool()
	case slog.KindDuration:
		return k, v.Duration()
	case slog.KindTime:
		return k, v.Time()
	default:
		return k, anyValueToString(v)
	}
}

func anyValueToString(v slog.Value) string {
	if tm, ok := v.Any().(encoding.TextMarshaler); ok {
		data, err := tm.MarshalText()
		if err != nil {
			return ""
		}

		return string(data)
	}

	return fmt.Sprintf("%+v", v.Any())
}

func buildExceptions(err error) map[string]any {
	return map[string]any{
		"kind":  reflect.TypeOf(err).String(),
		"error": err.Error(),
		"stack": nil, // @TODO
	}
}
