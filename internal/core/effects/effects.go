package effects

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/inventory"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

// rollStream is the RNG stream effect formulas draw their roll from.
const rollStream = "effects"

// Context is what an effect acts on. It assumes nothing about combat.
type Context struct {
	World      *sim.World
	Source     ecs.Entity         // the acting body, or zero
	SourcePart ecs.Entity         // the part acting for the body, when a part's own trigger fired; else zero
	Used       []timeline.Channel // the channels the action occupies, or none
	Target     ecs.Entity         // the target body, or zero
	TargetPart ecs.Entity         // the target part, or zero
	Dir        *api.Dir           // for direction-targeted actions
	Choices    map[string]string  // the action's chosen values, e.g. "type": "type.soak"
}

// Kind says how a primitive reads one of its parameters. The zero value is an Expr.
type Kind uint8

// The kinds of parameter.
const (
	KindExpr     Kind = iota // an Expr compiled against content.EffectEnv: Int, Float or Bool
	KindText                 // a literal string, or a formula when it starts with "=", read with String
	KindBodies               // a body selector, read with Bodies
	KindChannels             // a channel selector, read with Channels
)

// Primitive is one named effect.
type Primitive struct {
	Name string
	// Params names the required parameters. Optional names the ones that may be left out.
	Params   []string
	Optional []string
	// Kinds says how each parameter is read; a parameter not listed is an Expr.
	Kinds map[string]Kind
	// Verify, when set, checks a reference's raw parameters at content load, for rules that span
	// several parameters (such as "give exactly one of these").
	Verify func(raw map[string]string) error
	// VerifyContent, when set, checks a reference's raw parameters against content at load, for
	// literals such as item, status or faction IDs. It runs after Verify passes.
	VerifyContent func(reg *content.Registry, raw map[string]string) error
	Apply         func(ctx *Context, p Params) error
	// Check, when set, says whether the effect could apply right now, and why not.
	// Actions call it before starting and again when landing (M1-20).
	Check func(ctx *Context, p Params) (ok bool, reason api.Text)
}

// Params gives a primitive its parameters, evaluated against the context.
type Params interface {
	// Has reports whether the reference gave the parameter.
	Has(name string) bool
	Int(name string) (int, error)
	Float(name string) (float64, error)
	Bool(name string) (bool, error)
	String(name string) (string, error)               // a literal, or the result of a text formula ("= <formula>")
	Bodies(name string) ([]ecs.Entity, error)         // a body selector, sorted by UID
	Channels(name string) ([]timeline.Channel, error) // a channel selector, sorted by UID
	// For returns the same parameters evaluated for one channel: its body is the target and, when
	// the channel is a part, so is the part. A primitive that acts on a channel selector reads its
	// other parameters through it, once per selected part.
	For(c timeline.Channel) Params
}

// formulaPrefix starts a text parameter that is a formula rather than a literal.
const formulaPrefix = "="

// IsFormula reports whether a text parameter's value is a formula.
func IsFormula(value string) bool { return strings.HasPrefix(value, formulaPrefix) }

// formulaSource is the formula of a text parameter that is one.
func formulaSource(value string) string {
	return strings.TrimSpace(strings.TrimPrefix(value, formulaPrefix))
}

// programKey identifies one compiled formula.
type programKey struct {
	src string
	as  string // "int", "float" or "bool"
}

// program is one compiled formula and whether it reads roll.
type program struct {
	vm       *vm.Program
	usesRoll bool
}

// state is the registry and hooks effects keep in the world.
type state struct {
	primitives map[string]Primitive
	programs   map[programKey]*program
	// programsFor is the hash of the content the programs were compiled against: with rules
	// inlined, one source can compile to something else after the content changes.
	programsFor string
	hostile     func(w *sim.World, a, b ecs.Entity) bool
	maps        func(id api.MapID) *hexmap.Map
	triggers    *triggerIndex
}

func stateOf(w *sim.World) *state {
	res := ecs.NewResource[state](w.ECS)
	if res.Has() {
		return res.Get()
	}
	st := &state{
		primitives: map[string]Primitive{},
		programs:   map[programKey]*program{},
		hostile:    func(*sim.World, ecs.Entity, ecs.Entity) bool { return false },
	}
	res.Add(st)
	return st
}

