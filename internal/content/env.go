package content

// MoveEnv is everything a move_cost formula can see.
type MoveEnv struct {
	Encumbrance float64 `expr:"encumbrance"`
}

// SpeedEnv is everything a step_ticks formula can see.
type SpeedEnv struct {
	Speed int `expr:"speed"`
}

// GrantEnv is what the speed formula sees: the totals of a body's functioning parts' grants, by kind.
type GrantEnv struct {
	Grant map[string]int `expr:"grant"`
}

// Record is what a formula sees of one content record. Tags, props and lists it doesn't have read
// as empty: no tags, 0, an empty list.
type Record struct {
	ID    string              `expr:"id"`
	Tags  []string            `expr:"tags"`
	Props map[string]float64  `expr:"props"`
	Lists map[string][]string `expr:"lists"`
}

// BodyEnv is what a formula sees of one body: its template, one of its parts, and that part's material.
// A record is empty (ID "") when there's none.
type BodyEnv struct {
	Template Record `expr:"template"`
	Part     Record `expr:"part"`     // the part's template
	Material Record `expr:"material"` // the part's own material, which a body plan may override
	Grafts   int    `expr:"grafts"`   // how many of the body's parts are grafts

	// The rest are facts about the body or its part. They are false, 0 or empty when there is none.
	Integrity    int      `expr:"integrity"`     // the part's integrity
	MaxIntegrity int      `expr:"max_integrity"` // the part's integrity when whole
	Locked       bool     `expr:"locked"`        // the part is busy, as a lock leaves it
	Graft        bool     `expr:"graft"`         // the part is a graft
	Statuses     []string `expr:"statuses"`      // the body's status IDs
	WindingUp    bool     `expr:"winding_up"`    // the body has a pending action
	IsParty      bool     `expr:"is_party"`      // the body belongs to the party the player commands (set only when scoring)
}

// EffectEnv is everything an effect formula can see. Later tasks add fields deliberately.
type EffectEnv struct {
	Now      int64   `expr:"now"`
	Roll     float64 `expr:"roll"`     // a fresh draw in [0,1) from the "effects" stream; refused in checks
	Distance int     `expr:"distance"` // hexes from source to target, or 0

	// Source is the acting body; its part is the first channel the action occupies that is a part.
	Source BodyEnv `expr:"source"`
	// Target is the target body and the target part.
	Target BodyEnv `expr:"target"`
	// Chosen is every choice the action or event carries, as the record its value names. A value
	// that names no record, such as a direction, is Record{ID: value}.
	Chosen map[string]Record `expr:"chosen"`
	// Pack counts, per item ID, the stacks the source's pack can use on the target.
	Pack map[string]int `expr:"pack"`

	Height int `expr:"height"` // target depth minus source depth: positive when the source is above
	Cells  int `expr:"cells"`  // the cells a fall dropped, or 0
}

// ActionEnv is what an action's formulas see. Grow it deliberately.
type ActionEnv struct {
	Speed     int               `expr:"speed"`
	StepTicks int64             `expr:"step_ticks"` // the base ticks a step takes at this speed (timeline.step_ticks)
	Distance  int               `expr:"distance"`   // hexes from the actor to its target, or 0
	Used      int               `expr:"used"`       // how many channels it occupies
	Source    BodyEnv           `expr:"source"`     // the actor and the first part it uses
	Target    BodyEnv           `expr:"target"`     // the target body and part
	Chosen    map[string]Record `expr:"chosen"`     // the choices made, as the records they name
	Pack      map[string]int    `expr:"pack"`       // per item, the stacks the actor's pack can use on the target
}

// OfferEnv is what a choice's offer formula sees.
type OfferEnv struct {
	Record Record         `expr:"record"` // the candidate
	Source BodyEnv        `expr:"source"` // the actor; only its template is filled in
	Pack   map[string]int `expr:"pack"`   // the actor's whole pack: there's no target yet, so bound stacks count
}

// ScoreEnv is what combat.ai_score sees of one intent a creature could start.
type ScoreEnv struct {
	Source BodyEnv           `expr:"source"` // the actor and the first part it uses
	Target BodyEnv           `expr:"target"` // the body aimed at and the part
	Chosen map[string]Record `expr:"chosen"` // the choices made, as the records they name
	Pack   map[string]int    `expr:"pack"`   // per item, the stacks the actor's pack can use on the target
	// MaterialKnown is whether the source knows what the target part is made of.
	MaterialKnown bool `expr:"material_known"`
	// Windup is the ticks before the action would land.
	Windup int64 `expr:"windup"`
	// Height is the target's depth minus the actor's: positive when the actor is above.
	Height int `expr:"height"`
}
