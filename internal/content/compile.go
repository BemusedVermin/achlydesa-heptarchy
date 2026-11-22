package content

import (
	"fmt"
	"slices"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/ast"
	"github.com/expr-lang/expr/parser"
	"github.com/expr-lang/expr/vm"
)

// ruleVar is the name formulas write rules under: rule.<name>.
var ruleVar = "rule"

// Compile is the one way a formula is compiled. env is a value of the struct the formula reads
// (EffectEnv{}, ActionEnv{}, ...). Every rule.<name> in src is replaced by the rule's own source
// before the formula is type-checked, so a rule takes its type from where it is used and is
// checked in each environment that uses it.
func (r *Registry) Compile(src string, env any, options ...expr.Option) (*vm.Program, error) {
	inliner := &inliner{reg: r, failure: new(error)}
	all := append([]expr.Option{expr.Env(env), expr.Patch(inliner)}, options...)
	program, err := expr.Compile(src, all...)
	if *inliner.failure != nil {
		return nil, *inliner.failure
	}
	if err != nil {
		return nil, err
	}
	return program, nil
}

// ReadsVariable reports whether the formula, with its rules inlined, names the variable.
func (r *Registry) ReadsVariable(src, name string) (bool, error) {
	tree, err := parser.Parse(src)
	if err != nil {
		return false, err
	}
	inliner := &inliner{reg: r, failure: new(error)}
	ast.Walk(&tree.Node, inliner)
	if *inliner.failure != nil {
		return false, *inliner.failure
	}
	finder := &variableFinder{name: name}
	ast.Walk(&tree.Node, finder)
	return finder.found, nil
}

type variableFinder struct {
	name  string
	found bool
}

func (f *variableFinder) Visit(node *ast.Node) {
	if id, ok := (*node).(*ast.IdentifierNode); ok && id.Value == f.name {
		f.found = true
	}
}

// inliner is the patch that replaces rule.<name> with the rule's parsed source.
type inliner struct {
	reg     *Registry
	stack   []string // the rules being expanded, to catch cycles
	failure *error   // the first problem, shared with the inliners of nested rules
}

func (in *inliner) fail(err error) {
	if *in.failure == nil {
		*in.failure = err
	}
}

// Visit implements [ast.Visitor].
func (in *inliner) Visit(node *ast.Node) {
	member, ok := (*node).(*ast.MemberNode)
	if !ok {
		return
	}
	id, ok := member.Node.(*ast.IdentifierNode)
	if !ok || id.Value != ruleVar {
		return
	}
	property, ok := member.Property.(*ast.StringNode)
	if !ok {
		in.fail(fmt.Errorf("a rule is named like rule.<name>"))
		return
	}
	body, err := in.expand(property.Value)
	if err != nil {
		in.fail(err)
		return
	}
	ast.Patch(node, body)
}

// expand parses the named rule and inlines the rules it uses.
func (in *inliner) expand(name string) (ast.Node, error) {
	src, ok := in.reg.bundle.Rules[name]
	if !ok {
		return nil, fmt.Errorf("unknown rule %q", name)
	}
	if slices.Contains(in.stack, name) {
		return nil, fmt.Errorf("rule %q uses itself through %v", name, in.stack)
	}
	if in.reg.rulesUsed != nil {
		in.reg.rulesUsed[name] = true
	}
	tree, err := parser.Parse(src)
	if err != nil {
		return nil, fmt.Errorf("rule %q: %w", name, err)
	}
	nested := &inliner{reg: in.reg, stack: append(slices.Clone(in.stack), name), failure: in.failure}
	ast.Walk(&tree.Node, nested)
	return tree.Node, nil
}

// ruleRefs lists the rules a rule's source names, in the order found.
func ruleRefs(src string) ([]string, error) {
	tree, err := parser.Parse(src)
	if err != nil {
		return nil, err
	}
	finder := &ruleFinder{}
	ast.Walk(&tree.Node, finder)
	return finder.names, nil
}

type ruleFinder struct{ names []string }

func (f *ruleFinder) Visit(node *ast.Node) {
	member, ok := (*node).(*ast.MemberNode)
	if !ok {
		return
	}
	if id, ok := member.Node.(*ast.IdentifierNode); ok && id.Value == ruleVar {
		if property, ok := member.Property.(*ast.StringNode); ok {
			f.names = append(f.names, property.Value)
		}
	}
}

// checkRules checks rule names and syntax, that every rule they use exists, and that no rules use
// each other in a cycle.
func (r *Registry) checkRules() []error {
	var problems []error
	refs := map[string][]string{}
	for _, name := range sortedKeys(r.bundle.Rules) {
		path := fmt.Sprintf("rules[%q]", name)
		if !keyPattern.MatchString(name) {
			problems = append(problems, fmt.Errorf("%s: a rule name must be lower-case words", path))
		}
		used, err := ruleRefs(r.bundle.Rules[name])
		if err != nil {
			problems = append(problems, fmt.Errorf("%s: %w", path, err))
			continue
		}
		refs[name] = used
		for _, other := range used {
			if _, ok := r.bundle.Rules[other]; !ok {
				problems = append(problems, fmt.Errorf("%s: unknown rule %q", path, other))
			}
		}
	}
	state := map[string]int{} // 1 while on the walk, 2 when done
	var walk func(name string, trail []string)
	walk = func(name string, trail []string) {
		switch state[name] {
		case 1:
			problems = append(
				problems,
				fmt.Errorf("rules[%q]: rules use each other in a cycle: %v", name, append(trail, name)),
			)
			return
		case 2:
			return
		}
		state[name] = 1
		for _, other := range refs[name] {
			if _, ok := r.bundle.Rules[other]; ok {
				walk(other, append(slices.Clone(trail), name))
			}
		}
		state[name] = 2
	}
	for _, name := range sortedKeys(r.bundle.Rules) {
		walk(name, nil)
	}
	return problems
}

// warnUnusedRules records a warning for each rule no formula used.
func (r *Registry) warnUnusedRules() {
	for _, name := range sortedKeys(r.bundle.Rules) {
		if !r.rulesUsed[name] {
			r.warnings = append(r.warnings, fmt.Sprintf("rules[%q] is not used by any formula", name))
		}
	}
}
