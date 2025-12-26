package gocasl

import (
	"testing"
)

func TestIntrospection(t *testing.T) {
	read := DefineAction[mockSubject]("read")
	sub := mockSubject{ID: 1}

	t.Run("WhyNot", func(t *testing.T) {
		b := NewAbility()
		AddRule(b, Forbid(read).Because("Private").Build())
		a := b.Build()

		reason := WhyNot(a, read, sub)
		if reason != "Private" {
			t.Errorf("WhyNot expected 'Private', got '%s'", reason)
		}

		// Test generic deny
		b = NewAbility()
		a = b.Build()
		reason = WhyNot(a, read, sub)
		if reason == "" {
			t.Errorf("WhyNot expected reason for default deny")
		}

		// Test allowed
		b = NewAbility()
		AddRule(b, Allow(read).Build())
		a = b.Build()
		reason = WhyNot(a, read, sub)
		if reason != "" {
			t.Errorf("WhyNot expected empty string, got '%s'", reason)
		}
	})

	t.Run("RulesFor", func(t *testing.T) {
		b := NewAbility()
		AddRule(b, Allow(read).Where(Cond{"ID": 1}).Build())
		AddRule(b, Allow(read).Where(Cond{"ID": 2}).Build())
		a := b.Build()

		rules := RulesFor(a, read, sub)
		if len(rules) != 2 {
			t.Errorf("RulesFor expected 2 rules")
		}
		if !rules[0].Matched {
			t.Errorf("Rule 1 should match")
		}
		if rules[1].Matched {
			t.Errorf("Rule 2 should not match")
		}
	})

	t.Run("AllowedFields", func(t *testing.T) {
		b := NewAbility()
		AddRule(b, Allow(read).OnFields("A", "B").Build())
		AddRule(b, Allow(read).OnFields("B", "C").Build())
		AddRule(b, Forbid(read).OnFields("C").Build())
		a := b.Build()

		fields := AllowedFields(a, read, sub)
		// Expected: A, B (C is forbidden)
		if len(fields) != 2 {
			t.Errorf("Expected 2 allowed fields")
		}
		// Order might vary if using map iteration, but implementation uses append order
		// Wait, implementation uses iteration over allowed slice then check seen.
		// Order should be A, B, B, C -> filter C -> A, B, B -> unique -> A, B.
		// It preserves order of appearance.
		if fields[0] != "A" || fields[1] != "B" {
			t.Errorf("Expected [A, B], got %v", fields)
		}

		// Test wildcard
		b = NewAbility()
		AddRule(b, Allow(read).Build())
		a = b.Build()
		if AllowedFields(a, read, sub) != nil {
			t.Errorf("Expected nil for wildcard allowed")
		}
		
		// Test full forbid
		b = NewAbility()
		AddRule(b, Allow(read).OnFields("A").Build())
		AddRule(b, Forbid(read).Build())
		a = b.Build()
		if len(AllowedFields(a, read, sub)) != 0 {
			t.Errorf("Expected empty allowed fields for full forbid")
		}
	})

	t.Run("ForbiddenFields", func(t *testing.T) {
		b := NewAbility()
		AddRule(b, Forbid(read).OnFields("A", "B").Build())
		a := b.Build()

		fields := ForbiddenFields(a, read, sub)
		if len(fields) != 2 {
			t.Errorf("Expected 2 forbidden fields")
		}

		// Test wildcard
		b = NewAbility()
		AddRule(b, Forbid(read).Build())
		a = b.Build()
		if ForbiddenFields(a, read, sub) != nil {
			t.Errorf("Expected nil for wildcard forbidden")
		}
	})
}
