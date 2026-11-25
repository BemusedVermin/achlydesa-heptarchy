package procgen

import (
	"errors"
	"fmt"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
)

// Stage changes one chunk. It reads only its Context and its own random stream, so a chunk's result depends
// on nothing else. A stage that needs a neighbor chunk's contents computes them from the seed, as noise does,
// never from a loaded chunk.
type Stage interface {
	Run(ctx *Context, chunk *hexmap.Chunk) error
}

// CheckParams reports what is wrong with a stage's params in content.
type CheckParams func(params map[string]float64) error

type registeredStage struct {
	stage Stage
	check CheckParams
}

// Stages is the registry of named stages a geography may list. Register every stage before building a
// [Source]; after that it is only read, so sources may share it across goroutines.
type Stages struct {
	byName map[string]registeredStage
}

// NewStages returns an empty registry.
func NewStages() *Stages {
	return &Stages{byName: map[string]registeredStage{}}
}

// Register adds a stage under name. A nil check accepts any params.
func (s *Stages) Register(name string, stage Stage, check CheckParams) error {
	if name == "" || stage == nil {
		return errors.New("procgen: register: a stage needs a name and an implementation")
	}
	if _, taken := s.byName[name]; taken {
		return fmt.Errorf("procgen: register: stage %q is already registered", name)
	}
	s.byName[name] = registeredStage{stage: stage, check: check}
	return nil
}

// Validate checks every geography in content: each stage is registered and takes its params.
func (s *Stages) Validate(reg *content.Registry) error {
	var problems []error
	for _, id := range reg.GeographyIDs() {
		def, _ := reg.Geography(id)
		for i, stage := range def.Stages {
			if _, err := s.lookup(stage); err != nil {
				problems = append(problems, fmt.Errorf("%s: %w", content.GeographyPath(id, i), err))
			}
		}
	}
	return errors.Join(problems...)
}

func (s *Stages) lookup(def content.StageDef) (Stage, error) {
	registered, ok := s.byName[def.Stage]
	if !ok {
		return nil, fmt.Errorf("unknown stage %q", def.Stage)
	}
	if registered.check != nil {
		if err := registered.check(def.Params); err != nil {
			return nil, fmt.Errorf("stage %q: %w", def.Stage, err)
		}
	}
	return registered.stage, nil
}
