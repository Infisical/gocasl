package gocasl

// Subject is the interface that all subjects (resources/entities) must implement
// to be compatible with the authorization system.
type Subject interface {
	// SubjectType returns a unique string identifier for the subject type.
	// This is used for rule indexing and matching.
	SubjectType() string

	// GetField returns the value of the specified field.
	// It should return nil if the field does not exist.
	GetField(field string) any
}

// MapSubject wraps a map[string]any to implement the Subject interface.
// This is useful for evaluating conditions against array elements in operators
// like $elemMatch and $all.
type MapSubject map[string]any

func (m MapSubject) SubjectType() string   { return "" }
func (m MapSubject) GetField(f string) any { return m[f] }

// Cond represents a set of conditions that must be met for a rule to apply.
// It maps field names or operators to values.
type Cond map[string]any

// Op represents a set of operator constraints on a field.
// Example: {"$gt": 10, "$lt": 20}
type Op map[string]any

// VarRef represents a reference to a template variable.
type VarRef string

// CompileCtx is the context passed to FieldOp and CondOp functions during compilation.
// It provides access to the compiler so operators can recursively compile sub-conditions.
type CompileCtx struct {
	// Compile recursively compiles a Cond into a Condition function.
	Compile func(Cond) Condition
	// Resolve resolves VarRef values to their actual values.
	Resolve func(any) any
}

// FieldOp compiles a field-level operator into a Condition.
// It is called at build time when the compiler encounters {"field": {"$op": constraint}}.
// The returned Condition is evaluated at runtime — no recompilation needed.
type FieldOp func(cc *CompileCtx, field string, constraint any) Condition

// CondOp compiles a condition-level operator into a Condition.
// It is called at build time when the compiler encounters {"$op": value} at the top level of a Cond.
// Used for logical operators like $and, $or, $not.
type CondOp func(cc *CompileCtx, value any) Condition

// FieldOps is a registry of field-level operators.
type FieldOps map[string]FieldOp

// CondOps is a registry of condition-level operators.
type CondOps map[string]CondOp

// Compare wraps a simple comparison function into a FieldOp.
// Use this for operators that just compare a field value against a constraint.
func Compare(fn func(fieldValue, constraint any) bool) FieldOp {
	return func(cc *CompileCtx, field string, constraint any) Condition {
		resolved := cc.Resolve(constraint)
		return func(s Subject) bool {
			return fn(s.GetField(field), resolved)
		}
	}
}

// --- FieldOps methods ---

// Clone returns a shallow copy of the field ops map.
func (o FieldOps) Clone() FieldOps {
	newOps := make(FieldOps, len(o))
	for k, v := range o {
		newOps[k] = v
	}
	return newOps
}

// With adds or replaces a field operator.
func (o FieldOps) With(name string, fn FieldOp) FieldOps {
	newOps := o.Clone()
	newOps[name] = fn
	return newOps
}

// Without removes field operators by name.
func (o FieldOps) Without(names ...string) FieldOps {
	newOps := o.Clone()
	for _, name := range names {
		delete(newOps, name)
	}
	return newOps
}

// WithAll merges another set of field operators into this one.
func (o FieldOps) WithAll(other FieldOps) FieldOps {
	newOps := o.Clone()
	for k, v := range other {
		newOps[k] = v
	}
	return newOps
}

// --- CondOps methods ---

// Clone returns a shallow copy of the cond ops map.
func (o CondOps) Clone() CondOps {
	newOps := make(CondOps, len(o))
	for k, v := range o {
		newOps[k] = v
	}
	return newOps
}

// With adds or replaces a condition operator.
func (o CondOps) With(name string, fn CondOp) CondOps {
	newOps := o.Clone()
	newOps[name] = fn
	return newOps
}

// Without removes condition operators by name.
func (o CondOps) Without(names ...string) CondOps {
	newOps := o.Clone()
	for _, name := range names {
		delete(newOps, name)
	}
	return newOps
}

// WithAll merges another set of condition operators into this one.
func (o CondOps) WithAll(other CondOps) CondOps {
	newOps := o.Clone()
	for k, v := range other {
		newOps[k] = v
	}
	return newOps
}

// Condition is a compiled function that evaluates a subject against a set of rules.
type Condition func(subject Subject) bool
