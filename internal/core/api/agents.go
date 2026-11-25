package api

// Event kinds of the level of detail.
const (
	PromotedKind = "agents.promoted"
	DemotedKind  = "agents.demoted"
)

// FamilyCount is how many members of a family changed form.
type FamilyCount struct {
	Family ContentID `json:"family"`
	Count  int       `json:"count"`
}

// Promoted reports that a map's populations became bodies, with the counts per family in family ID
// order.
type Promoted struct {
	Map      MapID         `json:"map"`
	Families []FamilyCount `json:"families"`
}

// EventKind implements [Event].
func (Promoted) EventKind() string { return PromotedKind }

// Demoted reports that a map's bodies were counted back into its region, with the counts per
// family in family ID order.
type Demoted struct {
	Map      MapID         `json:"map"`
	Families []FamilyCount `json:"families"`
}

// EventKind implements [Event].
func (Demoted) EventKind() string { return DemotedKind }
