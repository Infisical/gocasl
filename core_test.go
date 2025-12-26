package gocasl

import (
	"reflect"
	"testing"
)

// mockSubject is a simple struct implementing Subject for testing
type mockSubject struct {
	ID    int
	Title string
	Tags  []string
}

func (m mockSubject) SubjectType() string {
	return "MockSubject"
}

func (m mockSubject) GetField(field string) any {
	switch field {
	case "ID":
		return m.ID
	case "Title":
		return m.Title
	case "Tags":
		return m.Tags
	default:
		return nil
	}
}

func TestSubjectInterface(t *testing.T) {
	s := mockSubject{
		ID:    123,
		Title: "Test Subject",
		Tags:  []string{"a", "b"},
	}

	// Verify SubjectType
	if s.SubjectType() != "MockSubject" {
		t.Errorf("Expected SubjectType 'MockSubject', got %s", s.SubjectType())
	}

	// Verify GetField existing
	if id := s.GetField("ID"); id != 123 {
		t.Errorf("Expected ID 123, got %v", id)
	}
	if title := s.GetField("Title"); title != "Test Subject" {
		t.Errorf("Expected Title 'Test Subject', got %v", title)
	}

	// Verify GetField slice
	tags := s.GetField("Tags")
	if !reflect.DeepEqual(tags, []string{"a", "b"}) {
		t.Errorf("Expected Tags ['a', 'b'], got %v", tags)
	}

	// Verify GetField missing
	if val := s.GetField("Missing"); val != nil {
		t.Errorf("Expected nil for missing field, got %v", val)
	}
}
