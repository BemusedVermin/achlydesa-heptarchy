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
	// Part is the part's template.
	Part Record `expr:"part"`
	// Material is the part's own material, which a body plan may override.
	Material Record `expr:"material"`
	// Grafts is how many of the body's parts are grafts.
	Grafts int `expr:"grafts"`

	// The rest are facts about the body or its part. They are false, 0 or empty when there is none.
	Integrity int `expr:"integrity"`
	// MaxIntegrity is the part's integrity when whole.
	MaxIntegrity int `expr:"max_integrity"`
	// Locked means the part is busy, as a lock leaves it.
	Locked bool `expr:"locked"`
	// Graft means the part is a graft.
	Graft bool `expr:"graft"`
	// Statuses are the body's status IDs.
	Statuses []string `expr:"statuses"`
	// WindingUp means the body has a pending action.
	WindingUp bool `expr:"winding_up"`
	// IsParty means the body belongs to the party the player commands (set only when scoring).
	IsParty bool `expr:"is_party"`
}

// EffectEnv is everything an effect formula can see.
type EffectEnv struct {
	Now int64 `expr:"now"`
	// Roll is a fresh draw in [0,1) from the "effects" stream; refused in checks.
	Roll float64 `expr:"roll"`
	// Distance is hexes from source to target, or 0.
	Distance int `expr:"distance"`

	// Source is the acting body; its part is the first channel the action occupies that is a part.
	Source BodyEnv `expr:"source"`
	// Target is the target body and the target part.
	Target BodyEnv `expr:"target"`
	// Chosen is every choice the action or event carries, as the record its value names. A value
	// that names no record, such as a direction, is Record{ID: value}.
	Chosen map[string]Record `expr:"chosen"`
	// Inventory counts, per item ID, the stacks the source's inventory can use on the target.
	Inventory map[string]int `expr:"inventory"`

	// Height is target depth minus source depth: positive when the source is above.
	Height int `expr:"height"`
	// Cells is the cells a fall dropped, or 0.
	Cells int `expr:"cells"`
}

// ActionEnv is what an action's formulas see.
type ActionEnv struct {
	Speed int `expr:"speed"`
	// StepTicks is the base ticks a step takes at this speed (timeline.step_ticks).
	StepTicks int64 `expr:"step_ticks"`
	// Distance is hexes from the actor to its target, or 0.
	Distance int `expr:"distance"`
	// Used is how many channels it occupies.
	Used int `expr:"used"`
	// Source is the actor and the first part it uses.
	Source BodyEnv `expr:"source"`
	// Target is the target body and part.
	Target BodyEnv `expr:"target"`
	// Chosen is the choices made, as the records they name.
	Chosen map[string]Record `expr:"chosen"`
	// Inventory is per item, the stacks the actor's inventory can use on the target.
	Inventory map[string]int `expr:"inventory"`
}

// OfferEnv is what a choice's offer formula sees.
type OfferEnv struct {
	// Record is the candidate.
	Record Record `expr:"record"`
	// Source is the actor; only its template is filled in.
	Source BodyEnv `expr:"source"`
	// Inventory is the actor's whole inventory: there's no target yet, so bound stacks count.
	Inventory map[string]int `expr:"inventory"`
}

// ScoreEnv is what combat.ai_score sees of one intent a creature could start.
type ScoreEnv struct {
	// Source is the actor and the first part it uses.
	Source BodyEnv `expr:"source"`
	// Target is the body aimed at and the part.
	Target BodyEnv `expr:"target"`
	// Chosen is the choices made, as the records they name.
	Chosen map[string]Record `expr:"chosen"`
	// Inventory is per item, the stacks the actor's inventory can use on the target.
	Inventory map[string]int `expr:"inventory"`
	// MaterialKnown is whether the source knows what the target part is made of.
	MaterialKnown bool `expr:"material_known"`
	// Windup is the ticks before the action would land.
	Windup int64 `expr:"windup"`
	// Height is the target's depth minus the actor's: positive when the actor is above.
	Height int `expr:"height"`
}

// DecayEnv is everything a geography's decay level formula can see.
type DecayEnv struct {
	// Depth is the cell's depth.
	Depth int `expr:"depth"`
	// Exports is the export count.
	Exports int `expr:"exports"`
}

// RefillEnv is what a source's refill formula sees.
type RefillEnv struct {
	Stock int `expr:"stock"`
	Max   int `expr:"max"`
}
