package save

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

// Format names the kind of file, so other JSON is rejected.
const Format = "heptarchy-save"

// CurrentVersion is the version Write produces. Raise it together with a migration from the old one.
const CurrentVersion = 6

// Header says what the file is and what it was made from.
type Header struct {
	Format      string `json:"format"`
	Version     int    `json:"version"`
	ContentHash string `json:"content_hash"`
	Seed        uint64 `json:"seed"`
	Tick        int64  `json:"tick"`
}

// Edit is one cell of the world changed from what its source supplied. The terrain is named by
// content ID, which is resolved against content on load.
type Edit struct {
	At      api.Cell      `json:"at"`
	Terrain api.ContentID `json:"terrain"`
}

// Exports is how many exports have begun and the day of the next one.
type Exports struct {
	Count   int `json:"count"`
	NextDay int `json:"next_day"`
}

// Chunk is the key of one chunk of the world.
type Chunk struct {
	Q     int `json:"q"`
	R     int `json:"r"`
	Depth int `json:"depth"`
}

// Snapshot is a whole game, as plain data.
type Snapshot struct {
	Header Header `json:"header"`
	// World is the ECS world as ark-serde wrote it, canonicalized by the caller.
	World json.RawMessage `json:"world"`
	// RNG is the state of each named random stream.
	RNG     map[string][]byte `json:"rng"`
	NextUID api.UID           `json:"next_uid"`
	// Timeline is the queued entries and the deciding body, which live outside the ECS components.
	Timeline timeline.Saved `json:"timeline"`
	// Exports is the export schedule's state.
	Exports Exports `json:"exports"`
	// Edits lists the world's edits, sorted by depth, R, Q. The world itself is rebuilt from content
	// on load.
	Edits []Edit `json:"edits"`
	// Populated lists the chunks whose starting bodies have been spawned, sorted by depth, R, Q.
	Populated []Chunk `json:"populated"`
	// History is every command applied so far, in the codec's envelope form.
	History []json.RawMessage `json:"history"`
	// Log is the message log the view shows.
	Log []api.LogLine `json:"log"`
}

// Write encodes s as one JSON document. Equal snapshots give equal bytes.
func Write(w io.Writer, s Snapshot) error {
	if s.Header.Format != Format || s.Header.Version != CurrentVersion {
		return fmt.Errorf("save: write: header must be format %q version %d", Format, CurrentVersion)
	}
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(s); err != nil {
		return fmt.Errorf("save: write: %w", err)
	}
	if _, err := w.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("save: write: %w", err)
	}
	return nil
}

// Read decodes a save, upgrading older versions through the migration chain. It rejects other
// formats and versions newer than CurrentVersion.
func Read(r io.Reader) (Snapshot, error) {
	return read(r, CurrentVersion, migrations)
}

// CheckContent returns a warning when the content the save was made with differs from hash. Content
// changes all the time in development, so it is a warning and not an error.
func (h Header) CheckContent(hash string) []string {
	if h.ContentHash == hash {
		return nil
	}
	return []string{fmt.Sprintf("save: made with content %.12s, loading with content %.12s", h.ContentHash, hash)}
}

func read(r io.Reader, current int, chain map[int]Migration) (Snapshot, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return Snapshot{}, fmt.Errorf("save: read: %w", err)
	}
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(data, &doc); err != nil {
		return Snapshot{}, fmt.Errorf("save: read: not a JSON object: %w", err)
	}
	var header Header
	if raw, ok := doc["header"]; !ok {
		return Snapshot{}, errors.New("save: read: no header")
	} else if err := json.Unmarshal(raw, &header); err != nil {
		return Snapshot{}, fmt.Errorf("save: read: header: %w", err)
	}
	switch {
	case header.Format != Format:
		return Snapshot{}, fmt.Errorf("save: read: format %q is not %q", header.Format, Format)
	case header.Version > current:
		return Snapshot{}, fmt.Errorf("save: read: version %d is newer than this game understands (%d)",
			header.Version, current)
	case header.Version < 0:
		return Snapshot{}, fmt.Errorf("save: read: version %d is not valid", header.Version)
	}
	if err := upgrade(doc, header.Version, current, chain); err != nil {
		return Snapshot{}, err
	}
	migrated, err := json.Marshal(doc)
	if err != nil {
		return Snapshot{}, fmt.Errorf("save: read: %w", err)
	}
	var snapshot Snapshot
	if err := json.Unmarshal(migrated, &snapshot); err != nil {
		return Snapshot{}, fmt.Errorf("save: read: %w", err)
	}
	return snapshot, nil
}
