package gocasl

import (
	"fmt"
)

// Condition is a compiled function that evaluates a subject against a set of rules.
type Condition func(subject Subject) bool

type condCompiler struct {
	ops  Operators
	vars map[string]any
}

func newCompiler(ops Operators, vars map[string]any) *condCompiler {
	if ops == nil {
		ops = DefaultOperators
	}
	return &condCompiler{
		ops:  ops,
		vars: vars,
	}
}

func (c *condCompiler) compile(cond Cond) Condition {
	if len(cond) == 0 {
		return func(subject Subject) bool { return true }
	}

	var checks []Condition

	for key, value := range cond {
		if key == "$and" {
			checks = append(checks, c.compileAnd(value))
		} else if key == "$or" {
			checks = append(checks, c.compileOr(value))
		} else if key == "$not" {
			checks = append(checks, c.compileNot(value))
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

func (c *condCompiler) compileAnd(val any) Condition {
	list, ok := val.([]any)
	if !ok {
		// Try casting to []Cond if possible, but map[string]any is messy in Go
		// Actually the helpers create []any.
		return func(s Subject) bool { return false }
	}

	var conditions []Condition
	for _, item := range list {
		if condMap, ok := item.(Cond); ok {
			conditions = append(conditions, c.compile(condMap))
		} else if mapItem, ok := item.(map[string]any); ok {
			conditions = append(conditions, c.compile(Cond(mapItem)))
		}
	}

	return func(subject Subject) bool {
		for _, cond := range conditions {
			if !cond(subject) {
				return false
			}
		}
		return true
	}
}

func (c *condCompiler) compileOr(val any) Condition {
	list, ok := val.([]any)
	if !ok {
		return func(s Subject) bool { return false }
	}

	var conditions []Condition
	for _, item := range list {
		if condMap, ok := item.(Cond); ok {
			conditions = append(conditions, c.compile(condMap))
		} else if mapItem, ok := item.(map[string]any); ok {
			conditions = append(conditions, c.compile(Cond(mapItem)))
		}
	}

	return func(subject Subject) bool {
		if len(conditions) == 0 {
			return false // OR with empty list is false? Or true? Usually false.
		}
		for _, cond := range conditions {
			if cond(subject) {
				return true
			}
		}
		return false
	}
}

func (c *condCompiler) compileNot(val any) Condition {
	var cond Condition
	if condMap, ok := val.(Cond); ok {
		cond = c.compile(condMap)
	} else if mapItem, ok := val.(map[string]any); ok {
		cond = c.compile(Cond(mapItem))
	} else {
		return func(s Subject) bool { return false }
	}

	return func(subject Subject) bool {
		return !cond(subject)
	}
}

func (c *condCompiler) compileField(field string, value any) Condition {
	// Value can be:
	// 1. A map of operators (Op) -> {"$gt": 10}
	// 2. A bare value -> 10 (implicit $eq)
	// 3. A VarRef -> Var("ID") (implicit $eq with resolved value)
	
	resolvedValue := c.resolveValue(value)

	// Check if it's an operator map
	if opMap, ok := resolvedValue.(map[string]any); ok {
		// Check if it looks like operators (keys start with $)
		isOp := true
		for k := range opMap {
			if len(k) == 0 || k[0] != '$' {
				isOp = false
				break
			}
		}
		
		// If it's a mix or doesn't start with $, treat as direct comparison (nested object or map equality)
		// But in CASL, operators must start with $.
		// If isOp is true, compile operators.
		if isOp {
			return c.compileOperators(field, opMap)
		}
	} else if opMap, ok := resolvedValue.(Op); ok {
		return c.compileOperators(field, map[string]any(opMap))
	} else if opMap, ok := resolvedValue.(Cond); ok {
		// Could be nested condition on a field?
		// e.g. "address": {"city": "London"} -> field "address" value is object, check equality?
		// Or "address": {"$and": ...} ?
		// For now, treat as direct value unless keys are operators.
		// Use same logic as map[string]any above.
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
	// If resolvedValue contains VarRef (not resolved earlier because it wasn't top level?), handle it?
	// resolveValue handles top level VarRef.
	
	// Create $eq check
	opEqFunc := c.ops["$eq"]
	return func(subject Subject) bool {
		fieldVal := subject.GetField(field)
		return opEqFunc(fieldVal, resolvedValue)
	}
}

func (c *condCompiler) compileOperators(field string, ops map[string]any) Condition {
	var checks []func(any) bool

	for opName, opVal := range ops {
		opFunc, ok := c.ops[opName]
		if !ok {
			// Operator not found - fail safe? or panic?
			// Fail safe: condition fails.
			fmt.Printf("Warning: Operator %s not found\n", opName)
			return func(s Subject) bool { return false }
		}

		resolvedOpVal := c.resolveValue(opVal)
		
		checks = append(checks, func(fieldVal any) bool {
			return opFunc(fieldVal, resolvedOpVal)
		})
	}

	return func(subject Subject) bool {
		fieldVal := subject.GetField(field)
		for _, check := range checks {
			if !check(fieldVal) {
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
		// If var not found, what to do?
		// Return nil or keep as VarRef?
		// Returning nil might match nil checks.
		// Returning special "undefined" might be better but complexity increases.
		return nil
	}
	return val
}
