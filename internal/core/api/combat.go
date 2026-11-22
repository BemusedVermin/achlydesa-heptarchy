package api

// Kinds of the events combat primitives publish.
const (
	PartDamagedKind  = "combat.part_damaged"
	PartSeveredKind  = "combat.part_severed"
	PartAttachedKind = "combat.part_attached"
	StatusAddedKind  = "combat.status_added"
)

// Causes of a [PartSevered].
const (
	SeveredWorn = "worn" // its integrity ran out
	SeveredCut  = "cut"  // an effect took it off
)

// Where an attached part came from, in [PartAttached].From.
const (
	AttachedFromPack    = "pack"    // a piece the party carried
	AttachedFromSevered = "severed" // a severed part lying where the body stands
)

// PartDamaged reports that a part lost integrity.
type PartDamaged struct {
	Body      UID `json:"body"`
	Part      UID `json:"part"`
	Amount    int `json:"amount"`    // the integrity it actually lost
	Integrity int `json:"integrity"` // what it has left
	// Type is the damage type the action chose, or empty when it chose none.
	Type ContentID `json:"type,omitempty"`
	// Source is the body that acted, or zero when nothing did.
	Source UID   `json:"source,omitempty"`
	Tick   int64 `json:"tick"`
}

// EventKind implements [Event].
func (PartDamaged) EventKind() string { return PartDamagedKind }

// PartSevered reports that a part came off its body. It stays in the world, where the body stands.
type PartSevered struct {
	Body  UID    `json:"body"`
	Part  UID    `json:"part"`
	Cause string `json:"cause"` // SeveredWorn or SeveredCut
	// Source is the body that acted, or zero when nothing did.
	Source UID   `json:"source,omitempty"`
	Tick   int64 `json:"tick"`
}

// EventKind implements [Event].
func (PartSevered) EventKind() string { return PartSeveredKind }

// PartAttached reports that a part was fitted onto a body, which gains its channel.
type PartAttached struct {
	Body UID    `json:"body"`
	Part UID    `json:"part"`
	From string `json:"from"` // AttachedFromPack or AttachedFromSevered
	// Source is the body that acted, or zero when nothing did.
	Source UID   `json:"source,omitempty"`
	Tick   int64 `json:"tick"`
}

// EventKind implements [Event].
func (PartAttached) EventKind() string { return PartAttachedKind }

// StatusAdded reports that a body came under a status, or had it extended.
type StatusAdded struct {
	Body   UID       `json:"body"`
	Status ContentID `json:"status"`
	Until  int64     `json:"until"` // the tick it expires
	// Source is the body that acted, or zero when nothing did.
	Source UID   `json:"source,omitempty"`
	Tick   int64 `json:"tick"`
}

// EventKind implements [Event].
func (StatusAdded) EventKind() string { return StatusAddedKind }

// Subjects implements [Subjected]: the body is the target, and the actor the source.
func (e PartDamaged) Subjects() map[string]UID { return partSubjects(e.Body, e.Part, e.Source) }

// Subjects implements [Subjected]: the body is the target, and the actor the source.
func (e PartSevered) Subjects() map[string]UID { return partSubjects(e.Body, e.Part, e.Source) }

// Subjects implements [Subjected]: the body is the target, and the actor the source.
func (e PartAttached) Subjects() map[string]UID { return partSubjects(e.Body, e.Part, e.Source) }

// Subjects implements [Subjected]: the body is the target, and the actor the source.
func (e StatusAdded) Subjects() map[string]UID {
	return withSource(map[string]UID{SubjectTarget: e.Body}, e.Source)
}

// Choices implements [Chosen]: the damage type, when there is one.
func (e PartDamaged) Choices() map[string]string {
	if e.Type == "" {
		return nil
	}
	return map[string]string{"type": string(e.Type)}
}

func partSubjects(body, part, source UID) map[string]UID {
	return withSource(map[string]UID{SubjectTarget: body, SubjectPart: part}, source)
}

// withSource adds the acting body to the subjects when there was one.
func withSource(subjects map[string]UID, source UID) map[string]UID {
	if source != 0 {
		subjects[SubjectSource] = source
	}
	return subjects
}
