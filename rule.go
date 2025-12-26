package gocasl

import "fmt"

// Rule represents a single authorization rule for a specific subject type S.
type Rule[S Subject] struct {
	Action   ActionFor[S]
	Inverted bool // Inverted is true for Forbid rules, false for Allow rules
	Conditions     Cond
	Fields   []string
	Reason   string
}

// RuleBuilder helps construct a Rule with a fluent API.
type RuleBuilder[S Subject] struct {
	rule Rule[S]
}

// Allow starts building a rule that allows the specified action.
func Allow[S Subject](action ActionFor[S]) *RuleBuilder[S] {
	return &RuleBuilder[S]{
		rule: Rule[S]{
			Action:   action,
			Inverted: false,
		},
	}
}

// Forbid starts building a rule that forbids the specified action.
func Forbid[S Subject](action ActionFor[S]) *RuleBuilder[S] {
	return &RuleBuilder[S]{
		rule: Rule[S]{
			Action:   action,
			Inverted: true,
		},
	}
}

// Where adds conditions to the rule.
func (b *RuleBuilder[S]) Where(cond Cond) *RuleBuilder[S] {
	b.rule.Conditions = cond
	return b
}

// OnFields limits the rule to specific fields.
func (b *RuleBuilder[S]) OnFields(fields ...string) *RuleBuilder[S] {
	b.rule.Fields = fields
	return b
}

// Because adds a reason for the rule (useful for Forbid rules).
func (b *RuleBuilder[S]) Because(reason string) *RuleBuilder[S] {
	b.rule.Reason = reason
	return b
}

// Build finalizes the rule construction.
func (b *RuleBuilder[S]) Build() Rule[S] {
	// Validation could happen here
	if b.rule.Action.name == "" {
		// This shouldn't happen if using DefineAction correctly, but good to know.
		// Since DefineAction is the only way to get ActionFor[S], name should be set.
	}
	// We return a copy since Rule is a struct (passed by value)
	return b.rule
}

// String provides a string representation of the rule for debugging.
func (r Rule[S]) String() string {
	typeStr := "Allow"
	if r.Inverted {
		typeStr = "Forbid"
	}
	return fmt.Sprintf("%s %s on %T Where %v Fields %v", typeStr, r.Action.Name(), *new(S), r.Conditions, r.Fields)
}
