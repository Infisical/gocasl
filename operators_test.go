package gocasl

import (
	"reflect"
	"testing"
)

func TestOperatorsMethods(t *testing.T) {
	ops := defaultOperators().Clone()

	// Test Clone
	if len(ops) != len(defaultOperators()) {
		t.Errorf("Clone failed size check")
	}

	// Test With
	ops = ops.With("$custom", func(a, b any) bool { return true })
	if _, ok := ops["$custom"]; !ok {
		t.Errorf("With failed to add operator")
	}

	// Test Without
	ops = ops.Without("$custom", "$eq")
	if _, ok := ops["$custom"]; ok {
		t.Errorf("Without failed to remove $custom")
	}
	if _, ok := ops["$eq"]; ok {
		t.Errorf("Without failed to remove $eq")
	}

	// Test WithAll
	ops2 := Operators{"$new": func(a, b any) bool { return true }}
	ops = ops.WithAll(ops2)
	if _, ok := ops["$new"]; !ok {
		t.Errorf("WithAll failed to add $new")
	}
}

func TestComparisonOperators(t *testing.T) {
	tests := []struct {
		op   string
		val  any
		cons any
		want bool
	}{
		// $eq
		{"$eq", 1, 1, true},
		{"$eq", 1, 2, false},
		{"$eq", "a", "a", true},
		{"$eq", 1, 1.0, true}, // numeric coercion

		// $ne
		{"$ne", 1, 2, true},
		{"$ne", 1, 1, false},

		// $gt
		{"$gt", 2, 1, true},
		{"$gt", 1, 1, false},
		{"$gt", 0, 1, false},

		// $gte
		{"$gte", 2, 1, true},
		{"$gte", 1, 1, true},
		{"$gte", 0, 1, false},

		// $lt
		{"$lt", 0, 1, true},
		{"$lt", 1, 1, false},

		// $lte
		{"$lte", 0, 1, true},
		{"$lte", 1, 1, true},
		{"$lte", 2, 1, false},
	}

	for _, tt := range tests {
		fn := defaultOperators()[tt.op]
		if got := fn(tt.val, tt.cons); got != tt.want {
			t.Errorf("%s(%v, %v) = %v; want %v", tt.op, tt.val, tt.cons, got, tt.want)
		}
	}
}

func TestArrayOperators(t *testing.T) {
	tests := []struct {
		op   string
		val  any
		cons any
		want bool
	}{
		// $in
		{"$in", 1, []int{1, 2, 3}, true},
		{"$in", 4, []int{1, 2, 3}, false},
		{"$in", "a", []string{"a", "b"}, true},
		
		// $nin
		{"$nin", 1, []int{1, 2, 3}, false},
		{"$nin", 4, []int{1, 2, 3}, true},
	}

	for _, tt := range tests {
		fn := defaultOperators()[tt.op]
		if got := fn(tt.val, tt.cons); got != tt.want {
			t.Errorf("%s(%v, %v) = %v; want %v", tt.op, tt.val, tt.cons, got, tt.want)
		}
	}
}

func TestStringOperators(t *testing.T) {
	tests := []struct {
		op   string
		val  any
		cons any
		want bool
	}{
		// $regex
		{"$regex", "hello world", "^hello", true},
		{"$regex", "hello world", "world$", true},
		{"$regex", "hello world", "^world", false},

		// $contains
		{"$contains", "hello world", "world", true},
		{"$contains", "hello world", "foo", false},
		{"$contains", []string{"a", "b"}, "a", true}, // contains on array
	}

	for _, tt := range tests {
		fn := defaultOperators()[tt.op]
		if got := fn(tt.val, tt.cons); got != tt.want {
			t.Errorf("%s(%v, %v) = %v; want %v", tt.op, tt.val, tt.cons, got, tt.want)
		}
	}
}

func TestOtherOperators(t *testing.T) {
	tests := []struct {
		op   string
		val  any
		cons any
		want bool
	}{
		// $exists
		{"$exists", 1, true, true},
		{"$exists", nil, true, false},
		{"$exists", 1, false, false},
		{"$exists", nil, false, true},

		// $size
		{"$size", "hello", 5, true},
		{"$size", []int{1, 2}, 2, true},
		{"$size", []int{1}, 2, false},
	}

	for _, tt := range tests {
		fn := defaultOperators()[tt.op]
		if got := fn(tt.val, tt.cons); got != tt.want {
			t.Errorf("%s(%v, %v) = %v; want %v", tt.op, tt.val, tt.cons, got, tt.want)
		}
	}
}

func TestDSL(t *testing.T) {
	// Check if DSL returns correct Op structure
	check := func(name string, op Op, expectedKey string, expectedVal any) {
		if val, ok := op[expectedKey]; !ok {
			t.Errorf("%s missing key %s", name, expectedKey)
		} else if !reflect.DeepEqual(val, expectedVal) {
			t.Errorf("%s value mismatch. Got %v, want %v", name, val, expectedVal)
		}
	}

	check("Eq", Eq(1), "$eq", 1)
	check("Ne", Ne(1), "$ne", 1)
	check("Gt", Gt(1), "$gt", 1)
	check("Gte", Gte(1), "$gte", 1)
	check("Lt", Lt(1), "$lt", 1)
	check("Lte", Lte(1), "$lte", 1)
	check("In", In(1, 2), "$in", []any{1, 2})
	check("Nin", Nin(1, 2), "$nin", []any{1, 2})
	check("Regex", Regex(".*"), "$regex", ".*")
	check("Contains", Contains("a"), "$contains", "a")
	check("Exists", Exists(true), "$exists", true)
	check("Size", Size(5), "$size", 5)

	// Test StartsWith / EndsWith
	sw := StartsWith("foo")
	if val, ok := sw["$regex"]; !ok || val != "^foo" {
		t.Errorf("StartsWith failed: %v", sw)
	}
	
	ew := EndsWith("bar")
	if val, ok := ew["$regex"]; !ok || val != "bar$" {
		t.Errorf("EndsWith failed: %v", ew)
	}
}
