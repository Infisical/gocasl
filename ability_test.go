package gocasl

import (
	"testing"
)

func TestAbilityBuilder(t *testing.T) {
	// Test NewAbility
	b := NewAbility()
	if b.ops == nil {
		t.Errorf("NewAbility should set default operators")
	}

	// Test WithOperators
	ops := Operators{}
	b.WithOperators(ops)
	if len(b.ops) != 0 {
		t.Errorf("WithOperators failed")
	}

	// Test WithVars
	vars := map[string]any{"a": 1}
	b.WithVars(vars)
	if b.vars["a"] != 1 {
		t.Errorf("WithVars failed")
	}

	// Test AddRule
	readPost := DefineAction[mockSubject]("read")
	r := Allow(readPost).Build()
	
	AddRule(b, r)
	
	if len(b.rules) != 1 {
		t.Errorf("AddRule failed to add rule")
	}
	if b.rules[0].SubjectType != "MockSubject" {
		t.Errorf("AddRule failed to capture subject type")
	}
	if b.rules[0].Action != "read" {
		t.Errorf("AddRule failed to capture action")
	}

	// Test AddRules
	AddRules(b, r, r)
	if len(b.rules) != 3 {
		t.Errorf("AddRules failed")
	}

	// Test Build
	a := b.Build()
	if a.index == nil {
		t.Errorf("Build failed to create index")
	}
	if a.compiler == nil {
		t.Errorf("Build failed to create compiler")
	}
}
