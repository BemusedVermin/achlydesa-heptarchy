package save

import (
	"encoding/json"
	"fmt"
)

// Migration upgrades a raw document one version, in place. It does not touch the header: upgrade
// sets the new version after each step.
type Migration func(doc map[string]json.RawMessage) error

// migrations[n] upgrades a version n document to version n+1. Every version from 0 below
// CurrentVersion that can still be loaded has an entry; there is none yet because version 1 is the
// first.
var migrations = map[int]Migration{}

// upgrade runs the migrations from version from up to version to, one at a time.
func upgrade(doc map[string]json.RawMessage, from, to int, chain map[int]Migration) error {
	for version := from; version < to; version++ {
		step, ok := chain[version]
		if !ok {
			return fmt.Errorf("save: no migration from version %d to %d", version, version+1)
		}
		if err := step(doc); err != nil {
			return fmt.Errorf("save: migrate version %d to %d: %w", version, version+1, err)
		}
		var header map[string]json.RawMessage
		if err := json.Unmarshal(doc["header"], &header); err != nil {
			return fmt.Errorf("save: migrate version %d to %d: header: %w", version, version+1, err)
		}
		header["version"] = json.RawMessage(fmt.Sprint(version + 1))
		raw, err := json.Marshal(header)
		if err != nil {
			return fmt.Errorf("save: migrate: %w", err)
		}
		doc["header"] = raw
	}
	return nil
}
