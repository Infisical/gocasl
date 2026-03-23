package gocasl

import (
	"reflect"
	"testing"
)

func TestFieldOpsMethods(t *testing.T) {
	ops := defaultFieldOps().Clone()

	// Test Clone
	if len(ops) != len(defaultFieldOps()) {
		t.Errorf("Clone failed size check")
	}

	// Test With
	ops = ops.With("$custom", Compare(func(a, b any) bool { return true }))
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
	ops2 := FieldOps{"$new": Compare(func(a, b any) bool { return true })}
	ops = ops.WithAll(ops2)
	if _, ok := ops["$new"]; !ok {
		t.Errorf("WithAll failed to add $new")
	}
}

func TestCondOpsMethods(t *testing.T) {
	ops := defaultCondOps().Clone()

	if len(ops) != len(defaultCondOps()) {
		t.Errorf("Clone failed size check")
	}

	ops = ops.With("$custom", func(cc *CompileCtx, value any) Condition {
		return func(s Subject) bool { return true }
	})
	if _, ok := ops["$custom"]; !ok {
		t.Errorf("With failed to add cond operator")
	}

	ops = ops.Without("$custom")
	if _, ok := ops["$custom"]; ok {
		t.Errorf("Without failed to remove $custom")
	}
}

// helper to evaluate a comparison function against field/constraint values
func evalFieldOp(op FieldOp, fieldVal, constraint any) bool {
	cc := &CompileCtx{
		Compile: func(c Cond) Condition { return func(s Subject) bool { return true } },
		Resolve: func(v any) any { return v },
	}
	cond := op(cc, "testField", constraint)
	sub := MapSubject{"testField": fieldVal}
	return cond(sub)
}

func TestComparisonOperators(t *testing.T) {
	ops := defaultFieldOps()

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
		fn := ops[tt.op]
		if got := evalFieldOp(fn, tt.val, tt.cons); got != tt.want {
			t.Errorf("%s(%v, %v) = %v; want %v", tt.op, tt.val, tt.cons, got, tt.want)
		}
	}
}

func TestArrayOperators(t *testing.T) {
	ops := defaultFieldOps()

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
		fn := ops[tt.op]
		if got := evalFieldOp(fn, tt.val, tt.cons); got != tt.want {
			t.Errorf("%s(%v, %v) = %v; want %v", tt.op, tt.val, tt.cons, got, tt.want)
		}
	}
}

func TestStringOperators(t *testing.T) {
	ops := defaultFieldOps()

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
		fn := ops[tt.op]
		if got := evalFieldOp(fn, tt.val, tt.cons); got != tt.want {
			t.Errorf("%s(%v, %v) = %v; want %v", tt.op, tt.val, tt.cons, got, tt.want)
		}
	}
}

func TestOtherOperators(t *testing.T) {
	ops := defaultFieldOps()

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
		fn := ops[tt.op]
		if got := evalFieldOp(fn, tt.val, tt.cons); got != tt.want {
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

	// Test ElemMatch
	em := ElemMatch(Cond{"status": "active"})
	if val, ok := em["$elemMatch"]; !ok {
		t.Errorf("ElemMatch missing key $elemMatch")
	} else if !reflect.DeepEqual(val, Cond{"status": "active"}) {
		t.Errorf("ElemMatch value mismatch: %v", val)
	}

	// Test All
	all := All(Cond{"score": Gte(50)})
	if _, ok := all["$all"]; !ok {
		t.Errorf("All missing key $all")
	}

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

func TestElemMatchOperator(t *testing.T) {
	// Subject with array of map elements
	type Item struct {
		Category string
		Score    int
	}

	type Doc struct {
		Items []map[string]any
	}

	sub := MapSubject{
		"items": []map[string]any{
			{"category": "sports", "score": 90},
			{"category": "tech", "score": 85},
			{"category": "tech", "score": 50},
		},
	}

	compiler := newCompiler(nil, nil, nil)

	tests := []struct {
		name string
		cond Cond
		want bool
	}{
		{
			name: "elemMatch matches element with all conditions",
			cond: Cond{"items": ElemMatch(Cond{
				"category": "tech",
				"score":    Op{"$gte": 80},
			})},
			want: true,
		},
		{
			name: "elemMatch fails when no single element matches all",
			cond: Cond{"items": ElemMatch(Cond{
				"category": "sports",
				"score":    Op{"$lt": 50},
			})},
			want: false,
		},
		{
			name: "elemMatch with simple equality",
			cond: Cond{"items": ElemMatch(Cond{
				"category": "tech",
			})},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			condFunc := compiler.compile(tt.cond)
			if got := condFunc(sub); got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAllOperator(t *testing.T) {
	sub := MapSubject{
		"scores": []map[string]any{
			{"value": 80, "passing": true},
			{"value": 90, "passing": true},
			{"value": 85, "passing": true},
		},
	}

	subWithFailing := MapSubject{
		"scores": []map[string]any{
			{"value": 80, "passing": true},
			{"value": 30, "passing": false},
			{"value": 85, "passing": true},
		},
	}

	compiler := newCompiler(nil, nil, nil)

	tests := []struct {
		name    string
		subject Subject
		cond    Cond
		want    bool
	}{
		{
			name:    "all elements match",
			subject: sub,
			cond: Cond{"scores": All(Cond{
				"passing": true,
			})},
			want: true,
		},
		{
			name:    "not all elements match",
			subject: subWithFailing,
			cond: Cond{"scores": All(Cond{
				"passing": true,
			})},
			want: false,
		},
		{
			name:    "all with operator condition",
			subject: sub,
			cond: Cond{"scores": All(Cond{
				"value": Op{"$gte": 75},
			})},
			want: true,
		},
		{
			name:    "all fails when one element doesn't match operator",
			subject: subWithFailing,
			cond: Cond{"scores": All(Cond{
				"value": Op{"$gte": 75},
			})},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			condFunc := compiler.compile(tt.cond)
			if got := condFunc(tt.subject); got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
