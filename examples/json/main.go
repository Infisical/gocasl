package main

import (
	"fmt"
	"github.com/akhilmhdh/gocasl"
	"log"
)

type Document struct {
	ID     int
	Owner  string
	Public bool
}

func (d Document) SubjectType() string { return "Document" }
func (d Document) GetField(f string) any {
	if f == "owner" {
		return d.Owner
	}
	if f == "public" {
		return d.Public
	}
	return nil
}

func main() {
	opts := gocasl.LoadOptions{
		Vars: map[string]any{"user": "alice"},
	}

	ability, err := gocasl.LoadFromFile("examples/json/rules.json", opts)
	if err != nil {
		log.Fatalf("Failed to load rules: %v", err)
	}

	read := gocasl.DefineAction[Document]("read")
	deleteOp := gocasl.DefineAction[Document]("delete")

	doc1 := Document{ID: 1, Owner: "bob", Public: true}
	doc2 := Document{ID: 2, Owner: "alice", Public: false}

	fmt.Printf("Alice can read doc1 (public)? %v\n", gocasl.Can(ability, read, doc1))
	fmt.Printf("Alice can delete doc1 (bob's)? %v\n", gocasl.Can(ability, deleteOp, doc1))
	fmt.Printf("Alice can delete doc2 (own)? %v\n", gocasl.Can(ability, deleteOp, doc2))
}