// Install registers the timing primitives and their events.
func Install(w *sim.World) error {
	stateOf(w)
	for _, register := range []func(*api.Codec) error{
		api.RegisterEvent[api.ChannelsLocked],
		api.RegisterEvent[api.ActionsInterrupted],
		api.RegisterEvent[api.ChannelsDelayed],
		api.RegisterEvent[api.ChannelsReadied],
	} {
		if err := register(w.Codec); err != nil {
			return err
		}
	}
	if err := registerTiming(w); err != nil {
		return err
	}
	return installTriggers(w)
}

// Register makes p available to content by name. Names are unique.
func Register(w *sim.World, p Primitive) error {
	if p.Name == "" || p.Apply == nil {
		return errors.New("effects: a primitive needs a name and an Apply function")
	}
	st := stateOf(w)
	if _, dup := st.primitives[p.Name]; dup {
		return fmt.Errorf("effects: primitive %q already registered", p.Name)
	}
	seen := map[string]bool{}
	for _, name := range slices.Concat(p.Params, p.Optional) {
		if seen[name] {
			return fmt.Errorf("effects: primitive %q lists parameter %q twice", p.Name, name)
		}
		seen[name] = true
	}
	for name := range p.Kinds {
		if !seen[name] {
			return fmt.Errorf("effects: primitive %q gives a kind to unknown parameter %q", p.Name, name)
		}
	}
	st.primitives[p.Name] = p
	return nil
}

// SetHostile replaces the test for whether a is hostile to b. The default is that nobody is hostile;
// M2-07 sets it from factions.
func SetHostile(w *sim.World, fn func(w *sim.World, a, b ecs.Entity) bool) { stateOf(w).hostile = fn }

// SetMaps says where to find a map by ID, for working out what a body sees. Without it nobody sees
// anybody; M1-10 sets it from explore.
func SetMaps(w *sim.World, lookup func(id api.MapID) *hexmap.Map) { stateOf(w).maps = lookup }

// Apply runs each effect in order and stops at the first error.
func Apply(ctx *Context, refs []content.EffectRef) error {
	st := stateOf(ctx.World)
	for i, ref := range refs {
		prim, ok := st.primitives[ref.Do]
		if !ok {
			return fmt.Errorf("effects: effect %d: unknown primitive %q", i, ref.Do)
		}
		if err := prim.Apply(ctx, &params{ctx: ctx, st: st, prim: prim, raw: ref.Params}); err != nil {
			return fmt.Errorf("effects: effect %d (%s): %w", i, ref.Do, err)
		}
	}
	return nil
}

// Check asks each effect whether it could apply right now and returns the first reason it could not.
func Check(ctx *Context, refs []content.EffectRef) (bool, api.Text) {
	st := stateOf(ctx.World)
	for _, ref := range refs {
		prim, ok := st.primitives[ref.Do]
		if !ok || prim.Check == nil {
			continue
		}
		if ok, reason := prim.Check(ctx, &params{ctx: ctx, st: st, prim: prim, raw: ref.Params, checking: true}); !ok {
			return false, reason
		}
	}
	return true, api.Text{}
}

// Validate checks effect references against the registered primitives: unknown primitives, unknown
// or missing parameters, formulas that do not compile and malformed selectors. Every problem is
// reported, each naming path.
func Validate(w *sim.World, path string, refs []content.EffectRef) error {
	st := stateOf(w)
	var problems []error
	for i, ref := range refs {
		at := fmt.Sprintf("%s[%d]", path, i)
		prim, ok := st.primitives[ref.Do]
		if !ok {
			problems = append(problems, fmt.Errorf("%s: unknown primitive %q", at, ref.Do))
			continue
		}
		problems = append(problems, validateRef(at, w.Content, prim, ref)...)
	}
	return errors.Join(problems...)
}

