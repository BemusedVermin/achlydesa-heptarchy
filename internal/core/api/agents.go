package api

// Event kinds of the level of detail.
const (
	PromotedKind = "agents.promoted"
	DemotedKind  = "agents.demoted"

	MigratedKind       = "agents.migrated"
	TownEmptiedKind    = "agents.town_emptied"
	PopulationDiedKind = "agents.population_died"
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

// Migrated reports that residents left a map for a neighbouring one: off screen because a stock
// ran short, or a body crossing a gate to a map the party is not on.
type Migrated struct {
	From   MapID     `json:"from"`
	To     MapID     `json:"to"`
	Family ContentID `json:"family"`
	Count  int       `json:"count"`
	// Short is the item that ran short; empty for a body that crossed a gate.
	Short ContentID `json:"short"`
}

// EventKind implements [Event].
func (Migrated) EventKind() string { return MigratedKind }

// TownEmptied reports that a town's last residents are gone.
type TownEmptied struct {
	Map    MapID     `json:"map"`
	Family ContentID `json:"family"`
}

// EventKind implements [Event].
func (TownEmptied) EventKind() string { return TownEmptiedKind }

// PopulationDied reports that members of a family died off screen.
type PopulationDied struct {
	Map    MapID     `json:"map"`
	Family ContentID `json:"family"`
	Count  int       `json:"count"`
}

// EventKind implements [Event].
func (PopulationDied) EventKind() string { return PopulationDiedKind }
