package content

// MoveEnv is everything a move_cost formula can see.
type MoveEnv struct {
	Encumbrance float64 `expr:"encumbrance"`
}

// SpeedEnv is everything a step_ticks formula can see.
type SpeedEnv struct {
	Speed int `expr:"speed"`
}

// EffectEnv is everything an effect formula can see. Later tasks add fields deliberately (M2-04
// adds weak, immune, integrity, height, statuses...).
type EffectEnv struct {
	Now       int64   `expr:"now"`
	Roll      float64 `expr:"roll"`       // a fresh draw in [0,1) from the "effects" stream; refused in checks
	Distance  int     `expr:"distance"`   // hexes from source to target, or 0
	WindingUp bool    `expr:"winding_up"` // the target body has a pending action
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