func validateRef(at string, reg *content.Registry, prim Primitive, ref content.EffectRef) []error {
	var problems []error
	known := slices.Concat(prim.Params, prim.Optional)
	for _, name := range prim.Params {
		if _, ok := ref.Params[name]; !ok {
			problems = append(problems, fmt.Errorf("%s (%s): missing parameter %q", at, ref.Do, name))
		}
	}
	names := make([]string, 0, len(ref.Params))
	for name := range ref.Params {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if !slices.Contains(known, name) {
			problems = append(problems, fmt.Errorf("%s (%s): unknown parameter %q", at, ref.Do, name))
			continue
		}
		src := ref.Params[name]
		var err error
		switch prim.Kinds[name] {
		case KindBodies:
			_, err = ParseBodies(src)
		case KindChannels:
			_, err = ParseChannels(src)
		case KindExpr:
			err = validateFormula(reg, prim, src, nil)
		case KindText:
			if IsFormula(src) {
				err = validateFormula(reg, prim, formulaSource(src), expr.AsKind(reflect.String))
			}
		}
		if err != nil {
			problems = append(problems, fmt.Errorf("%s (%s).%s: %w", at, ref.Do, name, err))
		}
	}
	if prim.Verify != nil && len(problems) == 0 {
		if err := prim.Verify(ref.Params); err != nil {
			problems = append(problems, fmt.Errorf("%s (%s): %w", at, ref.Do, err))
		}
	}
	if prim.VerifyContent != nil && reg != nil && len(problems) == 0 {
		literals := make(map[string]string, len(ref.Params))
		for name, value := range ref.Params {
			if prim.Kinds[name] != KindText || !IsFormula(value) {
				literals[name] = value
			}
		}
		if err := prim.VerifyContent(reg, literals); err != nil {
			problems = append(problems, fmt.Errorf("%s (%s): %w", at, ref.Do, err))
		}
	}
	return problems
}

// validateFormula compiles a parameter's formula and refuses a read of roll where the primitive has a Check.
func validateFormula(reg *content.Registry, prim Primitive, src string, option expr.Option) error {
	if reg == nil {
		return nil
	}
	options := []expr.Option{}
	if option != nil {
		options = append(options, option)
	}
	if _, err := reg.Compile(src, content.EffectEnv{}, options...); err != nil {
		return err
	}
	if prim.Check == nil {
		return nil
	}
	reads, err := reg.ReadsVariable(src, "roll")
	if err != nil {
		return err
	}
	if reads {
		return errors.New("a primitive with a Check cannot read roll: checks must not draw from the RNG")
	}
	return nil
}

// ValidateContent checks every effect site in the world's content.
func ValidateContent(w *sim.World) error {
	var problems []error
	for _, site := range w.Content.EffectSites() {
		if err := Validate(w, site.Path, site.Refs); err != nil {
			problems = append(problems, err)
		}
	}
	problems = append(problems, validateTriggers(w)...)
	if err := errors.Join(problems...); err != nil {
		return fmt.Errorf("effects: content names effects that do not exist or are malformed: %w", err)
	}
	return nil
}

// params is Params for one effect reference.
type params struct {
	ctx  *Context
	st   *state
	prim Primitive
	raw  map[string]string
	// checking is set while a Check runs. A check must not change the world, so a formula that reads
	// roll is an error there rather than a draw from the stream.
	checking bool
	// base is the formula environment before roll, worked out the first time a formula runs.
	base *content.EffectEnv
}

func (p *params) Has(name string) bool {
	_, ok := p.raw[name]
	return ok
}

func (p *params) source(name string, want Kind) (string, error) {
	if got := p.prim.Kinds[name]; got != want {
		return "", fmt.Errorf("parameter %q is not read that way", name)
	}
	src, ok := p.raw[name]
	if !ok {
		return "", fmt.Errorf("missing parameter %q", name)
	}
	return src, nil
}

// eval compiles (once) and runs an Expr parameter. A formula that reads roll draws a fresh one from
// the effects stream; a formula that does not leaves the stream alone.
func (p *params) eval(name, as string, option expr.Option) (any, error) {
	src, err := p.source(name, KindExpr)
	if err != nil {
		return nil, err
	}
	return p.run(name, src, as, option)
}

// run compiles (once per content) and runs a formula for the parameter called name.
func (p *params) run(name, src, as string, option expr.Option) (any, error) {
	reg := p.ctx.World.Content
	if hash := reg.Hash(); p.st.programsFor != hash {
		p.st.programs, p.st.programsFor = map[programKey]*program{}, hash
	}
	key := programKey{src: src, as: as}
	prog, ok := p.st.programs[key]
	if !ok {
		compiled, err := reg.Compile(src, content.EffectEnv{}, option)
		if err != nil {
			return nil, fmt.Errorf("parameter %q: %w", name, err)
		}
		reads, err := reg.ReadsVariable(src, "roll")
		if err != nil {
			return nil, fmt.Errorf("parameter %q: %w", name, err)
		}
		prog = &program{vm: compiled, usesRoll: reads}
		p.st.programs[key] = prog
	}
	if prog.usesRoll && p.checking {
		return nil, fmt.Errorf("parameter %q: a check cannot read roll", name)
	}
	out, err := expr.Run(prog.vm, p.env(prog.usesRoll))
	if err != nil {
		return nil, fmt.Errorf("parameter %q: %w", name, err)
	}
	return out, nil
}

