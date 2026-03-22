package gocasl

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
)

// StringOrSlice is a helper type for JSON unmarshalling that handles both single string and array of strings.
type StringOrSlice []string

func (s *StringOrSlice) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*s = nil
		return nil
	}
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		*s = []string{single}
		return nil
	}
	var slice []string
	if err := json.Unmarshal(data, &slice); err == nil {
		*s = slice
		return nil
	}
	return fmt.Errorf("expected string or array of strings")
}

// JSONRule represents the JSON structure of a rule.
type JSONRule struct {
	Action     StringOrSlice `json:"action"`
	Subject    StringOrSlice `json:"subject"`
	Fields     StringOrSlice `json:"fields"`
	Conditions Cond          `json:"conditions"`
	Inverted   bool          `json:"inverted"`
	Reason     string        `json:"reason"`
}

// DefaultMaxJSONSize is the default maximum size in bytes for JSON rule loading (1MB).
const DefaultMaxJSONSize = 1 << 20

// LoadOptions configures the JSON loading process.
type LoadOptions struct {
	// Operators to use in the Ability. Defaults to DefaultOperators.
	Operators Operators
	// Variables for template resolution.
	Vars map[string]any
	// MaxSize limits the maximum bytes read from a Reader.
	// Defaults to DefaultMaxJSONSize (1MB) if zero.
	MaxSize int64
}

// LoadFromJSON loads rules from a JSON byte slice.
func LoadFromJSON(data []byte, opts LoadOptions) (*Ability, error) {
	var jsonRules []JSONRule
	if err := json.Unmarshal(data, &jsonRules); err != nil {
		return nil, err
	}

	builder := NewAbility()
	if opts.Operators != nil {
		builder.WithOperators(opts.Operators)
	}
	if opts.Vars != nil {
		builder.WithVars(opts.Vars)
	}

	for _, jr := range jsonRules {
		// Process conditions to resolve template variables
		cond := processConditions(jr.Conditions)

		// Expand actions and subjects
		for _, subject := range jr.Subject {
			for _, action := range jr.Action {
				builder.addRawRule(rawRule{
					Action:      action,
					SubjectType: subject,
					Inverted:    jr.Inverted,
					Conditions:  cond,
					Fields:      jr.Fields,
					Reason:      jr.Reason,
				})
			}
		}
	}

	return builder.Build()
}

// LoadFromFile loads rules from a JSON file.
func LoadFromFile(path string, opts LoadOptions) (*Ability, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return LoadFromJSON(data, opts)
}

// LoadFromReader loads rules from an io.Reader.
// It limits the bytes read to opts.MaxSize (defaults to DefaultMaxJSONSize).
func LoadFromReader(r io.Reader, opts LoadOptions) (*Ability, error) {
	maxSize := opts.MaxSize
	if maxSize <= 0 {
		maxSize = DefaultMaxJSONSize
	}
	data, err := io.ReadAll(io.LimitReader(r, maxSize+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxSize {
		return nil, fmt.Errorf("JSON input exceeds maximum size of %d bytes", maxSize)
	}
	return LoadFromJSON(data, opts)
}

// Helper to add raw rule to builder (since rules field is private)
// We need to add this method to AbilityBuilder in ability.go or use reflection/unsafe?
// Since we are in the same package, we can access private fields/methods?
// Yes, LoadFromJSON is in package gocasl, AbilityBuilder is in package gocasl.
// We can modify ability.go to add `addRawRule` or just access `rules` directly.
// `rules` is `[]rawRule`. We can append to it.

func (b *AbilityBuilder) addRawRule(r rawRule) {
	b.rules = append(b.rules, r)
}

// Template variable regex: {{ .VarName }}
var templateVarRegex = regexp.MustCompile(`^\{\{\s*\.(\w+)\s*\}\}$`)

func processConditions(cond Cond) Cond {
	if cond == nil {
		return nil
	}
	newCond := make(Cond)
	for k, v := range cond {
		newCond[k] = processValue(v)
	}
	return newCond
}

func processValue(val any) any {
	switch v := val.(type) {
	case string:
		matches := templateVarRegex.FindStringSubmatch(v)
		if len(matches) == 2 {
			return Var(matches[1])
		}
		return v
	case map[string]any:
		newMap := make(map[string]any)
		for k, val := range v {
			newMap[k] = processValue(val)
		}
		return newMap
	case []any:
		newList := make([]any, len(v))
		for i, val := range v {
			newList[i] = processValue(val)
		}
		return newList
	default:
		return v
	}
}
