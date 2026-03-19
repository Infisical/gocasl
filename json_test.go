package gocasl

import (
	"encoding/json"
	"testing"
)

func TestStringOrSlice(t *testing.T) {
	var s StringOrSlice

	// Test single string
	if err := json.Unmarshal([]byte(`"foo"`), &s); err != nil {
		t.Errorf("Failed to unmarshal single string: %v", err)
	}
	if len(s) != 1 || s[0] != "foo" {
		t.Errorf("Expected [foo], got %v", s)
	}

	// Test array
	if err := json.Unmarshal([]byte(`["foo", "bar"]`), &s); err != nil {
		t.Errorf("Failed to unmarshal array: %v", err)
	}
	if len(s) != 2 || s[1] != "bar" {
		t.Errorf("Expected [foo, bar], got %v", s)
	}
}

type postSubject struct {
	ID int
}

func (p postSubject) SubjectType() string { return "Post" }
func (p postSubject) GetField(f string) any {
	if f == "ID" {
		return p.ID
	}
	return nil
}

func TestLoadFromJSON(t *testing.T) {
	jsonParams := `[
		{
			"action": "read",
			"subject": "Post",
			"conditions": {"ID": 1}
		},
		{
			"action": ["update"],
			"subject": ["Post", "Comment"],
			"inverted": true,
			"reason": "ReadOnly"
		},
		{
			"action": ["delete"],
			"subject": ["Post"]
		},
		{
			"action": ["delete"],
			"subject": ["Post"],
			"inverted": true,
			"conditions": {"ID": 1}
		},
		{
			"action": "create",
			"subject": "Post",
			"conditions": {"owner": "{{ .UserID }}"}
		}
	]`

	opts := LoadOptions{
		Vars: map[string]any{"UserID": 123},
	}

	a, err := LoadFromJSON([]byte(jsonParams), opts)
	if err != nil {
		t.Fatalf("LoadFromJSON failed: %v", err)
	}

	// Test Rule 1
	sub := postSubject{ID: 1}

	readPost := DefineAction[postSubject]("read")
	if !Can(a, readPost, sub) {
		t.Errorf("Rule 1 failed: should allow read Post ID=1")
	}

	// Test Rule 2 (Inverted, Multiple actions/subjects)
	updatePost := DefineAction[postSubject]("update")
	if Can(a, updatePost, sub) {
		t.Errorf("Rule 2 failed: should forbid update Post")
	}

	sub2 := postSubject{ID: 2}
	deletePost := DefineAction[postSubject]("delete")
	if !Can(a, deletePost, sub2) {
		t.Errorf("Rule 3 failed: should be able to delete all Post except ID 1")
	}

	if Can(a, deletePost, sub) {
		t.Errorf("failed: should forbid delete Post with ID 1")
	}
}

type ownedSubject struct {
	Owner int
}

func (o ownedSubject) SubjectType() string { return "Post" }
func (o ownedSubject) GetField(f string) any {
	if f == "owner" {
		return o.Owner
	}
	return nil
}

func TestTemplateVarJSON(t *testing.T) {
	jsonParams := `[
		{
			"action": "create",
			"subject": "Post",
			"conditions": {"owner": "{{ .UserID }}"}
		}
	]`
	opts := LoadOptions{
		Vars: map[string]any{"UserID": 123},
	}
	a, _ := LoadFromJSON([]byte(jsonParams), opts)

	create := DefineAction[ownedSubject]("create")

	s1 := ownedSubject{Owner: 123}
	if !Can(a, create, s1) {
		t.Errorf("Should allow create for owner 123")
	}

	s2 := ownedSubject{Owner: 456}
	if Can(a, create, s2) {
		t.Errorf("Should deny create for owner 456")
	}
}

func TestProcessConditions(t *testing.T) {
	input := Cond{
		"a": "{{ .Var }}",
		"b": map[string]any{
			"$eq": "{{ .Var }}",
		},
		"c": []any{"{{ .Var }}", "const"},
	}

	output := processConditions(input)

	if output["a"] != Var("Var") {
		t.Errorf("Top level var failed")
	}

	bMap := output["b"].(map[string]any)
	if bMap["$eq"] != Var("Var") {
		t.Errorf("Nested map var failed")
	}

	cList := output["c"].([]any)
	if cList[0] != Var("Var") {
		t.Errorf("List var failed")
	}
	if cList[1] != "const" {
		t.Errorf("List const failed")
	}
}
