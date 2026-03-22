package gocasl

import (
	"testing"
)

func TestPermissions(t *testing.T) {
	// Setup actions
	read := DefineAction[mockSubject]("read")
	update := DefineAction[mockSubject]("update")
	deleteOp := DefineAction[mockSubject]("delete")

	// Setup subject
	sub := mockSubject{ID: 1, Title: "My Post", Tags: []string{"go"}}

	t.Run("Basic Allow", func(t *testing.T) {
		b := NewAbility()
		AddRule(b, Allow(read).Build())
		a, err := b.Build()
		if err != nil {
			t.Fatal(err)
		}

		if !Can(a, read, sub) {
			t.Errorf("Should allow read")
		}
		if Can(a, update, sub) {
			t.Errorf("Should not allow update")
		}
	})

	t.Run("Allow with Condition", func(t *testing.T) {
		b := NewAbility()
		AddRule(b, Allow(read).Where(Cond{"ID": 1}).Build())
		AddRule(b, Allow(update).Where(Cond{"ID": 2}).Build())
		a, err := b.Build()
		if err != nil {
			t.Fatal(err)
		}

		if !Can(a, read, sub) {
			t.Errorf("Should allow read (ID=1)")
		}
		if Can(a, update, sub) {
			t.Errorf("Should not allow update (ID!=2)")
		}
	})

	t.Run("Forbid Precedence", func(t *testing.T) {
		b := NewAbility()
		AddRule(b, Allow(read).Build())
		AddRule(b, Forbid(read).Where(Cond{"ID": 1}).Build())
		a, err := b.Build()
		if err != nil {
			t.Fatal(err)
		}

		if Can(a, read, sub) {
			t.Errorf("Should forbid read (ID=1 matches forbid)")
		}

		sub2 := mockSubject{ID: 2}
		if !Can(a, read, sub2) {
			t.Errorf("Should allow read for ID=2")
		}
	})

	t.Run("Field Permissions", func(t *testing.T) {
		b := NewAbility()
		// Allow reading everything except Title
		AddRule(b, Allow(read).Build())
		AddRule(b, Forbid(read).OnFields("Title").Build())
		a, err := b.Build()
		if err != nil {
			t.Fatal(err)
		}

		// Resource level access should be true (can read other fields)
		if !Can(a, read, sub) {
			t.Errorf("Can(read) should be true (partial access)")
		}

		// Field level checks
		if CanWithField(a, read, sub, "Title") {
			t.Errorf("CanWithField(Title) should be false")
		}
		if !CanWithField(a, read, sub, "ID") {
			t.Errorf("CanWithField(ID) should be true")
		}
	})

	t.Run("Allow Specific Fields", func(t *testing.T) {
		b := NewAbility()
		AddRule(b, Allow(read).OnFields("ID").Build())
		a, err := b.Build()
		if err != nil {
			t.Fatal(err)
		}

		if !Can(a, read, sub) {
			t.Errorf("Can(read) should be true (partial access)")
		}

		if !CanWithField(a, read, sub, "ID") {
			t.Errorf("CanWithField(ID) should be true")
		}
		if CanWithField(a, read, sub, "Title") {
			t.Errorf("CanWithField(Title) should be false")
		}
	})

	t.Run("Cannot", func(t *testing.T) {
		b := NewAbility()
		a, err := b.Build()
		if err != nil {
			t.Fatal(err)
		}
		if !Cannot(a, read, sub) {
			t.Errorf("Cannot should be true when no rules")
		}
	})

	t.Run("CanAll and CanAny", func(t *testing.T) {
		b := NewAbility()
		AddRule(b, Allow(read).Build())
		AddRule(b, Allow(update).Build())
		a, err := b.Build()
		if err != nil {
			t.Fatal(err)
		}

		if !CanAll(a, sub, read, update) {
			t.Errorf("CanAll should be true")
		}
		if CanAll(a, sub, read, update, deleteOp) {
			t.Errorf("CanAll should be false (delete missing)")
		}

		if !CanAny(a, sub, read, deleteOp) {
			t.Errorf("CanAny should be true")
		}
		if CanAny(a, sub, deleteOp) {
			t.Errorf("CanAny should be false")
		}
	})

	t.Run("Priority Order Independence", func(t *testing.T) {
		read := DefineAction[mockSubject]("read")
		sub := mockSubject{ID: 1}

		t.Run("Allow then Forbid", func(t *testing.T) {
			b := NewAbility()
			AddRule(b, Allow(read).Build())
			AddRule(b, Forbid(read).Where(Cond{"ID": 1}).Build())
			a, _ := b.Build()

			if Can(a, read, sub) {
				t.Errorf("Should forbid read (ID=1 matches forbid)")
			}
		})

		t.Run("Forbid then Allow", func(t *testing.T) {
			b := NewAbility()
			AddRule(b, Forbid(read).Where(Cond{"ID": 1}).Build())
			AddRule(b, Allow(read).Build())
			a, _ := b.Build()

			if Can(a, read, sub) {
				t.Errorf("Should forbid read (ID=1 matches forbid)")
			}
		})

		t.Run("Forbid everything then Allow", func(t *testing.T) {
			b := NewAbility()
			AddRule(b, Forbid(read).Build())
			AddRule(b, Allow(read).Build())
			a, _ := b.Build()

			if Can(a, read, sub) {
				t.Errorf("Should forbid read")
			}
		})
	})
}
