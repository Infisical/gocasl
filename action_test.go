package gocasl

import "testing"

type Post struct {
	ID    int
	Title string
}

func (p Post) SubjectType() string {
	return "Post"
}

func (p Post) GetField(field string) any {
	if field == "ID" {
		return p.ID
	}
	return nil
}

type Comment struct {
	ID     int
	PostID int
}

func (c Comment) SubjectType() string {
	return "Comment"
}

func (c Comment) GetField(field string) any {
	if field == "ID" {
		return c.ID
	}
	return nil
}

func TestDefineAction(t *testing.T) {
	// Define actions for Post
	readPost := DefineAction[Post]("read")
	updatePost := DefineAction[Post]("update")

	if readPost.Name() != "read" {
		t.Errorf("Expected action name 'read', got %s", readPost.Name())
	}
	if updatePost.Name() != "update" {
		t.Errorf("Expected action name 'update', got %s", updatePost.Name())
	}

	// Define actions for Comment
	readComment := DefineAction[Comment]("read")
	
	if readComment.Name() != "read" {
		t.Errorf("Expected action name 'read', got %s", readComment.Name())
	}

	// Verify type safety (compile-time check mainly, but runtime verification here)
	// We can't really test compile-time errors in runtime tests easily, 
	// but the fact that this compiles is the test.
	
	// Ensure names are distinct even if string is same but type is different
	// (Though in runtime they are just structs with strings, the generics enforce usage)
}
