package gocasl

import (
	"testing"
)

func TestRuleIndex(t *testing.T) {
	compiler := newCompiler(defaultOperators(), nil)
	
	rules := []rawRule{
		{
			SubjectType: "Post",
			Action:      "read",
			Conditions:  Cond{"ID": 1},
		},
		{
			SubjectType: "Post",
			Action:      "update",
			Conditions:  Cond{"ID": 2},
		},
		{
			SubjectType: "Comment",
			Action:      "read",
			Conditions:  Cond{"ID": 3},
		},
	}

	idx := newRuleIndex(rules, compiler)

	// Test lookup
	postRead := idx.get("Post", "read")
	if len(postRead) != 1 {
		t.Errorf("Expected 1 Post/read rule, got %d", len(postRead))
	}

	postUpdate := idx.get("Post", "update")
	if len(postUpdate) != 1 {
		t.Errorf("Expected 1 Post/update rule, got %d", len(postUpdate))
	}

	commentRead := idx.get("Comment", "read")
	if len(commentRead) != 1 {
		t.Errorf("Expected 1 Comment/read rule, got %d", len(commentRead))
	}

	missing := idx.get("Post", "delete")
	if len(missing) != 0 {
		t.Errorf("Expected 0 Post/delete rules, got %d", len(missing))
	}
}

func TestLazyCompilation(t *testing.T) {
	// We want to verify that compilation happens only when match is called.
	// We can't easily mock the compiler inside standard package test without exposing internals.
	// But we can check if condition is nil before match and not nil after.
	
	compiler := newCompiler(defaultOperators(), nil)
	r := rawRule{
		SubjectType: "Post",
		Action:      "read",
		Conditions:  Cond{"ID": 1},
	}
	
	idx := newRuleIndex([]rawRule{r}, compiler)
	crs := idx.get("Post", "read")
	cr := crs[0]

	if cr.condition != nil {
		t.Errorf("Condition should be nil before first match (lazy compilation)")
	}

	sub := mockSubject{ID: 1, Title: "Test"} // Using mockSubject but treating as Post for this test?
	// Wait, match takes Subject interface. mockSubject implements it.
	// But SubjectType must match? No, match() just runs the condition. The index lookup handled the type.
	
	cr.match(sub)

	if cr.condition == nil {
		t.Errorf("Condition should be compiled after match")
	}
}
