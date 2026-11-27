package api

import "strings"

// Kinds of the events the fact store publishes.
const (
	FactLearnedKind = "knowledge.fact_learned"
	FactNoticedKind = "knowledge.fact_noticed"
)

// Fact is one thing a knower holds to be true about a subject.
type Fact struct {
	// Who or what the fact is about; 0 for a fact about a cell
	Subject UID `json:"subject"`
	// Cell is the cell a fact about a place is about, such as how decayed its ground is. Facts about
	// different cells are different facts.
	Cell  *Cell     `json:"cell,omitempty"`
	Claim ContentID `json:"claim"`
	// E.g. "material.clay"
	Value  string    `json:"value"`
	Origin ContentID `json:"origin"`
	// 0..1
	Confidence float64 `json:"confidence"`
	Tick       int64   `json:"tick"`
}

// FactLearned reports that a knower learned a fact it did not hold, or a new value or confidence for
// one it held from the same source.
type FactLearned struct {
	Knower UID  `json:"knower"`
	Fact   Fact `json:"fact"`
}

// EventKind implements [Event].
func (FactLearned) EventKind() string { return FactLearnedKind }

// FactNoticed is published when the controlled party or one of its members learns a fact, even one
// it already held, so that only what the player's side learns reaches the message log.
type FactNoticed struct {
	Fact Fact `json:"fact"`
	// Shown is the fact's value as player-facing text ("clay" for "material.clay"). Empty means
	// the value is shown as it is.
	Shown string `json:"shown,omitempty"`
}

// EventKind implements [Event].
func (FactNoticed) EventKind() string { return FactNoticedKind }

// LogText implements [Loggable]; see [KnowledgeTextID].
func (n FactNoticed) LogText() Text {
	value := n.Shown
	if value == "" {
		value = n.Fact.Value
	}
	return Text{ID: KnowledgeTextID(n.Fact.Claim), Args: map[string]string{"value": value}}
}

// KnowledgeTextID is the text ID that reports a fact of the given claim: "text.knowledge." plus the
// claim without its "claim." prefix.
func KnowledgeTextID(claim ContentID) ContentID {
	return ContentID("text.knowledge." + strings.TrimPrefix(string(claim), "claim."))
}