// env builds the formula environment from the context, drawing a fresh roll only when asked. The
// rest of it is worked out the first time and kept for the effect's other formulas.
func (p *params) env(withRoll bool) content.EffectEnv {
	if p.base == nil {
		base := baseEnv(p.ctx)
		p.base = &base
	}
	env := *p.base
	if withRoll {
		env.Roll = p.ctx.World.RNG.Get(rollStream).Float64()
	}
	return env
}

// baseEnv works out everything in the formula environment except roll.
func baseEnv(ctx *Context) content.EffectEnv {
	w := ctx.World
	env := content.EffectEnv{Now: w.Now, Chosen: ChosenRecords(w, ctx.Choices)}
	env.Cells, _ = strconv.Atoi(ctx.Choices["cells"]) // absent or malformed means no fall
	env.Source = SideOf(w, ctx.Source, firstPart(ctx.Used))
	env.Target = SideOf(w, ctx.Target, ctx.TargetPart)
	env.Pack = inventory.Usable(w, ctx.Source, ctx.Target)
	positions := ecs.NewMap[comp.Pos](w.ECS)
	if alive(w, ctx.Source) && alive(w, ctx.Target) && positions.Has(ctx.Source) && positions.Has(ctx.Target) {
		a, b := positions.Get(ctx.Source), positions.Get(ctx.Target)
		if a.Map == b.Map {
			env.Distance = hexmap.Distance(a.At.Hex, b.At.Hex)
			env.Height = b.At.Depth - a.At.Depth
		}
	}
	return env
}

func alive(w *sim.World, e ecs.Entity) bool { return !e.IsZero() && w.ECS.Alive(e) }

func (p *params) Int(name string) (int, error) {
	out, err := p.eval(name, "int", expr.AsInt64())
	if err != nil {
		return 0, err
	}
	return int(out.(int64)), nil
}

func (p *params) Float(name string) (float64, error) {
	out, err := p.eval(name, "float", expr.AsFloat64())
	if err != nil {
		return 0, err
	}
	return out.(float64), nil
}

func (p *params) Bool(name string) (bool, error) {
	out, err := p.eval(name, "bool", expr.AsBool())
	if err != nil {
		return false, err
	}
	return out.(bool), nil
}

func (p *params) String(name string) (string, error) {
	src, err := p.source(name, KindText)
	if err != nil {
		return "", err
	}
	if !IsFormula(src) {
		return src, nil
	}
	out, err := p.run(name, formulaSource(src), "string", expr.AsKind(reflect.String))
	if err != nil {
		return "", err
	}
	return out.(string), nil
}

// For implements [Params].
func (p *params) For(c timeline.Channel) Params {
	ctx := *p.ctx
	if !c.Body.IsZero() && c.Entity != c.Body {
		ctx.Target, ctx.TargetPart = c.Body, c.Entity
	} else {
		ctx.Target, ctx.TargetPart = c.Entity, ecs.Entity{}
	}
	return &params{ctx: &ctx, st: p.st, prim: p.prim, raw: p.raw, checking: p.checking}
}

func (p *params) Bodies(name string) ([]ecs.Entity, error) {
	src, err := p.source(name, KindBodies)
	if err != nil {
		return nil, err
	}
	sel, err := ParseBodies(src)
	if err != nil {
		return nil, fmt.Errorf("parameter %q: %w", name, err)
	}
	return sel.Resolve(p.ctx)
}

func (p *params) Channels(name string) ([]timeline.Channel, error) {
	src, err := p.source(name, KindChannels)
	if err != nil {
		return nil, err
	}
	sel, err := ParseChannels(src)
	if err != nil {
		return nil, fmt.Errorf("parameter %q: %w", name, err)
	}
	return sel.Resolve(p.ctx)
}

// Hostile reports whether a is hostile to b, as the world's hostility hook says.
func Hostile(w *sim.World, a, b ecs.Entity) bool { return stateOf(w).hostile(w, a, b) }
