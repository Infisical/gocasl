package gocasl

import (
	"testing"
)

type anotherMockSubject struct {
	ID int
}

func (anotherMockSubject) SubjectType() string {
	return "anotherMockSubject"
}

func (s anotherMockSubject) GetField(field string) any {
	switch field {
	case "ID":
		return s.ID
	default:
		return nil
	}
}

func TestIntrospection(t *testing.T) {
	read := DefineAction[mockSubject]("read")
	sub := mockSubject{ID: 1}

	t.Run("WhyNot", func(t *testing.T) {
		b := NewAbility()
		AddRule(b, Forbid(read).Because("Private").Build())
		a, err := b.Build()
		if err != nil {
			t.Fatal(err)
		}

		reason := WhyNot(a, read, sub)
		if reason != "Private" {
			t.Errorf("WhyNot expected 'Private', got '%s'", reason)
		}

		// Test generic deny
		b = NewAbility()
		a, err = b.Build()
		if err != nil {
			t.Fatal(err)
		}
		reason = WhyNot(a, read, sub)
		if reason == "" {
			t.Errorf("WhyNot expected reason for default deny")
		}

		// Test allowed
		b = NewAbility()
		AddRule(b, Allow(read).Build())
		a, err = b.Build()
		if err != nil {
			t.Fatal(err)
		}
		reason = WhyNot(a, read, sub)
		if reason != "" {
			t.Errorf("WhyNot expected empty string, got '%s'", reason)
		}
	})

	t.Run("RulesFor", func(t *testing.T) {
		b := NewAbility()
		AddRule(b, Allow(read).Where(Cond{"ID": 1}).Build())
		AddRule(b, Allow(read).Where(Cond{"ID": 2}).Build())
		a, err := b.Build()
		if err != nil {
			t.Fatal(err)
		}

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
		a, err := b.Build()
		if err != nil {
			t.Fatal(err)
		}

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
		a, err = b.Build()
		if err != nil {
			t.Fatal(err)
		}
		if AllowedFields(a, read, sub) != nil {
			t.Errorf("Expected nil for wildcard allowed")
		}

		// Test full forbid
		b = NewAbility()
		AddRule(b, Allow(read).OnFields("A").Build())
		AddRule(b, Forbid(read).Build())
		a, err = b.Build()
		if err != nil {
			t.Fatal(err)
		}
		if len(AllowedFields(a, read, sub)) != 0 {
			t.Errorf("Expected empty allowed fields for full forbid")
		}
	})

	t.Run("ForbiddenFields", func(t *testing.T) {
		b := NewAbility()
		AddRule(b, Forbid(read).OnFields("A", "B").Build())
		a, err := b.Build()
		if err != nil {
			t.Fatal(err)
		}

		fields := ForbiddenFields(a, read, sub)
		if len(fields) != 2 {
			t.Errorf("Expected 2 forbidden fields")
		}

		// Test wildcard
		b = NewAbility()
		AddRule(b, Forbid(read).Build())
		a, err = b.Build()
		if err != nil {
			t.Fatal(err)
		}
		if ForbiddenFields(a, read, sub) != nil {
			t.Errorf("Expected nil for wildcard forbidden")
		}
	})

	t.Run("PossibleRulesFor", func(t *testing.T) {
		update := DefineAction[mockSubject]("update")
		create := DefineAction[anotherMockSubject]("create") // Different subject type

		b := NewAbility()
		AddRule(b, Allow(read).Where(Cond{"ID": 1}).Build())    // Rule 1 for mockSubject, read
		AddRule(b, Allow(read).OnFields("Title").Build())       // Rule 2 for mockSubject, read
		AddRule(b, Forbid(update).Because("No update").Build()) // Rule 3 for mockSubject, update
		AddRule(b, Allow(create).Build())                       // Rule 4 for anotherMockSubject, create
		a, err := b.Build()
		if err != nil {
			t.Fatal(err)
		}

		// Test for mockSubject, read action
		rules := PossibleRulesFor(a, read)
		if len(rules) != 2 {
			t.Errorf("PossibleRulesFor for read (mockSubject) expected 2 rules, got %d", len(rules))
		}
		for _, r := range rules {
			if r.Action != "read" {
				t.Errorf("Expected action 'read', got '%s'", r.Action)
			}
			if r.SubjectType != "MockSubject" {
				t.Errorf("Expected subjectType 'mockSubject', got '%s'", r.SubjectType)
			}
			if r.Matched {
				t.Errorf("Expected Matched to be false, got true")
			}
		}

		// Test for mockSubject, update action
		updateRules := PossibleRulesFor(a, update)
		if len(updateRules) != 1 {
			t.Errorf("PossibleRulesFor for update (mockSubject) expected 1 rule, got %d", len(updateRules))
		}
		if updateRules[0].Action != "update" || updateRules[0].SubjectType != "MockSubject" || !updateRules[0].Inverted || updateRules[0].Reason != "No update" {
			t.Errorf("Update rule mismatch: %+v", updateRules[0])
		}

		// Test for a different subject type, create action
		createRules := PossibleRulesFor(a, create)
		if len(createRules) != 1 {
			t.Errorf("PossibleRulesFor for create (anotherMockSubject) expected 1 rule, got %d", len(createRules))
		}
		if createRules[0].Action != "create" || createRules[0].SubjectType != "anotherMockSubject" {
			t.Errorf("Create rule mismatch: %+v", createRules[0])
		}

		// Test for non-existent action/subject combination
		nonExistentAction := DefineAction[mockSubject]("nonExistent")
		emptyRules := PossibleRulesFor(a, nonExistentAction)
		if len(emptyRules) != 0 {
			t.Errorf("PossibleRulesFor for non-existent action expected 0 rules, got %d", len(emptyRules))
		}
	})
}
