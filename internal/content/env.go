package content

// MoveEnv is everything a move_cost formula can see.
type MoveEnv struct {
	Encumbrance float64 `expr:"encumbrance"`
}

// SpeedEnv is everything a step_ticks formula can see.
type SpeedEnv struct {
	Speed int `expr:"speed"`
}

// EffectEnv is everything an effect formula can see. Later tasks add fields deliberately.
type EffectEnv struct {
	Now       int64   `expr:"now"`
	Roll      float64 `expr:"roll"`       // a fresh draw in [0,1) from the "effects" stream; refused in checks
	Distance  int     `expr:"distance"`   // hexes from source to target, or 0
	WindingUp bool    `expr:"winding_up"` // the target body has a pending action

	// The rest describe the target part and body. They are false, 0 or empty when there is none.
	Weak         bool     `expr:"weak"`          // the part's material is weak to the action's chosen type
	Immune       bool     `expr:"immune"`        // the part's material is immune to the action's chosen type
	Integrity    int      `expr:"integrity"`     // the part's integrity
	MaxIntegrity int      `expr:"max_integrity"` // the part's integrity when whole
	Height       int      `expr:"height"`        // target depth minus source depth: positive when the source is above
	Statuses     []string `expr:"statuses"`      // the target body's status IDs
	Locked       bool     `expr:"locked"`        // the part is busy, as a lock leaves it
	Graft        bool     `expr:"graft"`         // the part is a graft
}

// ActionEnv is what an action's formulas see. Grow it deliberately.
type ActionEnv struct {
	Speed     int               `expr:"speed"`
	StepTicks int64             `expr:"step_ticks"` // the base ticks a step takes at this speed (timeline.step_ticks)
	Distance  int               `expr:"distance"`   // hexes from the actor to its target, or 0
	Used      int               `expr:"used"`       // how many channels it occupies
	Choice    map[string]string `expr:"choice"`     // the choices made, by provider name
}

// TypeEnv is what a damage type's condition can see.
type TypeEnv struct {
	TargetUID   uint64   `expr:"target_uid"`
	ItemBoundTo uint64   `expr:"item_bound_to"` // 0 when the item isn't bound
	TargetTags  []string `expr:"target_tags"`
}
