package main

import (
	"fmt"
	"log"
	"net"
	"time"

	slogsyslog "github.com/samber/slog-syslog"
	"golang.org/x/exp/slog"
)

func main() {
	// ncat -u -l 9999 -k
	writer, err := net.Dial("udp", "localhost:9999")
	if err != nil {
		log.Fatal(err)
	}

	logger := slog.New(slogsyslog.Option{Level: slog.LevelDebug, Writer: writer}.NewSyslogHandler())
	logger = logger.With("release", "v1.0.0")

	logger.
		With(
			slog.Group("user",
				slog.String("id", "user-123"),
				slog.Time("created_at", time.Now().AddDate(0, 0, -1)),
			),
		).
		With("environment", "dev").
		With("error", fmt.Errorf("an error")).
		Error("A message")
}
