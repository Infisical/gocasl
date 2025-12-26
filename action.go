package gocasl

// ActionFor represents an action that can be performed on a specific subject type S.
// It is generic to ensure type safety when defining rules.
type ActionFor[S Subject] struct {
	name string
}

// DefineAction creates a new typed action for the subject S.
// The name parameter uniquely identifies the action (e.g., "read", "create").
func DefineAction[S Subject](name string) ActionFor[S] {
	return ActionFor[S]{
		name: name,
	}
}

// Name returns the string representation of the action.
func (a ActionFor[S]) Name() string {
	return a.name
}
