package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"log/slog"

	slogsyslog "github.com/axon-expert/slog-syslog-rfc5424"
)

func main() {
	// ncat -u -l 9999 -k
	writer, err := net.Dial("udp", "localhost:514")
	if err != nil {
		log.Fatal(err)
	}

	logger := slog.New(slogsyslog.Option{Level: slog.LevelDebug, Writer: writer}.NewSyslogHandler("test app", "localhost"))
	logger = logger.With("release", "v1.0.0")

	logger.
		With(
			slog.Group("user",
				slog.String("id", "user-123"),
				slog.Time("created_at", time.Now().AddDate(0, 0, -1)),
			),
		).
		With("environment", "dev").
		With("error", fmt.Errorf("an error").Error()).
		Error("A message")

	time.Sleep(time.Second * 3)
}
