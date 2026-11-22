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
	Roll      float64 `expr:"roll"`       // a fresh draw in [0,1) from the "effects" stream
	Distance  int     `expr:"distance"`   // hexes from source to target, or 0
	WindingUp bool    `expr:"winding_up"` // the target body has a pending action
}
