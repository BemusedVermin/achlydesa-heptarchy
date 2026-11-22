package content

// MoveEnv is everything a move_cost formula can see.
type MoveEnv struct {
	Encumbrance float64 `expr:"encumbrance"`
}

// TurnEnv is everything a turn_ticks formula can see.
type TurnEnv struct {
	Speed int `expr:"speed"`
}
