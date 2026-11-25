package timeline

import (
	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// Suspend takes a body off the timeline and cancels its pending actions with the reason. The body
// keeps everything else, its position included. A body that is not on the timeline is left alone.
func Suspend(w *sim.World, body ecs.Entity, reason api.ContentID) {
	st := stateOf(w)
	if !w.ECS.Alive(body) || !st.actors.Has(body) {
		return
	}
	for _, pending := range PendingOf(w, w.UIDOf(body)) {
		Cancel(w, pending, reason)
	}
	actor := *st.actors.Get(body)
	st.actors.Remove(body)
	delete(st.queued, body)
	delete(st.sleeping, body)
	if st.hasCurrent && st.current == body {
		st.hasCurrent = false
	}
	sim.Put(w, body, comp.Dormant(actor))
}

// Resume puts a suspended body back on the timeline under its old controller. Any other body is
// left alone.
func Resume(w *sim.World, body ecs.Entity) {
	dormant := ecs.NewMap[comp.Dormant](w.ECS)
	if !w.ECS.Alive(body) || !dormant.Has(body) {
		return
	}
	st := stateOf(w)
	saved := *dormant.Get(body)
	dormant.Remove(body)
	st.actors.Add(body, &comp.Actor{Controller: saved.Controller, WaitUntil: max(saved.WaitUntil, w.Now)})
	st.schedule(w, body)
}
