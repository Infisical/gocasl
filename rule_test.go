package gocasl

import (
	"reflect"
	"testing"
)

func TestRuleBuilder(t *testing.T) {
	readAction := DefineAction[mockSubject]("read")

	// Test Allow
	r1 := Allow(readAction).Build()
	if r1.Inverted {
		t.Errorf("Allow rule should not be inverted")
	}
	if r1.Action.Name() != "read" {
		t.Errorf("Rule action name mismatch")
	}

	// Test Forbid
	r2 := Forbid(readAction).Build()
	if !r2.Inverted {
		t.Errorf("Forbid rule should be inverted")
	}

	// Test Where
	cond := Cond{"ID": 123}
	r3 := Allow(readAction).Where(cond).Build()
	if !reflect.DeepEqual(r3.Conditions, cond) {
		t.Errorf("Where failed to set conditions")
	}

	// Test OnFields
	r4 := Allow(readAction).OnFields("Title", "Tags").Build()
	if !reflect.DeepEqual(r4.Fields, []string{"Title", "Tags"}) {
		t.Errorf("OnFields failed to set fields")
	}

	// Test Because
	r5 := Forbid(readAction).Because("Not authorized").Build()
	if r5.Reason != "Not authorized" {
		t.Errorf("Because failed to set reason")
	}

	// Test Chaining
	r6 := Allow(readAction).
		Where(cond).
		OnFields("Title").
		Because("Owner").
		Build()

	if r6.Inverted {
		t.Errorf("Chained rule inverted wrong")
	}
	if !reflect.DeepEqual(r6.Conditions, cond) {
		t.Errorf("Chained rule cond wrong")
	}
	if !reflect.DeepEqual(r6.Fields, []string{"Title"}) {
		t.Errorf("Chained rule fields wrong")
	}
	if r6.Reason != "Owner" {
		t.Errorf("Chained rule reason wrong")
	}
}
