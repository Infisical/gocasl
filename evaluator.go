package gocasl

import (
	"fmt"
	"sort"
	"strings"
)

type condCompiler struct {
	fieldOps FieldOps
	condOps  CondOps
	vars     map[string]any
}

func newCompiler(fieldOps FieldOps, condOps CondOps, vars map[string]any) *condCompiler {
	if fieldOps == nil {
		fieldOps = defaultFieldOps()
	}
	if condOps == nil {
		condOps = defaultCondOps()
	}
	return &condCompiler{
		fieldOps: fieldOps,
		condOps:  condOps,
		vars:     vars,
	}
}

// ctx returns a CompileCtx that exposes the compiler's capabilities to operators.
func (c *condCompiler) ctx() *CompileCtx {
	return &CompileCtx{
		Compile: c.compile,
		Resolve: c.resolveValue,
	}
}

// validate checks all conditions for unresolved VarRefs and unknown operators.
// It returns an error describing all problems found.
func (c *condCompiler) validate(rules []rawRule) error {
	var errs []string
	for i, r := range rules {
		if r.Conditions == nil {
			continue
		}
		c.validateCond(r.Conditions, fmt.Sprintf("rule[%d](%s/%s)", i, r.SubjectType, r.Action), &errs)
	}
	if len(errs) > 0 {
		return fmt.Errorf("ability build validation failed:\n%s", joinErrors(errs))
	}
	return nil
}

func (c *condCompiler) validateCond(cond Cond, path string, errs *[]string) {
	for key, value := range cond {
		if _, ok := c.condOps[key]; ok {
			// Condition-level operator — validate its children
			c.validateCondOpValue(key, value, path, errs)
		} else {
			// Field condition — validate value for VarRefs and operator names
			c.validateValue(value, fmt.Sprintf("%s.%s", path, key), errs)
		}
	}
}

func (c *condCompiler) validateCondOpValue(opName string, value any, path string, errs *[]string) {
	condOp := c.condOps[opName]
	childPath := fmt.Sprintf("%s.%s", path, opName)
	if condOp.Validate != nil {
		condOp.Validate(c.validateCtx(childPath, errs), value)
	} else {
		c.validateValue(value, childPath, errs)
	}
}

func (c *condCompiler) validateValue(val any, path string, errs *[]string) {
	switch v := val.(type) {
	case VarRef:
		if _, found := c.vars[string(v)]; !found {
			*errs = append(*errs, fmt.Sprintf("%s: unresolved variable %q", path, string(v)))
		}
	case Op:
		c.validateOperatorMap(map[string]any(v), path, errs)
	case map[string]any:
		isOp := len(v) > 0
		for k := range v {
			if len(k) == 0 || k[0] != '$' {
				isOp = false
				break
			}
		}
		if isOp {
			c.validateOperatorMap(v, path, errs)
		}
	case Cond:
		isOp := len(v) > 0
		for k := range v {
			if len(k) == 0 || k[0] != '$' {
				isOp = false
				break
			}
		}
		if isOp {
			c.validateOperatorMap(map[string]any(v), path, errs)
		}
	}
}

func (c *condCompiler) validateOperatorMap(ops map[string]any, path string, errs *[]string) {
	// Sort keys for deterministic error messages
	keys := make([]string, 0, len(ops))
	for k := range ops {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, opName := range keys {
		fieldOp, ok := c.fieldOps[opName]
		if !ok {
			*errs = append(*errs, fmt.Sprintf("%s: unknown operator %q", path, opName))
			continue
		}
		childPath := fmt.Sprintf("%s.%s", path, opName)
		if fieldOp.Validate != nil {
			fieldOp.Validate(c.validateCtx(childPath, errs), ops[opName])
		} else {
			c.validateValue(ops[opName], childPath, errs)
		}
	}
}

func (c *condCompiler) validateCtx(path string, errs *[]string) *ValidateCtx {
	return &ValidateCtx{
		ValidateCond:  func(cond Cond, p string) { c.validateCond(cond, p, errs) },
		ValidateValue: func(val any, p string) { c.validateValue(val, p, errs) },
		Path:          path,
		Errs:          errs,
	}
}

func joinErrors(errs []string) string {
	var s strings.Builder
	for i, e := range errs {
		if i > 0 {
			s.WriteString("\n")
		}
		s.WriteString("  - " + e)
	}
	return s.String()
}

func (c *condCompiler) compile(cond Cond) Condition {
	if len(cond) == 0 {
		return func(subject Subject) bool { return true }
	}

	var checks []Condition

	for key, value := range cond {
		if condOp, ok := c.condOps[key]; ok {
			// Condition-level operator ($and, $or, $not, or custom)
			checks = append(checks, condOp.Compile(c.ctx(), value))
		} else {
			// Field condition
			checks = append(checks, c.compileField(key, value))
		}
	}

	// All checks must pass (implicit AND for top-level keys in a map)
	return func(subject Subject) bool {
		for _, check := range checks {
			if !check(subject) {
				return false
			}
		}
		return true
	}
}

func (c *condCompiler) compileField(field string, value any) Condition {
	resolvedValue := c.resolveValue(value)

	// Check if it's an operator map
	if opMap, ok := resolvedValue.(map[string]any); ok {
		isOp := true
		for k := range opMap {
			if len(k) == 0 || k[0] != '$' {
				isOp = false
				break
			}
		}
		if isOp {
			return c.compileOperators(field, opMap)
		}
	} else if opMap, ok := resolvedValue.(Op); ok {
		return c.compileOperators(field, map[string]any(opMap))
	} else if opMap, ok := resolvedValue.(Cond); ok {
		isOp := true
		for k := range opMap {
			if len(k) == 0 || k[0] != '$' {
				isOp = false
				break
			}
		}
		if isOp {
			return c.compileOperators(field, map[string]any(opMap))
		}
	}

	// Implicit $eq
	eqOp := c.fieldOps["$eq"]
	return eqOp.Compile(c.ctx(), field, resolvedValue)
}

func (c *condCompiler) compileOperators(field string, ops map[string]any) Condition {
	// Sort keys for deterministic evaluation order
	keys := make([]string, 0, len(ops))
	for k := range ops {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var checks []Condition

	for _, opName := range keys {
		opVal := ops[opName]
		fieldOp, ok := c.fieldOps[opName]
		if !ok {
			// Validated at Build() time, so this should not happen.
			return func(s Subject) bool { return false }
		}

		checks = append(checks, fieldOp.Compile(c.ctx(), field, opVal))
	}

	return func(subject Subject) bool {
		for _, check := range checks {
			if !check(subject) {
				return false
			}
		}
		return true
	}
}

func (c *condCompiler) resolveValue(val any) any {
	if v, ok := val.(VarRef); ok {
		if resolved, found := c.vars[string(v)]; found {
			return resolved
		}
		return unresolvedVarSentinel{}
	}
	return val
}

// unresolvedVarSentinel is a type that will never equal any real field value,
// ensuring unresolved variables always cause conditions to fail (deny access).
type unresolvedVarSentinel struct{}
