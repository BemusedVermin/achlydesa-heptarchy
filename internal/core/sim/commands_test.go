package sim

import (
	"slices"
	"testing"
)

type testCommand struct{ Cost int64 }

func (testCommand) CommandKind() string { return "test.cmd" }

type otherCommand struct{}

func (otherCommand) CommandKind() string { return "test.other" }

func TestCommandsDispatch(t *testing.T) {
	w := NewWorld(1)
	err := Handle(w, func(_ *World, cmd testCommand) (int64, error) { return cmd.Cost, nil })
	if err != nil {
		t.Fatal(err)
	}
	cost, err := w.Commands.Dispatch(w, testCommand{Cost: 12})
	if err != nil || cost != 12 {
		t.Fatalf("Dispatch() = %d, %v; want 12, nil", cost, err)
	}
	if _, err := w.Commands.Dispatch(w, otherCommand{}); err == nil {
		t.Error("unknown kind should error")
	}
	if got := w.Commands.Kinds(); !slices.Equal(got, []string{"test.cmd"}) {
		t.Errorf("Kinds() = %v", got)
	}
}

func TestHandleRegistersWithCodecAndRejectsDuplicates(t *testing.T) {
	w := NewWorld(1)
	handler := func(*World, testCommand) (int64, error) { return 0, nil }
	if err := Handle(w, handler); err != nil {
		t.Fatal(err)
	}
	if err := Handle(w, handler); err == nil {
		t.Error("duplicate Handle should error")
	}
	data, err := w.Codec.EncodeCommand(testCommand{Cost: 3})
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := w.Codec.DecodeCommand(data)
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := decoded.(testCommand); !ok || got.Cost != 3 {
		t.Errorf("round trip = %#v", decoded)
	}
}
