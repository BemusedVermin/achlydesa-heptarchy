package api

// Kinds of the events combat primitives publish.
const (
	PartDamagedKind  = "combat.part_damaged"
	PartSeveredKind  = "combat.part_severed"
	PartAttachedKind = "combat.part_attached"
	StatusAddedKind  = "combat.status_added"
	AlertStartedKind = "combat.alert_started"
	AlertEndedKind   = "combat.alert_ended"
	BodyDiedKind     = "combat.body_died"
	BodyDownedKind   = "combat.body_downed"
	PartyWipedKind   = "combat.party_wiped"
)

// Causes of a [PartSevered].
const (
	SeveredWorn = "worn" // its integrity ran out
	SeveredCut  = "cut"  // an effect took it off
)

// Where an attached part came from, in [PartAttached].From.
const (
	AttachedFromInventory = "inventory" // a piece the party carried
	AttachedFromSevered   = "severed"   // a severed part lying where the body stands
)

// PartDamaged reports that a part lost integrity.
type PartDamaged struct {
	Body      UID `json:"body"`
	Part      UID `json:"part"`
	Amount    int `json:"amount"`    // the integrity it actually lost
	Integrity int `json:"integrity"` // what it has left
	// Chosen is every choice of the action that dealt the damage, such as "type": "type.cut"; the
	// Choices method returns it, so trigger formulas see it as chosen.
	Chosen map[string]string `json:"choices,omitempty"`
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
	From string `json:"from"` // AttachedFromInventory or AttachedFromSevered
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

// AlertStarted reports that the controlled party has a hostile body in sight.
type AlertStarted struct {
	Party UID `json:"party"`
	// Hostile is the first hostile seen, the one with the lowest UID, and Name its rendered name.
	Hostile UID    `json:"hostile"`
	Name    string `json:"name"`
	Tick    int64  `json:"tick"`
}

// EventKind implements [Event].
func (AlertStarted) EventKind() string { return AlertStartedKind }

// LogText implements [Loggable].
func (a AlertStarted) LogText() Text {
	return Text{ID: "text.combat.alert_started", Args: map[string]string{"name": a.Name}}
}

// AlertEnded reports that the party has been out of sight of every hostile for the calm spell.
type AlertEnded struct {
	Party UID   `json:"party"`
	Tick  int64 `json:"tick"`
}

// EventKind implements [Event].
func (AlertEnded) EventKind() string { return AlertEndedKind }

// LogText implements [Loggable].
func (AlertEnded) LogText() Text { return Text{ID: "text.combat.alert_ended"} }

// BodyDied reports that a body died. It stays where it fell, as a corpse.
type BodyDied struct {
	Body UID `json:"body"`
	// Name is the body's rendered name.
	Name string `json:"name"`
	Tick int64  `json:"tick"`
}

// EventKind implements [Event].
func (BodyDied) EventKind() string { return BodyDiedKind }

// LogText implements [Loggable].
func (b BodyDied) LogText() Text {
	return Text{ID: "text.combat.body_died", Args: map[string]string{"name": b.Name}}
}

// BodyDowned reports that a protected body fell instead of dying.
type BodyDowned struct {
	Body UID `json:"body"`
	// Name is the body's rendered name.
	Name string `json:"name"`
	Tick int64  `json:"tick"`
}

// EventKind implements [Event].
func (BodyDowned) EventKind() string { return BodyDownedKind }

// LogText implements [Loggable].
func (b BodyDowned) LogText() Text {
	return Text{ID: "text.combat.body_downed", Args: map[string]string{"name": b.Name}}
}

// PartyWiped reports that every member of the controlled party is dead or downed.
type PartyWiped struct {
	Party UID   `json:"party"`
	Tick  int64 `json:"tick"`
}

// EventKind implements [Event].
func (PartyWiped) EventKind() string { return PartyWipedKind }

// LogText implements [Loggable].
func (PartyWiped) LogText() Text { return Text{ID: "text.combat.party_wiped"} }

// Params implements [Params]: every choice of the action that dealt the damage.
func (e PartDamaged) Params() map[string]string { return e.Chosen }

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
