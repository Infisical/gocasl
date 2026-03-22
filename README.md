# Go CASL

Go CASL is an isomorphic authorization library for Golang, inspired by [CASL.js](https://casl.js.org/). It provides a type-safe, flexible, and performant way to manage permissions in your Go applications.

## Features

- **Type-Safe**: Leverages Go generics to ensure compile-time safety for actions and subjects.
- **Isomorphic**: Supports loading rules from JSON (compatible with CASL.js) or defining them in Go.
- **Rich Query Language**: Supports MongoDB-like operators (`$eq`, `$gt`, `$in`, `$regex`, etc.).
- **High Performance**: O(1) rule lookup and lazy condition compilation.
- **Introspection**: Inspect allowed/forbidden fields and reasons for denial.

## Installation

```bash
go get github.com/infisical/gocasl
```

## Quick Start

```go
package main

import (
	"fmt"
	"github.com/infisical/gocasl"
)

// 1. Define your Subject
type Post struct {
	ID    int
	Owner int
}

func (p Post) SubjectType() string { return "Post" }
func (p Post) GetField(f string) any {
	if f == "ID" { return p.ID }
	if f == "Owner" { return p.Owner }
	return nil
}

func main() {
	// 2. Define Actions
	read := gocasl.DefineAction[Post]("read")
	update := gocasl.DefineAction[Post]("update")

	// 3. Define Rules
	builder := gocasl.NewAbility()
	
	// Allow reading all posts
	gocasl.AddRule(builder, gocasl.Allow(read).Build())
	
	// Allow updating own posts
	gocasl.AddRule(builder, gocasl.Allow(update).Where(gocasl.Cond{"Owner": 123}).Build())

	ability := builder.Build()

	// 4. Check Permissions
	post := Post{ID: 1, Owner: 123}
	otherPost := Post{ID: 2, Owner: 456}

	fmt.Println(gocasl.Can(ability, read, post))       // true
	fmt.Println(gocasl.Can(ability, update, post))     // true
	fmt.Println(gocasl.Can(ability, update, otherPost)) // false
}
```

## Documentation

### Core Concepts

- **Subject**: The resource you are protecting. Must implement `SubjectType()` and `GetField()`.
- **Action**: The operation being performed (`read`, `create`, `update`, etc.). Typed via `ActionFor[S]`.
- **Rule**: Defines a permission (Allow or Forbid).
- **Ability**: A collection of rules that can be checked against.

### Defining Rules

Use the `AbilityBuilder` to define rules fluently:

```go
gocasl.AddRule(builder, gocasl.Allow(read).Where(gocasl.Cond{
    "status": "published",
    "viewCount": gocasl.Op{"$gt": 10},
}).Build())
```

### JSON Loading

Load rules defined in JSON (e.g., from a database or frontend):

```go
jsonRules := `[
  {"action": "read", "subject": "Post", "conditions": {"owner": "{{ .userId }}"}}
]`

opts := gocasl.LoadOptions{
    Vars: map[string]any{"userId": 123},
}

ability, _ := gocasl.LoadFromJSON([]byte(jsonRules), opts)
```

### Supported Operators

- `$eq`, `$ne`
- `$gt`, `$gte`, `$lt`, `$lte`
- `$in`, `$nin`
- `$regex`, `$contains` (string/array)
- `$exists`
- `$size` (array/string)

### Custom Operators

You can extend the library with your own operators.

```go
// 1. Define the operator function
func opBetween(fieldValue, operand any) bool {
    // ... implementation ...
    return val >= min && val <= max
}

// 2. Register it when creating the Ability
ops := gocasl.DefaultOperators.With("$between", opBetween)
builder := gocasl.NewAbility().WithOperators(ops)

// 3. Use it in rules
gocasl.AddRule(builder, gocasl.Allow(read).Where(gocasl.Cond{
    "Price": gocasl.Op{"$between": []any{10, 100}},
}).Build())
```

See [examples/custom_operator](examples/custom_operator/main.go) for a full working example.

## Introspection

You can inspect the rules to debug or build UI based on permissions.

```go
// Get reason for denial
reason := gocasl.WhyNot(ability, update, post)

// Get all fields allowed for an action
fields := gocasl.AllowedFields(ability, update, post)

// Get all rules registered for a subject type (ignoring fields/conditions)
// Useful for debugging or pre-filtering
rules := gocasl.PossibleRulesFor(ability, read) 
```


Go CASL is designed for speed.

- **Simple Checks**: ~45ns
- **Complex Checks**: ~3µs (100 rules)
- **Indexing**: O(1) lookup by Subject/Action.
- **Lazy Compilation**: Conditions are compiled only when needed.

## License

MIT
