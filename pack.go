package gocasl

import (
	"strings"
)

// PackedRule represents a compacted rule definition.
// Structure: [actions, subjects, conditions, inverted, fields, reason]
// Trailing empty values (0 or "") are removed.
type PackedRule []any

// PackRules compresses a list of JSONRules into a compact format.
func PackRules(rules []JSONRule) []PackedRule {
	packed := make([]PackedRule, len(rules))
	for i, r := range rules {
		// 1. Actions
		actions := strings.Join(r.Action, ",")

		// 2. Subjects
		subjects := strings.Join(r.Subject, ",")

		// 3. Conditions
		var conditions any = 0
		if r.Conditions != nil && len(r.Conditions) > 0 {
			conditions = r.Conditions
		}

		// 4. Inverted
		inverted := 0
		if r.Inverted {
			inverted = 1
		}

		// 5. Fields
		var fields any = 0
		if len(r.Fields) > 0 {
			fields = strings.Join(r.Fields, ",")
		}

		// 6. Reason
		reason := r.Reason

		// Construct initial array
		pRule := PackedRule{actions, subjects, conditions, inverted, fields, reason}

		// Remove trailing empty values
		// "Empty" here means 0 (int) or "" (string)
		for len(pRule) > 0 {
			last := pRule[len(pRule)-1]
			if isFalsy(last) {
				pRule = pRule[:len(pRule)-1]
			} else {
				break
			}
		}

		packed[i] = pRule
	}
	return packed
}

func isFalsy(v any) bool {
	if v == nil {
		return true
	}
	switch val := v.(type) {
	case int:
		return val == 0
	case string:
		return val == ""
	case float64:
		return val == 0
	case bool:
		return !val
	}
	return false
}

// UnpackRules decompresses a list of PackedRules into JSONRules.
func UnpackRules(packed []PackedRule) []JSONRule {
	rules := make([]JSONRule, len(packed))
	for i, p := range packed {
		r := JSONRule{}

		// Helper to safely get index
		get := func(idx int) any {
			if idx < len(p) {
				return p[idx]
			}
			return nil
		}

		// 1. Actions
		if v, ok := get(0).(string); ok && v != "" {
			r.Action = strings.Split(v, ",")
		}

		// 2. Subjects
		if v, ok := get(1).(string); ok && v != "" {
			r.Subject = strings.Split(v, ",")
		}

		// 3. Conditions
		if v := get(2); v != nil {
			// Check if it's a map. It could be 0 (number) if used as a placeholder.
			if condMap, ok := v.(map[string]any); ok {
				r.Conditions = condMap
			} else if condMap, ok := v.(Cond); ok {
				r.Conditions = condMap
			}
		}

		// 4. Inverted
		if v := get(3); v != nil {
			if val, ok := v.(int); ok && val == 1 {
				r.Inverted = true
			} else if val, ok := v.(float64); ok && val == 1.0 {
				r.Inverted = true
			}
		}

		// 5. Fields
		if v := get(4); v != nil {
			if s, ok := v.(string); ok && s != "" {
				r.Fields = strings.Split(s, ",")
			}
		}

		// 6. Reason
		if v, ok := get(5).(string); ok {
			r.Reason = v
		}

		rules[i] = r
	}
	return rules
}
