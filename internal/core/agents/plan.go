package agents

import "github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"

// Plan is the steps a planner found for a goal. It lives in comp because a body's Mind holds one.
type Plan = comp.Plan

// Step is one operator of a plan, and what its action is aimed at.
type Step = comp.PlanStep

// Target is the thing a step's action is aimed at.
type Target = comp.PlanTarget
