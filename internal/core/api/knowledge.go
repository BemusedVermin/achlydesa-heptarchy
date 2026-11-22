package api

import "strings"

// Kinds of the events the fact store publishes.
const (
	FactLearnedKind = "knowledge.fact_learned"
	FactNoticedKind = "knowledge.fact_noticed"
)

// Fact is one thing a knower holds to be true about a subject.
type Fact struct {
	Subject    UID       `json:"subject"`    // who or what the fact is about
	Claim      ContentID `json:"claim"`      // "claim.material", "claim.name", ...
	Value      string    `json:"value"`      // e.g. "material.clay"
	Source     ContentID `json:"source"`     // "source.sweep", "source.rumor", ...
	Confidence float64   `json:"confidence"` // 0..1
	Tick       int64     `json:"tick"`
}

// FactLearned reports that a knower learned a fact, or learned it again from the same source.
type FactLearned struct {
	Knower UID  `json:"knower"`
	Fact   Fact `json:"fact"`
}

// EventKind implements [Event].
func (FactLearned) EventKind() string { return FactLearnedKind }

// FactNoticed is published beside [FactLearned] when the knower is the controlled party or one of
// its members, so that only what the player's side learns reaches the message log.
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
// claim without its "claim." prefix, so "text.knowledge.material" for "claim.material".
func KnowledgeTextID(claim ContentID) ContentID {
	return ContentID("text.knowledge." + strings.TrimPrefix(string(claim), "claim."))
}
