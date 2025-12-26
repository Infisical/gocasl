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

// Cond represents a set of conditions that must be met for a rule to apply.
// It maps field names or operators to values.
type Cond map[string]any

// Op represents a set of operator constraints on a field.
// Example: {"$gt": 10, "$lt": 20}
type Op map[string]any

// VarRef represents a reference to a template variable.
type VarRef string

// OperatorFunc defines the signature for operator implementation functions.
// It compares a field value against a constraint value.
type OperatorFunc func(fieldValue any, constraint any) bool

// Operators is a registry of available operators.
type Operators map[string]OperatorFunc
