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

// ValidateCtx is the context passed to operator Validate functions during build-time validation.
// It provides access to recursive validation so operators can validate nested conditions.
type ValidateCtx struct {
	// ValidateCond recursively validates a nested Cond.
	ValidateCond func(cond Cond, path string)
	// ValidateValue validates a scalar value (checks for VarRefs and operator maps).
	ValidateValue func(val any, path string)
	// Path is the current validation path for error messages.
	Path string
	// Errs collects validation errors.
	Errs *[]string
}

// FieldOp defines a field-level operator.
// Compile is called at build time when the compiler encounters {"field": {"$op": constraint}}.
// The returned Condition is evaluated at runtime — no recompilation needed.
// Validate is optional — if set, it is called during build-time validation to recursively
// validate the operator's constraint. If nil, the constraint is validated as a scalar value.
type FieldOp struct {
	Compile  func(cc *CompileCtx, field string, constraint any) Condition
	Validate func(vc *ValidateCtx, constraint any)
}

// CondOp defines a condition-level operator.
// Compile is called at build time when the compiler encounters {"$op": value} at the top level of a Cond.
// Used for logical operators like $and, $or, $not.
// Validate is optional — if set, it is called during build-time validation to recursively
// validate the operator's value. If nil, the value is validated as a scalar.
type CondOp struct {
	Compile  func(cc *CompileCtx, value any) Condition
	Validate func(vc *ValidateCtx, value any)
}

// FieldOps is a registry of field-level operators.
type FieldOps map[string]FieldOp

// CondOps is a registry of condition-level operators.
type CondOps map[string]CondOp

// Compare wraps a simple comparison function into a FieldOp.
// Use this for operators that just compare a field value against a constraint.
// The returned FieldOp has no Validate function (nil), meaning the constraint
// is validated as a scalar value by default.
func Compare(fn func(fieldValue, constraint any) bool) FieldOp {
	return FieldOp{
		Compile: func(cc *CompileCtx, field string, constraint any) Condition {
			resolved := cc.Resolve(constraint)
			return func(s Subject) bool {
				return fn(s.GetField(field), resolved)
			}
		},
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

// ValidateCondConstraint is a Validate helper for operators whose constraint is a single
// nested Cond (e.g., $elemMatch, $all, $not). It recursively validates the nested condition.
func ValidateCondConstraint(vc *ValidateCtx, constraint any) {
	switch v := constraint.(type) {
	case Cond:
		vc.ValidateCond(v, vc.Path)
	case map[string]any:
		vc.ValidateCond(Cond(v), vc.Path)
	}
}

// ValidateCondSliceConstraint is a Validate helper for operators whose constraint is a
// slice of Cond values (e.g., $and, $or). It recursively validates each nested condition.
func ValidateCondSliceConstraint(vc *ValidateCtx, constraint any) {
	list, ok := constraint.([]any)
	if !ok {
		return
	}
	for i, item := range list {
		itemPath := vc.Path + "[" + itoa(i) + "]"
		switch v := item.(type) {
		case Cond:
			vc.ValidateCond(v, itemPath)
		case map[string]any:
			vc.ValidateCond(Cond(v), itemPath)
		}
	}
}

// itoa is a simple int-to-string helper to avoid importing strconv.
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var buf [20]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	return string(buf[pos:])
}

// Condition is a compiled function that evaluates a subject against a set of rules.
type Condition func(subject Subject) bool
