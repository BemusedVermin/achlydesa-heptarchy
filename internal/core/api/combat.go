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
	Body      UID   `json:"body"`
	Part      UID   `json:"part"`
	Amount    int   `json:"amount"`    // the integrity it actually lost
	Integrity int   `json:"integrity"` // what it has left
	Tick      int64 `json:"tick"`
}

// EventKind implements [Event].
func (PartDamaged) EventKind() string { return PartDamagedKind }

// PartSevered reports that a part came off its body. It stays in the world, where the body stands.
type PartSevered struct {
	Body  UID    `json:"body"`
	Part  UID    `json:"part"`
	Cause string `json:"cause"` // SeveredWorn or SeveredCut
	Tick  int64  `json:"tick"`
}

// EventKind implements [Event].
func (PartSevered) EventKind() string { return PartSeveredKind }

// PartAttached reports that a part was fitted onto a body, which gains its channel.
type PartAttached struct {
	Body UID    `json:"body"`
	Part UID    `json:"part"`
	From string `json:"from"` // AttachedFromPack or AttachedFromSevered
	Tick int64  `json:"tick"`
}

// EventKind implements [Event].
func (PartAttached) EventKind() string { return PartAttachedKind }

// StatusAdded reports that a body came under a status, or had it extended.
type StatusAdded struct {
	Body   UID       `json:"body"`
	Status ContentID `json:"status"`
	Until  int64     `json:"until"` // the tick it expires
	Tick   int64     `json:"tick"`
}

// EventKind implements [Event].
func (StatusAdded) EventKind() string { return StatusAddedKind }
