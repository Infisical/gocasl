package gocasl

import (
	"testing"
)

func TestAbilityBuilder(t *testing.T) {
	// Test NewAbility
	b := NewAbility()
	if b.fieldOps == nil {
		t.Errorf("NewAbility should set default field operators")
	}
	if b.condOps == nil {
		t.Errorf("NewAbility should set default cond operators")
	}

	// Test WithFieldOps
	ops := FieldOps{}
	b.WithFieldOps(ops)
	if len(b.fieldOps) != 0 {
		t.Errorf("WithFieldOps failed")
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

	// Need to restore field ops for Build to work
	b.WithFieldOps(defaultFieldOps())

	// Test Build
	a, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	if a.index == nil {
		t.Errorf("Build failed to create index")
	}
	if a.compiler == nil {
		t.Errorf("Build failed to create compiler")
	}
}
