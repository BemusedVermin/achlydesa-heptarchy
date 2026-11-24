package api

import (
	"bytes"
	"reflect"
	"slices"
	"strings"
	"testing"

	"pgregory.net/rapid"
)

type testCommand struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
	At    Hex    `json:"at"`
}

func (testCommand) CommandKind() string { return "test.command" }

type otherCommand struct {
	Target UID `json:"target"`
}

func (otherCommand) CommandKind() string { return "test.other" }

type testEvent struct {
	Who UID `json:"who"`
}

func (testEvent) EventKind() string { return "test.event" }

type badKindCommand struct{}

func (badKindCommand) CommandKind() string { return "BadKind" }

func newTestCodec(tb testing.TB) *Codec {
	tb.Helper()
	c := NewCodec()
	if err := RegisterCommand[testCommand](c); err != nil {
		tb.Fatal(err)
	}
	if err := RegisterCommand[otherCommand](c); err != nil {
		tb.Fatal(err)
	}
	if err := RegisterEvent[testEvent](c); err != nil {
		tb.Fatal(err)
	}
	return c
}

func TestRoundTrip(t *testing.T) {
	c := newTestCodec(t)

	cmd := testCommand{Name: "walk", Count: 3, At: Hex{Q: -1, R: 2}}
	data, err := c.EncodeCommand(cmd)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"kind":"test.command","data":{"name":"walk","count":3,"at":{"q":-1,"r":2}}}`
	if string(data) != want {
		t.Errorf("envelope = %s, want %s", data, want)
	}
	got, err := c.DecodeCommand(data)
	if err != nil {
		t.Fatal(err)
	}
	if got != Command(cmd) {
		t.Errorf("command = %#v, want %#v", got, cmd)
	}

	events := []Event{
		testEvent{Who: 7},
		Rejected{Command: "test.command", Reason: Text{ID: "text.no", Args: map[string]string{"why": "x"}}},
	}
	for _, event := range events {
		data, err := c.EncodeEvent(event)
		if err != nil {
			t.Fatal(err)
		}
		decoded, err := c.DecodeEvent(data)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(decoded, event) {
			t.Errorf("event = %#v, want %#v", decoded, event)
		}
	}
}

func TestRejectedIsLoggable(t *testing.T) {
	reason := Text{ID: "text.blocked"}
	var event Event = Rejected{Command: "x.y", Reason: reason}
	loggable, ok := event.(Loggable)
	if !ok || loggable.LogText().ID != reason.ID {
		t.Errorf("Rejected not loggable with its reason: %v", event)
	}
	if event.EventKind() != "core.rejected" {
		t.Errorf("kind = %q", event.EventKind())
	}
}

func TestErrors(t *testing.T) {
	c := newTestCodec(t)
	tests := map[string]func() error{
		"unknown command kind": func() error {
			_, err := c.DecodeCommand([]byte(`{"kind":"no.such","data":{}}`))
			return err
		},
		"unknown event kind": func() error {
			_, err := c.DecodeEvent([]byte(`{"kind":"no.such","data":{}}`))
			return err
		},
		"duplicate command": func() error { return RegisterCommand[testCommand](c) },
		"duplicate event":   func() error { return RegisterEvent[Rejected](c) },
		"malformed kind":    func() error { return RegisterCommand[badKindCommand](c) },
		"bad envelope": func() error {
			_, err := c.DecodeCommand([]byte(`not json`))
			return err
		},
		"bad data": func() error {
			_, err := c.DecodeCommand([]byte(`{"kind":"test.command","data":{"count":"x"}}`))
			return err
		},
		"nil command": func() error {
			_, err := c.EncodeCommand(nil)
			return err
		},
	}
	for name, run := range tests {
		t.Run(name, func(t *testing.T) {
			if err := run(); err == nil {
				t.Error("want an error, got nil")
			}
		})
	}
}

func TestLogRoundTrip(t *testing.T) {
	c := newTestCodec(t)
	cmds := []Command{
		testCommand{Name: "a", Count: 1},
		otherCommand{Target: 42},
		testCommand{Name: "b", At: Hex{Q: 3, R: -3}},
	}
	var buf bytes.Buffer
	if err := c.WriteLog(&buf, cmds); err != nil {
		t.Fatal(err)
	}
	if lines := bytes.Count(buf.Bytes(), []byte("\n")); lines != len(cmds) {
		t.Errorf("log has %d lines, want %d", lines, len(cmds))
	}
	got, err := c.ReadLog(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(got, cmds) {
		t.Errorf("log = %#v, want %#v", got, cmds)
	}
}

func TestReadLogReportsBadLine(t *testing.T) {
	c := newTestCodec(t)
	log := `{"kind":"test.other","data":{}}` + "\n" + `{"kind":"no.such","data":{}}` + "\n"
	_, err := c.ReadLog(strings.NewReader(log))
	if err == nil || !strings.Contains(err.Error(), "line 2") {
		t.Errorf("err = %v, want one naming line 2", err)
	}
}

func TestRoundTripProperty(t *testing.T) {
	c := newTestCodec(t)
	rapid.Check(t, func(t *rapid.T) {
		cmd := testCommand{
			Name:  rapid.String().Draw(t, "name"),
			Count: rapid.Int().Draw(t, "count"),
			At:    Hex{Q: rapid.Int().Draw(t, "q"), R: rapid.Int().Draw(t, "r")},
		}
		data, err := c.EncodeCommand(cmd)
		if err != nil {
			t.Fatal(err)
		}
		got, err := c.DecodeCommand(data)
		if err != nil {
			t.Fatal(err)
		}
		if got != Command(cmd) {
			t.Fatalf("got %#v, want %#v", got, cmd)
		}
	})
}
