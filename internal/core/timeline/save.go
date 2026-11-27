package timeline

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// Saved is the part of the timeline that lives outside the ECS components: who is queued for when,
// and who is deciding now. Bodies are named by UID. It is plain data so a save can hold it.
type Saved struct {
	// Line lists the live entries, sorted by tick then UID.
	Line []SavedSlot `json:"line"`
	// Current is the body deciding now, or 0 when there is none.
	Current api.UID `json:"current"`
	// SystemsDue says the last entry has not had its per-entry systems run yet.
	SystemsDue bool `json:"systems_due"`
	// Sleeping lists, by UID, the bodies whose last decision only waited.
	Sleeping []api.UID `json:"sleeping,omitempty"`
	// Changes is the counter that tells whether a controller did anything.
	Changes int64 `json:"changes"`
	// Seq is the last Seq handed to a pending action, including those that have landed since.
	Seq uint64 `json:"seq"`
}

// SavedSlot is one live timeline entry.
type SavedSlot struct {
	Tick int64   `json:"tick"`
	UID  api.UID `json:"uid"`
}

// Export returns the timeline's state for a save. Stale entries are left out.
func Export(w *sim.World) Saved {
	st := stateOf(w)
	saved := Saved{Line: []SavedSlot{}, SystemsDue: st.systemsDue, Changes: st.changes, Seq: st.seq}
	for _, s := range st.line {
		if st.live(w, s) {
			saved.Line = append(saved.Line, SavedSlot{Tick: s.tick, UID: s.uid})
		}
	}
	slices.SortFunc(saved.Line, func(a, b SavedSlot) int {
		return cmp.Or(cmp.Compare(a.Tick, b.Tick), cmp.Compare(a.UID, b.UID))
	})
	for e := range st.sleeping {
		if st.live(w, slot{tick: st.queued[e], e: e}) {
			saved.Sleeping = append(saved.Sleeping, w.UIDOf(e))
		}
	}
	slices.Sort(saved.Sleeping)
	if st.hasCurrent {
		saved.Current = w.UIDOf(st.current)
	}
	return saved
}

// Import replaces the timeline's queue and current body with saved. The components must already be
// restored and the UID index rebuilt.
func Import(w *sim.World, saved Saved) error {
	st := stateOf(w)
	st.line = st.line[:0]
	st.queued = map[ecs.Entity]int64{}
	st.sleeping = map[ecs.Entity]bool{}
	for _, s := range saved.Line {
		e, ok := w.ByUID(s.UID)
		if !ok || !st.actors.Has(e) {
			return fmt.Errorf("timeline: saved entry for UID %d has no actor", s.UID)
		}
		st.line = append(st.line, slot{tick: s.Tick, uid: s.UID, e: e})
		st.queued[e] = s.Tick
	}
	for _, uid := range saved.Sleeping {
		e, ok := w.ByUID(uid)
		if !ok || !st.actors.Has(e) {
			return fmt.Errorf("timeline: saved sleeping body %d has no actor", uid)
		}
		st.sleeping[e] = true
	}
	// Sorted by (tick, uid) is already a valid min-heap.
	st.current, st.hasCurrent = ecs.Entity{}, false
	if saved.Current != 0 {
		e, ok := w.ByUID(saved.Current)
		if !ok {
			return fmt.Errorf("timeline: saved current body %d does not exist", saved.Current)
		}
		st.current, st.hasCurrent = e, true
	}
	st.systemsDue, st.changes, st.seq = saved.SystemsDue, saved.Changes, saved.Seq
	return nil
}
