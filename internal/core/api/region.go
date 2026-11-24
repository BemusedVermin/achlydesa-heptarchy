package api

// RegionView is the region the party is in.
type RegionView struct {
	Map  MapID  `json:"map"`
	Name string `json:"name"`
	Town bool   `json:"town"`
	// Stock is the town's stock, in item ID order; it is empty outside a town.
	Stock []StockLine `json:"stock,omitempty"`
}

// StockLine is how much of one item a stock holds.
type StockLine struct {
	Item  ContentID `json:"item"`
	Name  string    `json:"name"`
	Count int       `json:"count"`
}
