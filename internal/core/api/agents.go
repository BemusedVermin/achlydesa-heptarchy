package api

import "strconv"

// Event kinds of the level of detail.
const (
	PromotedKind = "agents.promoted"
	DemotedKind  = "agents.demoted"

	MigratedKind       = "agents.migrated"
	TownEmptiedKind    = "agents.town_emptied"
	PopulationDiedKind = "agents.population_died"

	HireKind  = "agents.hire"
	HiredKind = "agents.hired"

	GrumbledKind = "agents.grumbled"
	DesertedKind = "agents.deserted"
	FledKind     = "agents.fled"
)

// Grumbled reports that an unpaid recruit complained.
type Grumbled struct {
	Party   UID `json:"party"`
	Recruit UID `json:"recruit"`
}

// EventKind implements [Event].
func (Grumbled) EventKind() string { return GrumbledKind }

// LogText implements [Loggable].
func (Grumbled) LogText() Text { return Text{ID: "text.agents.grumbled"} }

// Deserted reports that a recruit left the party and lives on as an individual.
type Deserted struct {
	Party   UID `json:"party"`
	Recruit UID `json:"recruit"`
}

// EventKind implements [Event].
func (Deserted) EventKind() string { return DesertedKind }

// LogText implements [Loggable].
func (Deserted) LogText() Text { return Text{ID: "text.agents.deserted"} }

// Fled reports that a recruit's terror took it out of the party's hands.
type Fled struct {
	Recruit UID `json:"recruit"`
}

// EventKind implements [Event].
func (Fled) EventKind() string { return FledKind }

// LogText implements [Loggable].
func (Fled) LogText() Text { return Text{ID: "text.agents.fled"} }

// OptionHirePrefix is followed by the UID of the recruit to hire.
const OptionHirePrefix = "hire/"

// HireOption is the ID of the option that hires the recruit with the UID.
func HireOption(recruit UID) string {
	return OptionHirePrefix + strconv.FormatUint(uint64(recruit), 10)
}

// ItemCount is some of one kind of item.
type ItemCount struct {
	Item  ContentID `json:"item"`
	Count int       `json:"count"`
}

// Hire pays a hirable person's price from the party's inventory and adds them to the party. It
// is allowed while a member of the party is deciding and the recruit stands next to its leader.
type Hire struct {
	Party   UID `json:"party"`
	Recruit UID `json:"recruit"`
}

// CommandKind implements [Command].
func (Hire) CommandKind() string { return HireKind }

// Hired reports that a recruit joined the party, and what it was paid.
type Hired struct {
	Party   UID         `json:"party"`
	Recruit UID         `json:"recruit"`
	Paid    []ItemCount `json:"paid"`
}

// EventKind implements [Event].
func (Hired) EventKind() string { return HiredKind }

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

// LogText implements [Loggable].
func (Hired) LogText() Text { return Text{ID: "text.agents.hired"} }
