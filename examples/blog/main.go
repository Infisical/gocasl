package main

import (
	"fmt"
	"github.com/akhilmhdh/gocasl"
	"log"
)

type Article struct {
	ID        int
	AuthorID  int
	Published bool
}

var (
	ArticleReadAction   = gocasl.DefineAction[Article]("read")
	ArticleUpdateAction = gocasl.DefineAction[Article]("update")
)

func (a Article) SubjectType() string { return "Article" }
func (a Article) GetField(f string) any {
	if f == "ID" {
		return a.ID
	}
	if f == "AuthorID" {
		return a.AuthorID
	}
	if f == "Published" {
		return a.Published
	}
	return nil
}

type User struct {
	ID int
}

func (a User) SubjectType() string { return "User" }
func (a User) GetField(f string) any {
	if f == "ID" {
		return a.ID
	}
	return nil
}

var (
	UserReadAction   = gocasl.DefineAction[User]("read")
	UserUpdateAction = gocasl.DefineAction[User]("update")
)

func main() {

	currentUser := 123

	builder := gocasl.NewAbility()
	builder.WithVars(map[string]any{"user": currentUser})

	// Rule: Can read any published article
	gocasl.AddRule(builder, gocasl.Allow(ArticleReadAction).Where(gocasl.Cond{"Published": true}).Build())

	// Rule: Can read and update own articles
	// We use {{ .user }} template variable
	gocasl.AddRule(builder, gocasl.Allow(ArticleReadAction).Where(gocasl.Cond{"AuthorID": gocasl.Var("user")}).Build())
	gocasl.AddRule(builder, gocasl.Allow(ArticleUpdateAction).Where(gocasl.Cond{"AuthorID": gocasl.Var("user")}).Build())

	ability, err := builder.Build()
	if err != nil {
		log.Fatal(err)
	}

	publicArticle := Article{ID: 1, AuthorID: 999, Published: true}
	privateArticle := Article{ID: 2, AuthorID: 999, Published: false}
	ownArticle := Article{ID: 3, AuthorID: 123, Published: false}

	fmt.Printf("Can read public? %v\n", gocasl.Can(ability, ArticleReadAction, publicArticle))   // true
	fmt.Printf("Can read private? %v\n", gocasl.Can(ability, ArticleReadAction, privateArticle)) // false
	fmt.Printf("Can read own? %v\n", gocasl.Can(ability, ArticleReadAction, ownArticle))         // true
	fmt.Printf("Can update own? %v\n", gocasl.Can(ability, ArticleReadAction, ownArticle))       // true
}
