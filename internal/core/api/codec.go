package api

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
)

var kindPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*\.[a-z][a-z0-9_]*$`)

// envelope is the wire form of a command or event.
type envelope struct {
	Kind string          `json:"kind"`
	Data json.RawMessage `json:"data"`
}

// Codec encodes and decodes commands and events as JSON envelopes
// {"kind":"explore.step","data":{...}}. Decoded values are the registered
// struct types, not pointers.
type Codec struct {
	commands kindTable[Command]
	events   kindTable[Event]
}

// kindTable maps the kinds of one family (commands or events) to their decoders.
type kindTable[T any] struct {
	family   string // "command" or "event", for error messages
	decoders map[string]func(json.RawMessage) (T, error)
}

func newKindTable[T any](family string) kindTable[T] {
	return kindTable[T]{family: family, decoders: map[string]func(json.RawMessage) (T, error){}}
}

func (t kindTable[T]) register(kind string, decode func(json.RawMessage) (T, error)) error {
	if err := checkKind(kind); err != nil {
		return err
	}
	if _, dup := t.decoders[kind]; dup {
		return fmt.Errorf("api: %s kind %q already registered", t.family, kind)
	}
	t.decoders[kind] = decode
	return nil
}

func (t kindTable[T]) decode(data []byte) (T, error) {
	var zero T
	env, err := decodeEnvelope(data)
	if err != nil {
		return zero, err
	}
	decode, ok := t.decoders[env.Kind]
	if !ok {
		return zero, fmt.Errorf("api: unknown %s kind %q", t.family, env.Kind)
	}
	value, err := decode(env.Data)
	if err != nil {
		return zero, fmt.Errorf("api: decode %s %q: %w", t.family, env.Kind, err)
	}
	return value, nil
}

// NewCodec returns a codec with the core's own types ([Rejected]) registered.
func NewCodec() *Codec {
	c := &Codec{commands: newKindTable[Command]("command"), events: newKindTable[Event]("event")}
	if err := RegisterEvent[Rejected](c); err != nil {
		panic(err) // a programmer bug: the built-in kind is valid and unique
	}
	return c
}

func checkKind(kind string) error {
	if !kindPattern.MatchString(kind) {
		return fmt.Errorf("api: kind %q does not match %s", kind, kindPattern)
	}
	return nil
}

func decodeInto[T any](data json.RawMessage) (T, error) {
	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		return value, err
	}
	return value, nil
}

// RegisterCommand registers T, reading its kind from the zero value. It fails
// on a malformed or duplicate kind.
func RegisterCommand[T Command](c *Codec) error {
	var zero T
	return c.commands.register(zero.CommandKind(), func(data json.RawMessage) (Command, error) {
		return decodeInto[T](data)
	})
}

// RegisterEvent registers T, reading its kind from the zero value. It fails on
// a malformed or duplicate kind.
func RegisterEvent[T Event](c *Codec) error {
	var zero T
	return c.events.register(zero.EventKind(), func(data json.RawMessage) (Event, error) {
		return decodeInto[T](data)
	})
}

func encode(kind string, value any) ([]byte, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("api: encode %q: %w", kind, err)
	}
	return json.Marshal(envelope{Kind: kind, Data: data})
}

func decodeEnvelope(data []byte) (envelope, error) {
	var env envelope
	if err := json.Unmarshal(data, &env); err != nil {
		return env, fmt.Errorf("api: bad envelope: %w", err)
	}
	return env, nil
}

// EncodeCommand encodes cmd as an envelope.
func (c *Codec) EncodeCommand(cmd Command) ([]byte, error) {
	if cmd == nil {
		return nil, errors.New("api: nil command")
	}
	return encode(cmd.CommandKind(), cmd)
}

// DecodeCommand decodes an envelope into its registered command type.
func (c *Codec) DecodeCommand(data []byte) (Command, error) {
	return c.commands.decode(data)
}

// EncodeEvent encodes e as an envelope.
func (c *Codec) EncodeEvent(e Event) ([]byte, error) {
	if e == nil {
		return nil, errors.New("api: nil event")
	}
	return encode(e.EventKind(), e)
}

// DecodeEvent decodes an envelope into its registered event type.
func (c *Codec) DecodeEvent(data []byte) (Event, error) {
	return c.events.decode(data)
}

// HasEvent reports whether an event kind is registered.
func (c *Codec) HasEvent(kind string) bool {
	_, ok := c.events.decoders[kind]
	return ok
}

// WriteLog writes a replay log: one command envelope per line (JSON Lines).
func (c *Codec) WriteLog(w io.Writer, cmds []Command) error {
	for i, cmd := range cmds {
		line, err := c.EncodeCommand(cmd)
		if err != nil {
			return fmt.Errorf("api: log entry %d: %w", i, err)
		}
		if _, err := w.Write(append(line, '\n')); err != nil {
			return err
		}
	}
	return nil
}

// ReadLog reads a replay log written by [Codec.WriteLog]. Blank lines are skipped.
func (c *Codec) ReadLog(r io.Reader) ([]Command, error) {
	var cmds []Command
	scanner := bufio.NewScanner(r)
	scanner.Buffer(nil, 1<<20)
	for lineNumber := 1; scanner.Scan(); lineNumber++ {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		cmd, err := c.DecodeCommand(line)
		if err != nil {
			return nil, fmt.Errorf("api: log line %d: %w", lineNumber, err)
		}
		cmds = append(cmds, cmd)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return cmds, nil
}
