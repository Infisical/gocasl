package gocasl

// Var creates a reference to a template variable.
// These are resolved at evaluation time using the variables map provided to the Ability.
func Var(name string) VarRef {
	return VarRef(name)
}

// And combines multiple conditions with a logical AND.
// All conditions must evaluate to true.
func And(conds ...Cond) Cond {
	// Convert []Cond to []any because that's what the map expects
	args := make([]any, len(conds))
	for i, c := range conds {
		args[i] = c
	}
	return Cond{"$and": args}
}

// Or combines multiple conditions with a logical OR.
// At least one condition must evaluate to true.
func Or(conds ...Cond) Cond {
	args := make([]any, len(conds))
	for i, c := range conds {
		args[i] = c
	}
	return Cond{"$or": args}
}

// Not negates a condition.
// The condition must evaluate to false.
func Not(cond Cond) Cond {
	return Cond{"$not": cond}
}
