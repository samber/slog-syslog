package slogsyslog

import (
	"log/slog"
	"time"
)

type Priority int

const (
	Emergency Priority = iota
	Alert
	Crit
	Error
	Warning
	Notice
	Info
	Debug
)

type Message struct {
	Priority       Priority         `json:"priority"`
	Timestamp      time.Time        `json:"timestamp"`
	Hostname       string           `json:"hostname"`
	AppName        string           `json:"appname"`
	ProcessID      string           `json:"process_id"`
	MessageID      string           `json:"message_id"`
	StructuredData []StructuredData `json:"structured_data"`
	Message        string           `json:"message"`
}

func (m *Message) AddStructureData(ID string, Name string, Value string) {
	if m.StructuredData == nil {
		m.StructuredData = []StructuredData{}
	}
	for i, sd := range m.StructuredData {
		if sd.ID == ID {
			sd.Parameters = append(sd.Parameters, SDParam{Name: Name, Value: Value})
			m.StructuredData[i] = sd
			return
		}
	}

	m.StructuredData = append(m.StructuredData, StructuredData{
		ID: ID,
		Parameters: []SDParam{
			{
				Name:  Name,
				Value: Value,
			},
		},
	})
}

type SDParam struct {
	Name  string
	Value string
}

type StructuredData struct {
	ID         string
	Parameters []SDParam
}

func ConvertSlogToSyslogSeverity(lvl slog.Level) Priority {
	switch lvl {
	case slog.LevelDebug:
		return Debug
	case slog.LevelError:
		return Error
	case slog.LevelInfo:
		return Info
	case slog.LevelWarn:
		return Warning
	}

	return Emergency
}
