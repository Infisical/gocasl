package gocasl

import (
	"reflect"
	"slices"
)

// isNilSubject checks whether a Subject value is a nil pointer/interface.
func isNilSubject(s Subject) bool {
	if s == nil {
		return true
	}
	v := reflect.ValueOf(s)
	return v.Kind() == reflect.Pointer && v.IsNil()
}

// Can checks if the subject is allowed to perform the action.
// It checks for resource-level access, ignoring field-level restrictions
// unless a forbid rule blocks the entire resource.
// It returns false (deny) if subject is nil.
func Can[S Subject](a *Ability, action ActionFor[S], subject S) bool {
	if isNilSubject(subject) {
		return false
	}
	rules := a.index.get(subject.SubjectType(), action.Name())
	if rules == nil {
		return false
	}

	granted := false

	for _, rule := range rules {
		// Evaluate condition
		if !rule.match(subject) {
			continue
		}

		if rule.rule.Inverted {
			// Forbid rule
			// If rule applies to specific fields, it doesn't forbid the whole resource.
			// If rule has NO fields, it forbids the resource.
			if len(rule.rule.Fields) == 0 {
				return false
			}
		} else {
			// Allow rule
			// Any matching allow rule grants access (unless overridden by Forbid)
			granted = true
		}
	}

	return granted
}

// Cannot checks if the subject is forbidden from performing the action.
func Cannot[S Subject](a *Ability, action ActionFor[S], subject S) bool {
	return !Can(a, action, subject)
}

// CanWithField checks if the subject is allowed to perform the action on a specific field.
// It returns false (deny) if subject is nil.
func CanWithField[S Subject](a *Ability, action ActionFor[S], subject S, field string) bool {
	if isNilSubject(subject) {
		return false
	}
	rules := a.index.get(subject.SubjectType(), action.Name())
	if rules == nil {
		return false
	}

	granted := false

	for _, rule := range rules {
		if !rule.match(subject) {
			continue
		}

		if rule.rule.Inverted {
			// Forbid rule
			// Blocks if it applies to all fields OR explicitly includes this field
			if len(rule.rule.Fields) == 0 || slices.Contains(rule.rule.Fields, field) {
				return false
			}
		} else {
			// Allow rule
			// Grants if it applies to all fields OR explicitly includes this field
			if len(rule.rule.Fields) == 0 || slices.Contains(rule.rule.Fields, field) {
				granted = true
			}
		}
	}

	return granted
}

// CanAll checks if the subject is allowed to perform ALL of the specified actions.
func CanAll[S Subject](a *Ability, subject S, actions ...ActionFor[S]) bool {
	for _, action := range actions {
		if !Can(a, action, subject) {
			return false
		}
	}
	return true
}

// CanAny checks if the subject is allowed to perform ANY of the specified actions.
func CanAny[S Subject](a *Ability, subject S, actions ...ActionFor[S]) bool {
	for _, action := range actions {
		if Can(a, action, subject) {
			return true
		}
	}
	return false
}
