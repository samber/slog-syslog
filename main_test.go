package slogsyslog

import (
	"bytes"
	"io"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)

}

func TestHandler(t *testing.T) {
	// defer goleak.VerifyNone(t)
	w := &FakeWriter{}

	opt := Option{
		Level:     slog.LevelDebug,
		Writer:    w,
		Converter: FakeConverter,
	}

	handler := opt.NewSyslogHandler()
	slog.SetDefault(slog.New(handler))

	logMsg := "test"

	slog.Info(logMsg)
	time.Sleep(time.Second * 2)

	expectedByteLen := 88

	rb := make([]byte, w.Len())
	if _, err := w.Read(rb); err != nil {
		t.Errorf("Failed read logs from buffer: %s", err)
		return
	}

	logMsgAr := strings.Split(string(rb), " ")
	actualByteLen, err := strconv.Atoi(logMsgAr[0])
	if err != nil {
		t.Errorf("First word in log not word: %s", err)
		return
	}

	if actualByteLen != expectedByteLen {
		t.Errorf("Expected log len %d, actual %d", actualByteLen, expectedByteLen)
		return
	}

	actualLogMsg := logMsgAr[len(logMsgAr)-1]
	if actualLogMsg != logMsg {
		t.Errorf("Expected log message `%s`, actual `%s`", logMsg, actualLogMsg)
		return
	}
}

var _ io.Reader = (*FakeWriter)(nil)
var _ io.Writer = (*FakeWriter)(nil)

type FakeWriter struct {
	buf bytes.Buffer
	mut sync.RWMutex
}

func NewFakeWriter() *FakeWriter {
	return &FakeWriter{
		mut: sync.RWMutex{},
	}
}

func (w *FakeWriter) Read(buf []byte) (int, error) {
	w.mut.RLock()
	defer w.mut.RUnlock()

	return w.buf.Read(buf)
}

func (w *FakeWriter) Write(buf []byte) (int, error) {
	w.mut.Lock()
	defer w.mut.Unlock()

	n, err := w.buf.Write(buf)
	return n, err
}

func (w *FakeWriter) Len() int {
	w.mut.RLock()
	defer w.mut.RUnlock()

	return w.buf.Len()
}

func (w *FakeWriter) String() string {
	w.mut.RLock()
	defer w.mut.RUnlock()

	return w.buf.String()
}

func FakeConverter(addSource bool, replaceAttr func(groups []string, a slog.Attr) slog.Attr, loggerAttr []slog.Attr, groups []string, record *slog.Record) Message {
	return Message{
		AppName:   "appName",
		Hostname:  "hostName",
		Priority:  ConvertSlogToSyslogSeverity(record.Level),
		Timestamp: time.Time{},
		MessageID: uuid.New().String(),
		Message:   []byte(record.Message),
		ProcessID: "1",
	}
}
