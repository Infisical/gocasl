package gocasl

// AbilityBuilder builds an Ability instance.
type AbilityBuilder struct {
	ops   Operators
	vars  map[string]any
	rules []rawRule
}

// rawRule is a type-erased representation of a Rule.
type rawRule struct {
	Action      string
	SubjectType string
	Inverted    bool
	Conditions  Cond
	Fields      []string
	Reason      string
}

// NewAbility creates a new AbilityBuilder with default operators.
func NewAbility() *AbilityBuilder {
	return &AbilityBuilder{
		ops: defaultOperators(),
	}
}

// WithOperators sets the operators to be used.
func (b *AbilityBuilder) WithOperators(ops Operators) *AbilityBuilder {
	b.ops = ops
	return b
}

// WithVars sets the template variables.
func (b *AbilityBuilder) WithVars(vars map[string]any) *AbilityBuilder {
	b.vars = vars
	return b
}

// AddRule adds a typed rule to the builder.
func AddRule[S Subject](b *AbilityBuilder, rule Rule[S]) *AbilityBuilder {
	var s S // zero value to get type
	b.rules = append(b.rules, rawRule{
		Action:      rule.Action.Name(),
		SubjectType: s.SubjectType(),
		Inverted:    rule.Inverted,
		Conditions:  rule.Conditions,
		Fields:      rule.Fields,
		Reason:      rule.Reason,
	})
	return b
}

// AddRules adds multiple typed rules.
func AddRules[S Subject](b *AbilityBuilder, rules ...Rule[S]) *AbilityBuilder {
	for _, r := range rules {
		AddRule(b, r)
	}
	return b
}

// Build creates the immutable Ability instance.
// It validates all rules, returning an error if any conditions reference
// unresolved template variables or unknown operators.
func (b *AbilityBuilder) Build() (*Ability, error) {
	compiler := newCompiler(b.ops, b.vars)

	if err := compiler.validate(b.rules); err != nil {
		return nil, err
	}

	index := newRuleIndex(b.rules, compiler)

	return &Ability{
		index:    index,
		compiler: compiler,
	}, nil
}

// Ability represents a set of permissions.
type Ability struct {
	index    *ruleIndex
	compiler *condCompiler
}
