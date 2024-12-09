package slogsyslog

import (
	"bytes"
	"fmt"
	"log/slog"
	"time"
)

type Priority int

const rfc3339Micro = "2006-01-02T15:04:05.999999Z07:00"

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
	Timestamp      time.Time
	Hostname       string
	AppName        string
	ProcessID      string
	MessageID      string
	StructuredData []StructuredData
	Message        []byte
	Priority       Priority
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

func marshalBinary(m any) ([]byte, error) {
	var obj Message

	if v, ok := m.(Message); ok {
		obj = v
	} else {
		obj = Message{
			Message: []byte("non correct message"),
		}
	}

	b := bytes.NewBuffer(nil)
	fmt.Fprintf(b, "<%d>1 %s %s %s %s %s ",
		obj.Priority,
		obj.Timestamp.Format(rfc3339Micro),
		nilify(obj.Hostname),
		nilify(obj.AppName),
		nilify(obj.ProcessID),
		nilify(obj.MessageID))

	if len(obj.StructuredData) == 0 {
		fmt.Fprint(b, "-")
	}
	for _, sdElement := range obj.StructuredData {
		fmt.Fprintf(b, "[%s", sdElement.ID)
		for _, sdParam := range sdElement.Parameters {
			fmt.Fprintf(b, " %s=\"%s\"", sdParam.Name,
				escapeSDParam(sdParam.Value))
		}
		fmt.Fprintf(b, "]")
	}

	if len(obj.Message) > 0 {
		fmt.Fprint(b, " ")
		b.Write(obj.Message)
	}
	return b.Bytes(), nil
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

func nilify(x string) string {
	if x == "" {
		return "-"
	}
	return x
}

func escapeSDParam(s string) string {
	escapeCount := 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '\\', '"', ']':
			escapeCount++
		}
	}
	if escapeCount == 0 {
		return s
	}

	t := make([]byte, len(s)+escapeCount)
	j := 0
	for i := 0; i < len(s); i++ {
		switch c := s[i]; c {
		case '\\', '"', ']':
			t[j] = '\\'
			t[j+1] = c
			j += 2
		default:
			t[j] = s[i]
			j++
		}
	}
	return string(t)
}
