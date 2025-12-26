package gocasl

import (
	"reflect"
	"testing"
)

func TestVar(t *testing.T) {
	v := Var("UserID")
	if v != "UserID" {
		t.Errorf("Expected VarRef 'UserID', got %s", v)
	}
}

func TestConditionHelpers(t *testing.T) {
	c1 := Cond{"status": "active"}
	c2 := Cond{"age": Op{"$gt": 18}}
	c3 := Cond{"role": "admin"}

	// Test And
	andCond := And(c1, c2)
	expectedAnd := Cond{
		"$and": []any{c1, c2},
	}
	if !reflect.DeepEqual(andCond, expectedAnd) {
		t.Errorf("And() failed.\nExpected: %v\nGot: %v", expectedAnd, andCond)
	}

	// Test Or
	orCond := Or(c1, c3)
	expectedOr := Cond{
		"$or": []any{c1, c3},
	}
	if !reflect.DeepEqual(orCond, expectedOr) {
		t.Errorf("Or() failed.\nExpected: %v\nGot: %v", expectedOr, orCond)
	}

	// Test Not
	notCond := Not(c1)
	expectedNot := Cond{
		"$not": c1,
	}
	if !reflect.DeepEqual(notCond, expectedNot) {
		t.Errorf("Not() failed.\nExpected: %v\nGot: %v", expectedNot, notCond)
	}

	// Test Nested
	nested := And(c1, Or(c2, c3))
	expectedNested := Cond{
		"$and": []any{
			c1,
			Cond{"$or": []any{c2, c3}},
		},
	}
	if !reflect.DeepEqual(nested, expectedNested) {
		t.Errorf("Nested helpers failed.\nExpected: %v\nGot: %v", expectedNested, nested)
	}
}
