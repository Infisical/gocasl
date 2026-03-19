package gocasl

import (
	"testing"
)

// Define domain models for integration test
type User struct {
	ID    int
	Roles []string
}

type Article struct {
	ID        int
	AuthorID  int
	Published bool
	Content   string
}

func (a Article) SubjectType() string { return "Article" }
func (a Article) GetField(f string) any {
	switch f {
	case "ID": return a.ID
	case "AuthorID": return a.AuthorID
	case "Published": return a.Published
	case "Content": return a.Content
	}
	return nil
}

func TestBlogIntegration(t *testing.T) {
	// Define actions
	read := DefineAction[Article]("read")
	update := DefineAction[Article]("update")
	deleteOp := DefineAction[Article]("delete")
	publish := DefineAction[Article]("publish")

	// Users
	// admin := User{ID: 1, Roles: []string{"admin"}}
	author := User{ID: 2, Roles: []string{"author"}}
	// guest := User{ID: 3, Roles: []string{"guest"}}

	// Articles
	myArticle := Article{ID: 100, AuthorID: 2, Published: false, Content: "Draft"}
	otherArticle := Article{ID: 101, AuthorID: 4, Published: true, Content: "Public"}
	
	// Define Rules for each role
	
	// Admin Ability
	adminBuilder := NewAbility()
	// Admin can manage everything
	AddRule(adminBuilder, Allow(read).Build())
	AddRule(adminBuilder, Allow(update).Build())
	AddRule(adminBuilder, Allow(deleteOp).Build())
	AddRule(adminBuilder, Allow(publish).Build())
	// In reality we might use "manage" "all" but we use specific actions here
	adminAbility, err := adminBuilder.Build()
	if err != nil {
		t.Fatal(err)
	}

	if !Can(adminAbility, deleteOp, otherArticle) {
		t.Errorf("Admin should be able to delete any article")
	}

	// Author Ability
	authorBuilder := NewAbility()
	authorBuilder.WithVars(map[string]any{"UserID": author.ID})
	
	// Can read any published article
	AddRule(authorBuilder, Allow(read).Where(Cond{"Published": true}).Build())
	// Can read own articles
	AddRule(authorBuilder, Allow(read).Where(Cond{"AuthorID": Var("UserID")}).Build())
	// Can update own articles if not published (naive check, usually separate rule)
	AddRule(authorBuilder, Allow(update).Where(Cond{"AuthorID": Var("UserID")}).Build())
	// Cannot delete anything (implicit)
	
	authorAbility, err := authorBuilder.Build()
	if err != nil {
		t.Fatal(err)
	}

	if !Can(authorAbility, read, otherArticle) {
		t.Errorf("Author should be able to read public article")
	}
	if !Can(authorAbility, read, myArticle) {
		t.Errorf("Author should be able to read own draft")
	}
	if !Can(authorAbility, update, myArticle) {
		t.Errorf("Author should be able to update own draft")
	}
	if Can(authorAbility, update, otherArticle) {
		t.Errorf("Author should not be able to update others article")
	}
	if Can(authorAbility, deleteOp, myArticle) {
		t.Errorf("Author should not be able to delete own article (not granted)")
	}

	// Guest Ability via JSON
	jsonRules := `[
		{
			"action": "read",
			"subject": "Article",
			"conditions": {"Published": true}
		}
	]`
	
	guestAbility, err := LoadFromJSON([]byte(jsonRules), LoadOptions{})
	if err != nil {
		t.Fatalf("Failed to load guest rules: %v", err)
	}

	if !Can(guestAbility, read, otherArticle) {
		t.Errorf("Guest should be able to read public article")
	}
	if Can(guestAbility, read, myArticle) {
		t.Errorf("Guest should not be able to read draft article")
	}
	if Can(guestAbility, update, otherArticle) {
		t.Errorf("Guest should not be able to update")
	}
}
