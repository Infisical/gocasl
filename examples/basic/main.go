package main

import (
	"fmt"
	"github.com/akhilmhdh/gocasl"
	"log"
)

type Post struct {
	ID    int
	Title string
	Owner int
}

func (p Post) SubjectType() string {
	return "Post"
}

func (p Post) GetField(field string) any {
	switch field {
	case "ID":
		return p.ID
	case "Title":
		return p.Title
	case "Owner":
		return p.Owner
	}
	return nil
}

func main() {
	// Define actions
	read := gocasl.DefineAction[Post]("read")
	update := gocasl.DefineAction[Post]("update")

	// Create ability builder
	builder := gocasl.NewAbility()

	// Define rules
	// Allow everyone to read posts
	gocasl.AddRule(builder, gocasl.Allow(read).Build())

	// Allow owner to update posts (Owner ID = 1)
	gocasl.AddRule(builder, gocasl.Allow(update).Where(gocasl.Cond{"Owner": 1}).Build())

	// Build ability
	ability, err := builder.Build()
	if err != nil {
		log.Fatal(err)
	}

	// Check permissions
	post1 := Post{ID: 100, Title: "Hello", Owner: 1}
	post2 := Post{ID: 101, Title: "World", Owner: 2}

	fmt.Printf("Can read post1? %v\n", gocasl.Can(ability, read, post1))     // true
	fmt.Printf("Can update post1? %v\n", gocasl.Can(ability, update, post1)) // true (Owner 1)

	fmt.Printf("Can read post2? %v\n", gocasl.Can(ability, read, post2))     // true
	fmt.Printf("Can update post2? %v\n", gocasl.Can(ability, update, post2)) // false (Owner 2 != 1)
}
