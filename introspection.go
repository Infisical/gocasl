package gocasl

import (
	"slices"
)

// WhyNot returns the reason why an action is denied.
// It returns an empty string if the action is allowed.
// The evaluation logic mirrors Can() exactly.
func WhyNot[S Subject](a *Ability, action ActionFor[S], subject S) string {
	rules := a.index.get(subject.SubjectType(), action.Name())
	if rules == nil {
		return "No rules found for this subject/action"
	}

	granted := false

	for _, rule := range rules {
		if !rule.match(subject) {
			continue
		}

		if rule.rule.Inverted {
			// Forbid rule — mirrors Can() exactly
			if len(rule.rule.Fields) == 0 {
				if rule.rule.Reason != "" {
					return rule.rule.Reason
				}
				return "Forbidden by rule"
			}
		} else {
			granted = true
		}
	}

	if granted {
		return ""
	}

	return "Not allowed by any rule"
}

// RuleInfo describes a rule and whether it matched the subject.
type RuleInfo struct {
	Action      string
	SubjectType string
	Inverted    bool
	Conditions  Cond
	Fields      []string
	Reason      string
	Matched     bool
}

// RulesFor returns all rules that apply to the subject/action, indicating which ones matched.
// This function evaluates conditions against the provided subject instance.
func RulesFor[S Subject](a *Ability, action ActionFor[S], subject S) []RuleInfo {
	rules := a.index.get(subject.SubjectType(), action.Name())
	var infos []RuleInfo

	for _, rule := range rules {
		matched := rule.match(subject)
		infos = append(infos, RuleInfo{
			Action:      rule.rule.Action,
			SubjectType: rule.rule.SubjectType,
			Inverted:    rule.rule.Inverted,
			Conditions:  rule.rule.Conditions,
			Fields:      rule.rule.Fields,
			Reason:      rule.rule.Reason,
			Matched:     matched,
		})
	}

	return infos
}

// PossibleRulesFor returns all registered rules for the provided action and subject type.
// It ignores field restrictions and does not evaluate conditions (Matched will always be false).
// This is useful for debugging to see which rules *might* apply to this subject type.
func PossibleRulesFor[S Subject](a *Ability, action ActionFor[S]) []RuleInfo {
	var s S // Get zero value of S to access SubjectType()
	rules := a.index.get(s.SubjectType(), action.Name())
	var infos []RuleInfo

	for _, rule := range rules {
		infos = append(infos, RuleInfo{
			Action:      rule.rule.Action,
			SubjectType: rule.rule.SubjectType,
			Inverted:    rule.rule.Inverted,
			Conditions:  rule.rule.Conditions,
			Fields:      rule.rule.Fields,
			Reason:      rule.rule.Reason,
			Matched:     false, // Conditions are not evaluated against an instance
		})
	}
	return infos
}

// AllowedFields returns a list of fields that are explicitly allowed.
// It returns nil if all fields are allowed (wildcard access).
// It subtracts forbidden fields from the allowed list.
func AllowedFields[S Subject](a *Ability, action ActionFor[S], subject S) []string {
	rules := a.index.get(subject.SubjectType(), action.Name())
	if rules == nil {
		return []string{}
	}

	var allowed []string
	wildcard := false
	var forbidden []string

	for _, rule := range rules {
		if !rule.match(subject) {
			continue
		}

		if rule.rule.Inverted {
			if len(rule.rule.Fields) == 0 {
				// Resource forbidden -> no fields allowed
				return []string{}
			}
			forbidden = append(forbidden, rule.rule.Fields...)
		} else {
			if len(rule.rule.Fields) == 0 {
				wildcard = true
			} else {
				allowed = append(allowed, rule.rule.Fields...)
			}
		}
	}

	if wildcard {
		// If we allow all, we can't list them without schema.
		// Even if some are forbidden, we return nil (meaning "All except forbidden").
		return nil
	}

	// Filter allowed fields
	var result []string
	seen := make(map[string]bool)
	
	for _, f := range allowed {
		if !slices.Contains(forbidden, f) && !seen[f] {
			result = append(result, f)
			seen[f] = true
		}
	}

	return result
}

// ForbiddenFields returns a list of fields that are explicitly forbidden.
// It returns nil if all fields are forbidden.
func ForbiddenFields[S Subject](a *Ability, action ActionFor[S], subject S) []string {
	rules := a.index.get(subject.SubjectType(), action.Name())
	if rules == nil {
		// If no rules, effectively everything is forbidden (default deny).
		// But explicit forbidden fields? None.
		return []string{}
	}

	var forbidden []string
	wildcard := false

	for _, rule := range rules {
		if !rule.match(subject) {
			continue
		}

		if rule.rule.Inverted {
			if len(rule.rule.Fields) == 0 {
				wildcard = true
			} else {
				forbidden = append(forbidden, rule.rule.Fields...)
			}
		}
	}

	if wildcard {
		return nil // All forbidden
	}

	// Unique
	var result []string
	seen := make(map[string]bool)
	for _, f := range forbidden {
		if !seen[f] {
			result = append(result, f)
			seen[f] = true
		}
	}
	return result
}