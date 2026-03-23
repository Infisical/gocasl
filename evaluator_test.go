package gocasl

import (
	"testing"
)

func TestEvaluator(t *testing.T) {
	// Setup subject
	sub := mockSubject{
		ID:    123,
		Title: "Test Article",
		Tags:  []string{"news", "go"},
	}

	// Setup vars
	vars := map[string]any{
		"UserID": 123,
		"Admin":  true,
	}

	compiler := newCompiler(defaultFieldOps(), defaultCondOps(), vars)

	tests := []struct {
		name string
		cond Cond
		want bool
	}{
		{
			name: "Simple Equality",
			cond: Cond{"ID": 123},
			want: true,
		},
		{
			name: "Simple Equality Fail",
			cond: Cond{"ID": 456},
			want: false,
		},
		{
			name: "Implicit Eq with String",
			cond: Cond{"Title": "Test Article"},
			want: true,
		},
		{
			name: "Operator Gt",
			cond: Cond{"ID": Op{"$gt": 100}},
			want: true,
		},
		{
			name: "Operator Gt Fail",
			cond: Cond{"ID": Op{"$gt": 200}},
			want: false,
		},
		{
			name: "Operator In",
			cond: Cond{"ID": Op{"$in": []int{123, 456}}},
			want: true,
		},
		{
			name: "Operator Contains (Array Field)",
			cond: Cond{"Tags": Op{"$contains": "go"}},
			want: true,
		},
		{
			name: "Variable Resolution",
			cond: Cond{"ID": Var("UserID")},
			want: true,
		},
		{
			name: "Variable Resolution Fail",
			cond: Cond{"ID": Var("OtherVar")}, // OtherVar missing -> nil -> 123 != nil
			want: false,
		},
		{
			name: "And Condition",
			cond: And(
				Cond{"ID": 123},
				Cond{"Title": "Test Article"},
			),
			want: true,
		},
		{
			name: "And Condition Fail",
			cond: And(
				Cond{"ID": 123},
				Cond{"Title": "Wrong"},
			),
			want: false,
		},
		{
			name: "Or Condition",
			cond: Or(
				Cond{"ID": 999},
				Cond{"ID": 123},
			),
			want: true,
		},
		{
			name: "Not Condition",
			cond: Not(Cond{"ID": 999}),
			want: true,
		},
		{
			name: "Nested Logic",
			cond: And(
				Cond{"ID": 123},
				Or(
					Cond{"Title": "Wrong"},
					Cond{"Tags": Op{"$contains": "news"}},
				),
			),
			want: true,
		},
		{
			name: "Var in Operator",
			cond: Cond{"ID": Op{"$eq": Var("UserID")}},
			want: true,
		},
		{
			name: "Missing Field",
			cond: Cond{"MissingField": Op{"$exists": false}},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			condFunc := compiler.compile(tt.cond)
			if got := condFunc(sub); got != tt.want {
				t.Errorf("Condition %v failed. Want %v, got %v", tt.cond, tt.want, got)
			}
		})
	}
}
