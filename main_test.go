package slogsyslog

import (
	"encoding/json"
	"log/slog"
	"net"
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestMain(m *testing.M) {
	m.Run()

}

func TestForDebugLog(t *testing.T) {
	handler, err := getSlogSyslogHandler()

	assert.Equal(t, err, nil)

	logger := slog.New(handler)
	slog.SetDefault(logger)

	slog.Info("test rsys log", slog.String("key1", "val1"))
}

func getSlogSyslogHandler() (slog.Handler, error) {
	writer, err := net.Dial("udp", ":514")
	if err != nil {
		return nil, err
	}

	return Option{
		Converter: DefaultConverter,
		Level:     slog.LevelDebug,
		Writer:    writer,
		Marshaler: json.Marshal,
	}.NewSyslogHandler(), nil
}
