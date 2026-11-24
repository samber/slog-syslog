package slogsyslog

import (
	"log/slog"
	"reflect"
	"strings"
	"testing"
	"time"
)

type chanWriter chan string

func (w chanWriter) Write(p []byte) (int, error) {
	w <- string(p)
	return len(p), nil
}

func TestSyslogHandler_WithGroupSiblings(t *testing.T) {
	root := &SyslogHandler{
		option: Option{},
		attrs:  []slog.Attr{},
		groups: make([]string, 0, 8),
	}

	parent := root.WithGroup("a").WithGroup("b").WithGroup("c")

	first := parent.WithGroup("d").(*SyslogHandler)
	second := parent.WithGroup("e").(*SyslogHandler)

	if want := []string{"a", "b", "c", "d"}; !reflect.DeepEqual(first.groups, want) {
		t.Fatalf("first sibling groups = %v, want %v", first.groups, want)
	}
	if want := []string{"a", "b", "c", "e"}; !reflect.DeepEqual(second.groups, want) {
		t.Fatalf("second sibling groups = %v, want %v", second.groups, want)
	}
}

func TestSyslogHandler_SiblingLoggersKeepTheirGroup(t *testing.T) {
	out := make(chanWriter, 2)
	handler := Option{Writer: out}.NewSyslogHandler()

	// Three levels of WithGroup grow the slice with spare capacity.
	parent := slog.New(handler).WithGroup("a").WithGroup("b").WithGroup("c")
	first := parent.WithGroup("d")
	second := parent.WithGroup("e")

	first.Info("first", "k", "v")
	second.Info("second", "k", "v")

	for i := 0; i < 2; i++ {
		select {
		case line := <-out:
			switch {
			case strings.Contains(line, `"first"`):
				if !strings.Contains(line, `"d"`) || strings.Contains(line, `"e"`) {
					t.Fatalf("first logger lost its group: %s", line)
				}
			case strings.Contains(line, `"second"`):
				if !strings.Contains(line, `"e"`) || strings.Contains(line, `"d"`) {
					t.Fatalf("second logger lost its group: %s", line)
				}
			default:
				t.Fatalf("unexpected line: %s", line)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("timed out waiting for syslog output")
		}
	}
}
