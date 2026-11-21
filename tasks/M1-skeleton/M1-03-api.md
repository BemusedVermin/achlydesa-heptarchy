---
id: M1-03
title: "api: the plain-data boundary and its codec"
status: done
milestone: M1
kind: seam
depends_on: [M1-01]
discovered_by: null
size: S
---

# M1-03 · api: the plain-data boundary and its codec

## Goal

Commands, events and the view are plain serializable data in one package, so frontends, the harness and replay logs all speak the same types, and a protobuf boundary later is a mechanical translation.

## Context

- Tech design: "Designing for pivots" (the Commands and events, Choices come from the core, and View sections rows) and "Frontends → An out-of-process frontend".
- `AGENTS.md` rules 6 and 7.

## Deliverables

- `internal/core/api/api.go`: the core interfaces and identity types.
- `internal/core/api/view.go`: the view and its general parts.
- `internal/core/api/codec.go`: kind registry, JSON envelopes, replay logs.
- `internal/core/api/codec_test.go`

## Spec

Firm:

```go
// Command is a request from the player or the harness. Implementations are plain structs.
type Command interface{ CommandKind() string }

// Event is something that happened. Implementations are plain structs.
type Event interface{ EventKind() string }

// Loggable events produce a line in the player's message log.
type Loggable interface{ LogText() Text }

type UID uint64       // stable identity of an individual; 0 means none
type ContentID string // a namespaced content ID such as "material.clay"
type MapID uint32     // a map, hand-made or generated; 0 means none
type Hex struct { // axial hex coordinates, independent of the hex library
	Q int `json:"q"`
	R int `json:"r"`
}

// Text refers to player-facing text: a content text ID plus named arguments.
type Text struct {
	ID   ContentID         `json:"id"`
	Args map[string]string `json:"args,omitempty"`
}

// Rejected reports a command the core refused.
type Rejected struct {
	Command string `json:"command"` // the refused command's kind
	Reason  Text   `json:"reason"`
}
```

`Rejected` implements `Event` (kind `core.rejected`) and `Loggable`. The kind methods are named differently on purpose, so no type can be a command and an event by accident.

The view (`view.go`) holds display-ready strings; the core resolves text before building it. Start with only the general parts; later tasks each add a section as a pointer field:

```go
type View struct {
	Tick    int64     `json:"tick"`
	Options []Option  `json:"options"`
	Log     []LogLine `json:"log"`
	// Sections are added by later tasks: Clock (M1-08), Map and Party (M1-10), Combat (M2).
}

// Option is one thing the player can do right now. Frontends render Label and
// send Command when it's chosen; they never build commands themselves.
type Option struct {
	ID      string  `json:"id"`              // stable within a view, e.g. "step.nw"; slash-separated for menus, e.g. "act/strike/cut"
	Label   string  `json:"label"`
	Command Command `json:"-"`
	Enabled bool    `json:"enabled"`
	Reason  string  `json:"reason,omitempty"` // why it's disabled
}

type LogLine struct {
	Tick int64  `json:"tick"`
	Text string `json:"text"`
}
```

Codec (signatures are suggestions; behavior is firm):

```go
type Codec struct{ /* kind -> decoder */ }
func NewCodec() *Codec
func RegisterCommand[T Command](c *Codec) error
func RegisterEvent[T Event](c *Codec) error
func (c *Codec) EncodeCommand(cmd Command) ([]byte, error)
func (c *Codec) DecodeCommand(data []byte) (Command, error)
func (c *Codec) EncodeEvent(e Event) ([]byte, error)
func (c *Codec) DecodeEvent(data []byte) (Event, error)
func (c *Codec) WriteLog(w io.Writer, cmds []Command) error
func (c *Codec) ReadLog(r io.Reader) ([]Command, error)
```

- The envelope is `{"kind":"explore.step","data":{...}}`.
- A replay log is one command envelope per line (JSON Lines).
- Registration reads the kind from the zero value of `T` and rejects duplicates and kinds that don't match `^[a-z][a-z0-9_]*\.[a-z][a-z0-9_]*$`.
- Decoding returns `T` values, not pointers, and an error for unknown kinds.
- Register `Rejected` in `NewCodec`.

## Acceptance criteria

- [x] Round trip: a test command and event encode and decode to equal values.
- [x] Unknown kind, duplicate registration and a malformed kind each return an error.
- [x] A three-command log written with `WriteLog` reads back equal with `ReadLog`.
- [x] A `rapid` property: random field values on a test command survive a round trip.
- [x] `api` imports only the standard library.
- [x] `go tool task check` passes.

## Out of scope

- Any domain commands, events or view sections. Each owning task adds its own.
- Protobuf.

## Notes

- Implemented as specced. Codec signatures follow the suggestions; the kind registries are split per command and event, so a kind may be used once as each.
- Added the exported constant `RejectedKind` (`core.rejected`) so callers need not repeat the string.
- `ReadLog` skips blank lines and names the failing line number in errors.
- `rapid` is used in tests only; non-test `api` code imports only the standard library.
- Files touched outside the list: none. Follow-ups filed: none.
