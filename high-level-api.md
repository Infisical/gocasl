# Go CASL: Complete API Specification v2

A type-safe, high-performance authorization library for Go inspired by [CASL](https://casl.js.org/).

---

## Table of Contents

1. [Design Principles](#design-principles)
2. [Core Types](#core-types)
3. [Subject & Action Definition](#subject--action-definition)
4. [Conditions System](#conditions-system)
5. [Operators](#operators)
6. [Rule Definition](#rule-definition)
7. [Ability Builder](#ability-builder)
8. [Permission Checking](#permission-checking)
9. [JSON Loading](#json-loading)
10. [Performance & Indexing](#performance--indexing)
11. [Complete Example](#complete-example)

---

## Design Principles

| Principle           | Description                                          |
| ------------------- | ---------------------------------------------------- |
| **Type Safety**     | Actions are bound to subjects at compile time        |
| **No Global State** | All dependencies (operators, vars) passed explicitly |
| **Unified DSL**     | Same condition syntax in Go and JSON                 |
| **Performance**     | O(1) lookups via indexing, lazy compilation, caching |
| **Extensible**      | Custom operators, built-in field access              |
| **Immutable**       | Builders return new instances                        |

---

## Core Types

### Subject Interface

```go
// Subject identifies a resource type for authorization
// and provides field access for condition evaluation
type Subject interface {
    // SubjectType returns the resource type identifier
    SubjectType() string

    // GetField returns the value of a named field
    // Returns nil if field doesn't exist
    GetField(field string) any
}
```

**Key Change**: Field access is now part of the Subject interface, eliminating the need for separate registration.

### Action Type

```go
// ActionFor is a type-safe action bound to a subject type
type ActionFor[S Subject] struct {
    name string
}

// Name returns the action name
func (a ActionFor[S]) Name() string
```

### Condition Type

```go
// Cond represents conditions in MongoDB/CASL query format
// Used identically in Go DSL and JSON
type Cond map[string]any
```

### Operator Types

```go
// Op represents an operator expression like { "$eq": value }
type Op map[string]any

// OperatorFunc evaluates an operator against a field value
type OperatorFunc func(fieldValue any, operand any) bool

// Operators holds registered operator functions
type Operators map[string]OperatorFunc
```

### Variable Reference

```go
// VarRef references a template variable resolved at runtime
type VarRef string
```

---

## Subject & Action Definition

### Defining a Subject

```go
type Article struct {
    ID        string
    AuthorID  string
    OrgID     string
    Status    string
    Published bool
    Priority  int
    Tags      []string
}

// Implement Subject interface - SubjectType
func (Article) SubjectType() string {
    return "Article"
}

// Implement Subject interface - GetField
// This method enables automatic field access in conditions
func (a Article) GetField(field string) any {
    switch field {
    case "id":        return a.ID
    case "authorID":  return a.AuthorID
    case "orgID":     return a.OrgID
    case "status":    return a.Status
    case "published": return a.Published
    case "priority":  return a.Priority
    case "tags":      return a.Tags
    default:          return nil
    }
}
```

**Note**: Both `SubjectType()` and `GetField()` are required to satisfy the Subject interface. No separate field accessor registration is needed.

### Pointer Receivers

If your subjects are typically used as pointers, use pointer receivers:

```go
func (a *Article) GetField(field string) any {
    switch field {
    case "id":        return a.ID
    case "authorID":  return a.AuthorID
    // ...
    }
}
```

### Defining Actions

```go
// DefineAction creates a type-safe action for a subject
func DefineAction[S Subject](name string) ActionFor[S]
```

```go
// Define actions for Article (compile-time bound)
var (
    CreateArticle  = casl.DefineAction[Article]("create")
    ReadArticle    = casl.DefineAction[Article]("read")
    UpdateArticle  = casl.DefineAction[Article]("update")
    DeleteArticle  = casl.DefineAction[Article]("delete")
    PublishArticle = casl.DefineAction[Article]("publish")
)

// Define actions for Comment (separate namespace)
var (
    CreateComment  = casl.DefineAction[Comment]("create")
    ReadComment    = casl.DefineAction[Comment]("read")
    UpdateComment  = casl.DefineAction[Comment]("update")
    DeleteComment  = casl.DefineAction[Comment]("delete")
    ApproveComment = casl.DefineAction[Comment]("approve")
)
```

### Compile-Time Safety

```go
// ✅ Compiles - matching types
casl.Can(ability, ReadArticle, article)
casl.Can(ability, ApproveComment, comment)

// ❌ Compile error - type mismatch
casl.Can(ability, ReadArticle, comment)    // ActionFor[Article] vs Comment
casl.Can(ability, ApproveComment, article) // ActionFor[Comment] vs Article
```

---

## Conditions System

### Condition Format (MongoDB/CASL Style)

Conditions use the same format in both Go and JSON:

```go
// Go
casl.Cond{
    "authorID":  "user-123",
    "status":    casl.In("draft", "review"),
    "priority":  casl.Gte(5),
    "published": true,
}

// Equivalent JSON
{
    "authorID": "user-123",
    "status": { "$in": ["draft", "review"] },
    "priority": { "$gte": 5 },
    "published": true
}
```

The library automatically uses `subject.GetField(fieldName)` to retrieve field values for condition evaluation.

### Implicit AND

Multiple fields at the same level are implicitly AND'd:

```go
// Both conditions must match
casl.Cond{
    "authorID": "user-123",
    "published": true,
}
```

### Logical Operators

```go
// AND - explicit
func And(conds ...Cond) Cond

// OR
func Or(conds ...Cond) Cond

// NOT
func Not(cond Cond) Cond
```

```go
// Go
casl.Or(
    casl.Cond{"published": true},
    casl.Cond{"authorID": casl.Var("UserID")},
)

// JSON
{
    "$or": [
        { "published": true },
        { "authorID": "{{ .UserID }}" }
    ]
}
```

### Template Variables

```go
// Go - use Var()
func Var(name string) VarRef

casl.Cond{
    "authorID": casl.Var("UserID"),
    "orgID":    casl.Var("OrgID"),
}

// JSON - use {{ .Name }} syntax
{
    "authorID": "{{ .UserID }}",
    "orgID": "{{ .OrgID }}"
}
```

Variables are resolved when building the ability via `WithVars()`.

---

## Operators

### Built-in Operators

| Operator         | Go DSL           | JSON                 | Description                       |
| ---------------- | ---------------- | -------------------- | --------------------------------- |
| Equals           | `Eq(v)`          | `{"$eq": v}`         | Equality (implicit if bare value) |
| Not Equals       | `Ne(v)`          | `{"$ne": v}`         | Inequality                        |
| Greater Than     | `Gt(v)`          | `{"$gt": v}`         | >                                 |
| Greater or Equal | `Gte(v)`         | `{"$gte": v}`        | >=                                |
| Less Than        | `Lt(v)`          | `{"$lt": v}`         | <                                 |
| Less or Equal    | `Lte(v)`         | `{"$lte": v}`        | <=                                |
| In Array         | `In(a, b, ...)`  | `{"$in": [a, b]}`    | Value in array                    |
| Not In Array     | `Nin(a, b, ...)` | `{"$nin": [a, b]}`   | Value not in array                |
| Regex            | `Regex(p)`       | `{"$regex": p}`      | Regex match                       |
| Contains         | `Contains(s)`    | `{"$contains": s}`   | String contains                   |
| Starts With      | `StartsWith(s)`  | `{"$startsWith": s}` | String prefix                     |
| Ends With        | `EndsWith(s)`    | `{"$endsWith": s}`   | String suffix                     |
| Exists           | `Exists(bool)`   | `{"$exists": bool}`  | Field existence                   |
| Size             | `Size(n)`        | `{"$size": n}`       | Array length                      |

### Operator Functions

```go
func Eq(value any) Op
func Ne(value any) Op
func Gt(value any) Op
func Gte(value any) Op
func Lt(value any) Op
func Lte(value any) Op
func In(values ...any) Op
func Nin(values ...any) Op
func Regex(pattern string) Op
func Contains(substr string) Op
func StartsWith(prefix string) Op
func EndsWith(suffix string) Op
func Exists(exists bool) Op
func Size(n int) Op

// Combine multiple operators on same field
func All(ops ...Op) Op
```

### Default Operators Set

```go
var DefaultOperators = Operators{
    "$eq":         opEq,
    "$ne":         opNe,
    "$gt":         opGt,
    "$gte":        opGte,
    "$lt":         opLt,
    "$lte":        opLte,
    "$in":         opIn,
    "$nin":        opNin,
    "$regex":      opRegex,
    "$exists":     opExists,
    "$contains":   opContains,
    "$startsWith": opStartsWith,
    "$endsWith":   opEndsWith,
    "$size":       opSize,
}
```

### Operators Methods

```go
// Clone creates a copy of the operators map
func (o Operators) Clone() Operators

// With adds/overrides an operator, returns new map
func (o Operators) With(name string, fn OperatorFunc) Operators

// WithAll merges another operators map, returns new map
func (o Operators) WithAll(ops Operators) Operators

// Without removes operators, returns new map
func (o Operators) Without(names ...string) Operators
```

### Custom Operators

```go
// 1. Define operator function
func opBetween(fieldValue, operand any) bool {
    bounds := operand.([]any)
    min, max := bounds[0], bounds[1]
    return compare(fieldValue, min) >= 0 &&
           compare(fieldValue, max) <= 0
}

// 2. Define DSL helper
func Between(min, max any) Op {
    return Op{"$between": []any{min, max}}
}

// 3. Register when building ability
ability := casl.NewAbility().
    WithOperators(casl.DefaultOperators.With("$between", opBetween)).
    Build()

// 4. Use in conditions
casl.Cond{"priority": Between(1, 10)}

// JSON equivalent
{"priority": {"$between": [1, 10]}}
```

---

## Rule Definition

### Rule Structure

```go
type Rule[S Subject] struct {
    Action   ActionFor[S]  // The action this rule applies to
    Inverted bool          // true = forbid, false = allow
    Cond     Cond          // Conditions (nil = unconditional)
    Fields   []string      // Field restrictions (nil = all fields)
    Reason   string        // Explanation for forbid rules
}
```

### Rule Builder

```go
// Allow creates an allow rule builder
func Allow[S Subject](action ActionFor[S]) *RuleBuilder[S]

// Forbid creates a forbid rule builder
func Forbid[S Subject](action ActionFor[S]) *RuleBuilder[S]
```

### RuleBuilder Methods

```go
type RuleBuilder[S Subject] struct { ... }

// Where sets the conditions for the rule
func (rb *RuleBuilder[S]) Where(cond Cond) *RuleBuilder[S]

// OnFields restricts the rule to specific fields
func (rb *RuleBuilder[S]) OnFields(fields ...string) *RuleBuilder[S]

// Because sets the reason (useful for forbid rules)
func (rb *RuleBuilder[S]) Because(reason string) *RuleBuilder[S]

// Build returns the constructed rule
func (rb *RuleBuilder[S]) Build() Rule[S]
```

### Rule Examples

```go
// Unconditional allow
casl.Allow(ReadArticle).Build()

// Allow with conditions
casl.Allow(ReadArticle).
    Where(casl.Cond{"published": true}).
    Build()

// Allow with field restrictions
casl.Allow(UpdateArticle).
    Where(casl.Cond{"authorID": casl.Var("UserID")}).
    OnFields("title", "body").
    Build()

// Forbid with reason
casl.Forbid(DeleteArticle).
    Where(casl.Cond{"published": true}).
    Because("cannot delete published articles").
    Build()

// Complex conditions
casl.Allow(PublishArticle).
    Where(casl.Cond{
        "authorID": casl.Var("UserID"),
        "status":   "review",
        "priority": casl.Gte(5),
    }).
    Build()
```

---

## Ability Builder

### AbilityBuilder

```go
type AbilityBuilder struct { ... }

// NewAbility creates a new builder with default operators
func NewAbility() *AbilityBuilder
```

### AbilityBuilder Methods

```go
// WithOperators sets the operators (replaces default)
func (ab *AbilityBuilder) WithOperators(ops Operators) *AbilityBuilder

// WithVars sets template variables for condition resolution
func (ab *AbilityBuilder) WithVars(vars map[string]any) *AbilityBuilder

// Build creates the final Ability
func (ab *AbilityBuilder) Build() *Ability
```

### Adding Rules

```go
// AddRule adds a typed rule to the builder
func AddRule[S Subject](ab *AbilityBuilder, rule Rule[S]) *AbilityBuilder

// AddRules adds multiple rules of the same subject type
func AddRules[S Subject](ab *AbilityBuilder, rules ...Rule[S]) *AbilityBuilder
```

### Builder Example

```go
ability := casl.NewAbility().
    // Set operators (optional - defaults to DefaultOperators)
    WithOperators(casl.DefaultOperators.
        With("$between", opBetween)).

    // Set template variables
    WithVars(map[string]any{
        "UserID": currentUser.ID,
        "OrgID":  currentUser.OrgID,
        "Role":   currentUser.Role,
    }).

    // Build
    Build()

// Add rules (can be before or after other config)
casl.AddRule(ab, casl.Allow(ReadArticle).
    Where(casl.Cond{"published": true}).
    Build())

casl.AddRule(ab, casl.Allow(UpdateArticle).
    Where(casl.Cond{
        "authorID": casl.Var("UserID"),
        "status":   casl.In("draft", "review"),
    }).
    OnFields("title", "body").
    Build())

casl.AddRule(ab, casl.Forbid(DeleteArticle).
    Where(casl.Cond{"published": true}).
    Because("cannot delete published articles").
    Build())

ability := ab.Build()
```

**Note**: No `RegisterFieldAccessor` calls needed! Field access is automatic through the `Subject.GetField()` interface method.

---

## Permission Checking

### Ability

```go
type Ability struct { ... }
```

### Check Functions

```go
// Can checks if action is allowed on subject
func Can[S Subject](a *Ability, action ActionFor[S], subject S) bool

// Cannot checks if action is forbidden on subject
func Cannot[S Subject](a *Ability, action ActionFor[S], subject S) bool

// CanWithField checks if action is allowed on a specific field
func CanWithField[S Subject](a *Ability, action ActionFor[S], subject S, field string) bool

// CanAll checks if all actions are allowed
func CanAll[S Subject](a *Ability, subject S, actions ...ActionFor[S]) bool

// CanAny checks if any action is allowed
func CanAny[S Subject](a *Ability, subject S, actions ...ActionFor[S]) bool
```

### Introspection Functions

```go
// WhyNot returns the reason an action is forbidden
func WhyNot[S Subject](a *Ability, action ActionFor[S], subject S) string

// RulesFor returns all matching rules for an action/subject
func RulesFor[S Subject](a *Ability, action ActionFor[S], subject S) []RuleInfo

// AllowedFields returns fields the user can access for an action
func AllowedFields[S Subject](a *Ability, action ActionFor[S], subject S) []string

// ForbiddenFields returns fields the user cannot access
func ForbiddenFields[S Subject](a *Ability, action ActionFor[S], subject S) []string
```

### RuleInfo

```go
type RuleInfo struct {
    Action     string
    Subject    string
    Inverted   bool
    Fields     []string
    Reason     string
    Matched    bool  // Did conditions match?
}
```

### Check Examples

```go
article := Article{
    ID:        "1",
    AuthorID:  "user-123",
    Status:    "draft",
    Published: false,
}

// Basic checks
if casl.Can(ability, ReadArticle, article) {
    // Allow read
}

if casl.Cannot(ability, DeleteArticle, article) {
    reason := casl.WhyNot(ability, DeleteArticle, article)
    fmt.Println("Denied:", reason)
}

// Field-level check
if casl.CanWithField(ability, UpdateArticle, article, "title") {
    // Allow title update
}

// Get allowed fields
fields := casl.AllowedFields(ability, UpdateArticle, article)
// ["title", "body"]

// Check multiple actions
if casl.CanAll(ability, article, ReadArticle, UpdateArticle) {
    // Can both read and update
}
```

---

## JSON Loading

### JSON Rule Format

```json
{
  "rules": [
    {
      "action": "read",
      "subject": "Article",
      "conditions": {
        "published": true
      }
    },
    {
      "action": "update",
      "subject": "Article",
      "conditions": {
        "authorID": "{{ .UserID }}",
        "status": { "$in": ["draft", "review"] }
      },
      "fields": ["title", "body"]
    },
    {
      "action": "delete",
      "subject": "Article",
      "inverted": true,
      "conditions": {
        "published": true
      },
      "reason": "cannot delete published articles"
    },
    {
      "action": ["read", "update"],
      "subject": "Article",
      "conditions": {
        "orgID": "{{ .OrgID }}"
      }
    }
  ]
}
```

### JSON Types

```go
type JSONRule struct {
    Action     StringOrSlice `json:"action"`
    Subject    string        `json:"subject"`
    Inverted   bool          `json:"inverted,omitempty"`
    Conditions Cond          `json:"conditions,omitempty"`
    Fields     []string      `json:"fields,omitempty"`
    Reason     string        `json:"reason,omitempty"`
}

type JSONRuleSet struct {
    Rules []JSONRule `json:"rules"`
}

// StringOrSlice handles action being string or []string
type StringOrSlice []string
```

### Load Options

```go
type LoadOptions struct {
    // Operators to use (defaults to DefaultOperators)
    Operators Operators

    // Template variables
    Vars map[string]any

    // Subject type registry for dynamic subject creation
    // Maps subject type names to factory functions
    SubjectRegistry map[string]func() Subject

    // Action validators per subject type (optional)
    ActionValidators map[string]func(action string) bool
}
```

**Key Change**: Instead of `FieldAccessors map[string]FieldAccessorAny`, we use `SubjectRegistry` to create subject instances. Field access is automatic through the `GetField()` interface method.

### Load Functions

```go
// LoadFromJSON parses JSON and creates an Ability
func LoadFromJSON(data []byte, opts LoadOptions) (*Ability, error)

// LoadFromFile loads rules from a JSON file
func LoadFromFile(path string, opts LoadOptions) (*Ability, error)

// LoadFromReader loads rules from an io.Reader
func LoadFromReader(r io.Reader, opts LoadOptions) (*Ability, error)
```

### Load Example

```go
jsonData := `{
    "rules": [
        {
            "action": "read",
            "subject": "Article",
            "conditions": { "published": true }
        },
        {
            "action": "update",
            "subject": "Article",
            "conditions": {
                "authorID": "{{ .UserID }}",
                "status": { "$in": ["draft", "review"] }
            }
        }
    ]
}`

ability, err := casl.LoadFromJSON([]byte(jsonData), casl.LoadOptions{
    Operators: casl.DefaultOperators.With("$between", opBetween),
    Vars: map[string]any{
        "UserID": currentUser.ID,
        "OrgID":  currentUser.OrgID,
    },
    SubjectRegistry: map[string]func() casl.Subject{
        "Article": func() casl.Subject { return &Article{} },
        "Comment": func() casl.Subject { return &Comment{} },
    },
})
```

**Note**: The subject registry provides factory functions to create empty instances. The library uses `subject.GetField()` automatically for field access during condition evaluation.

---

## Performance & Indexing

### Indexing Strategy (Like CASL)

```go
// Internal index structure
type ruleIndex struct {
    // Primary index: subjectType -> action -> rules
    bySubjectAction map[string]map[string][]indexedRule

    // Track if any rules have field restrictions
    hasPerFieldRules bool

    // Cached merged rule lists
    mergedCache map[cacheKey]*mergedRules
}
```

### Performance Characteristics

| Operation             | Complexity | Notes                               |
| --------------------- | ---------- | ----------------------------------- |
| Can (no conditions)   | O(1)       | Map lookup + cached array           |
| Can (with conditions) | O(n)       | n = matching rules, not total rules |
| RulesFor              | O(1)       | Cached merged list                  |
| Build index           | O(n)       | One-time at ability creation        |
| GetField              | O(1)       | Direct method call on subject       |

### Optimizations

1. **Rule Indexing**: Rules indexed by `subject type → action` at build time
2. **Merged Caching**: Merged rule lists cached after first access
3. **Lazy Compilation**: Conditions compiled on first evaluation
4. **Field Short-circuit**: Skip field filtering if no field rules exist
5. **Subject Detection**: Optimized based on registered subject types
6. **Direct Field Access**: No lookup overhead - `GetField()` is a direct method call

### Lazy Condition Compilation

```go
type compiledRule struct {
    rule        Rule
    compiled    Condition  // Lazy compiled
    compileOnce sync.Once
}

func (cr *compiledRule) matches(subject any, compiler *condCompiler) bool {
    cr.compileOnce.Do(func() {
        cr.compiled = compiler.compile(cr.rule.Cond)
    })
    return cr.compiled(subject)
}
```

### Field Access Performance

Field access via `GetField()` is highly efficient:

- No registration lookup needed
- Direct method dispatch
- Compiler can inline simple cases
- Zero allocation for most field types

---

## Complete Example

```go
package main

import (
    "fmt"
    "github.com/yourorg/casl"
)

// ════════════════════════════════════════════════════════════════════════════
// DOMAIN TYPES
// ════════════════════════════════════════════════════════════════════════════

type Article struct {
    ID        string
    AuthorID  string
    OrgID     string
    Status    string
    Published bool
    Priority  int
    Tags      []string
}

// Implement Subject interface
func (Article) SubjectType() string {
    return "Article"
}

func (a Article) GetField(field string) any {
    switch field {
    case "id":        return a.ID
    case "authorID":  return a.AuthorID
    case "orgID":     return a.OrgID
    case "status":    return a.Status
    case "published": return a.Published
    case "priority":  return a.Priority
    case "tags":      return a.Tags
    default:          return nil
    }
}

type Comment struct {
    ID        string
    ArticleID string
    AuthorID  string
    Content   string
    Approved  bool
}

// Implement Subject interface
func (Comment) SubjectType() string {
    return "Comment"
}

func (c Comment) GetField(field string) any {
    switch field {
    case "id":        return c.ID
    case "articleID": return c.ArticleID
    case "authorID":  return c.AuthorID
    case "content":   return c.Content
    case "approved":  return c.Approved
    default:          return nil
    }
}

// ════════════════════════════════════════════════════════════════════════════
// ACTIONS
// ════════════════════════════════════════════════════════════════════════════

var (
    CreateArticle  = casl.DefineAction[Article]("create")
    ReadArticle    = casl.DefineAction[Article]("read")
    UpdateArticle  = casl.DefineAction[Article]("update")
    DeleteArticle  = casl.DefineAction[Article]("delete")
    PublishArticle = casl.DefineAction[Article]("publish")
)

var (
    CreateComment  = casl.DefineAction[Comment]("create")
    ReadComment    = casl.DefineAction[Comment]("read")
    UpdateComment  = casl.DefineAction[Comment]("update")
    DeleteComment  = casl.DefineAction[Comment]("delete")
    ApproveComment = casl.DefineAction[Comment]("approve")
)

// ════════════════════════════════════════════════════════════════════════════
// CUSTOM OPERATORS
// ════════════════════════════════════════════════════════════════════════════

func opBetween(fieldValue, operand any) bool {
    bounds := operand.([]any)
    min, max := bounds[0], bounds[1]
    return casl.Compare(fieldValue, min) >= 0 &&
           casl.Compare(fieldValue, max) <= 0
}

func Between(min, max any) casl.Op {
    return casl.Op{"$between": []any{min, max}}
}

func opHasTag(fieldValue, operand any) bool {
    tags, ok := fieldValue.([]string)
    if !ok {
        return false
    }
    tag := operand.(string)
    for _, t := range tags {
        if t == tag {
            return true
        }
    }
    return false
}

func HasTag(tag string) casl.Op {
    return casl.Op{"$hasTag": tag}
}

// ════════════════════════════════════════════════════════════════════════════
// BUILD ABILITY
// ════════════════════════════════════════════════════════════════════════════

type User struct {
    ID      string
    OrgID   string
    Role    string
    IsAdmin bool
}

func BuildAbility(user User) *casl.Ability {
    ab := casl.NewAbility().
        WithOperators(casl.DefaultOperators.
            With("$between", opBetween).
            With("$hasTag", opHasTag)).
        WithVars(map[string]any{
            "UserID": user.ID,
            "OrgID":  user.OrgID,
            "Role":   user.Role,
        })

    // No RegisterFieldAccessor calls needed!
    // Field access is automatic via Subject.GetField()

    // ─── Article Rules ──────────────────────────────────────────────────

    // Anyone can read published articles
    casl.AddRule(ab, casl.Allow(ReadArticle).
        Where(casl.Cond{"published": true}).
        Build())

    // Users can read articles in their org
    casl.AddRule(ab, casl.Allow(ReadArticle).
        Where(casl.Cond{"orgID": casl.Var("OrgID")}).
        Build())

    // Users can create articles
    casl.AddRule(ab, casl.Allow(CreateArticle).Build())

    // Users can update their own draft/review articles
    casl.AddRule(ab, casl.Allow(UpdateArticle).
        Where(casl.Cond{
            "authorID": casl.Var("UserID"),
            "status":   casl.In("draft", "review"),
        }).
        OnFields("title", "body", "tags").
        Build())

    // Users can delete their own unpublished articles
    casl.AddRule(ab, casl.Allow(DeleteArticle).
        Where(casl.Cond{
            "authorID":  casl.Var("UserID"),
            "published": false,
        }).
        Build())

    // Cannot delete published articles (overrides above)
    casl.AddRule(ab, casl.Forbid(DeleteArticle).
        Where(casl.Cond{"published": true}).
        Because("cannot delete published articles").
        Build())

    // Users can publish their own reviewed high-priority articles
    casl.AddRule(ab, casl.Allow(PublishArticle).
        Where(casl.Cond{
            "authorID": casl.Var("UserID"),
            "status":   "review",
            "priority": Between(5, 10),
        }).
        Build())

    // ─── Comment Rules ──────────────────────────────────────────────────

    // Anyone can read approved comments
    casl.AddRule(ab, casl.Allow(ReadComment).
        Where(casl.Cond{"approved": true}).
        Build())

    // Users can create comments
    casl.AddRule(ab, casl.Allow(CreateComment).Build())

    // Users can update their own comments
    casl.AddRule(ab, casl.Allow(UpdateComment).
        Where(casl.Cond{"authorID": casl.Var("UserID")}).
        OnFields("content").
        Build())

    // Users can delete their own comments
    casl.AddRule(ab, casl.Allow(DeleteComment).
        Where(casl.Cond{"authorID": casl.Var("UserID")}).
        Build())

    // ─── Admin Rules ────────────────────────────────────────────────────

    if user.IsAdmin {
        casl.AddRule(ab, casl.Allow(PublishArticle).Build())
        casl.AddRule(ab, casl.Allow(DeleteArticle).Build())
        casl.AddRule(ab, casl.Allow(ApproveComment).Build())
    }

    return ab.Build()
}

// ════════════════════════════════════════════════════════════════════════════
// USAGE
// ════════════════════════════════════════════════════════════════════════════

func main() {
    user := User{
        ID:      "user-123",
        OrgID:   "org-456",
        Role:    "editor",
        IsAdmin: false,
    }

    ability := BuildAbility(user)

    // Test articles
    ownDraft := Article{
        ID:        "1",
        AuthorID:  "user-123",
        OrgID:     "org-456",
        Status:    "draft",
        Published: false,
        Priority:  3,
    }

    ownReviewed := Article{
        ID:        "2",
        AuthorID:  "user-123",
        OrgID:     "org-456",
        Status:    "review",
        Published: false,
        Priority:  7,
    }

    otherPublished := Article{
        ID:        "3",
        AuthorID:  "user-other",
        OrgID:     "org-other",
        Status:    "published",
        Published: true,
        Priority:  5,
    }

    fmt.Println("=== Own Draft ===")
    fmt.Printf("Read:    %v\n", casl.Can(ability, ReadArticle, ownDraft))     // true (same org)
    fmt.Printf("Update:  %v\n", casl.Can(ability, UpdateArticle, ownDraft))   // true
    fmt.Printf("Delete:  %v\n", casl.Can(ability, DeleteArticle, ownDraft))   // true
    fmt.Printf("Publish: %v\n", casl.Can(ability, PublishArticle, ownDraft))  // false (not review, low priority)

    fmt.Println("\n=== Own Reviewed ===")
    fmt.Printf("Publish: %v\n", casl.Can(ability, PublishArticle, ownReviewed)) // true

    fmt.Println("\n=== Other's Published ===")
    fmt.Printf("Read:    %v\n", casl.Can(ability, ReadArticle, otherPublished))   // true (published)
    fmt.Printf("Update:  %v\n", casl.Can(ability, UpdateArticle, otherPublished)) // false
    fmt.Printf("Delete:  %v\n", casl.Can(ability, DeleteArticle, otherPublished)) // false
    fmt.Printf("Reason:  %s\n", casl.WhyNot(ability, DeleteArticle, otherPublished))

    fmt.Println("\n=== Field Access ===")
    fmt.Printf("Can update title: %v\n", casl.CanWithField(ability, UpdateArticle, ownDraft, "title")) // true
    fmt.Printf("Can update status: %v\n", casl.CanWithField(ability, UpdateArticle, ownDraft, "status")) // false
    fmt.Printf("Allowed fields: %v\n", casl.AllowedFields(ability, UpdateArticle, ownDraft)) // [title, body, tags]

    // ─── Type Safety Demo ───────────────────────────────────────────────

    comment := Comment{ID: "1", AuthorID: "user-123"}

    // ✅ These compile
    casl.Can(ability, ReadArticle, ownDraft)
    casl.Can(ability, ReadComment, comment)

    // ❌ These would NOT compile
    // casl.Can(ability, ReadArticle, comment)      // Type mismatch
    // casl.Can(ability, ApproveComment, ownDraft)  // Type mismatch
}
```

---

## API Summary

### Core Functions

| Function                              | Description                |
| ------------------------------------- | -------------------------- |
| `DefineAction[S](name)`               | Create type-safe action    |
| `NewAbility()`                        | Start building ability     |
| `Allow[S](action)`                    | Create allow rule builder  |
| `Forbid[S](action)`                   | Create forbid rule builder |
| `AddRule[S](ab, rule)`                | Add rule to builder        |
| `Can[S](ability, action, subject)`    | Check permission           |
| `Cannot[S](ability, action, subject)` | Check denial               |
| `CanWithField[S](...)`                | Check field permission     |
| `WhyNot[S](...)`                      | Get denial reason          |
| `LoadFromJSON(data, opts)`            | Load from JSON             |

### Condition Helpers

| Function                                | Description          |
| --------------------------------------- | -------------------- |
| `Cond{}`                                | Create condition map |
| `Var(name)`                             | Variable reference   |
| `And(conds...)`                         | Logical AND          |
| `Or(conds...)`                          | Logical OR           |
| `Not(cond)`                             | Logical NOT          |
| `Eq/Ne/Gt/Gte/Lt/Lte(v)`                | Comparison operators |
| `In/Nin(values...)`                     | Array operators      |
| `Regex/Contains/StartsWith/EndsWith(s)` | String operators     |
| `Exists(bool)` / `Size(n)`              | Other operators      |

### Builder Methods

| Method               | Description            |
| -------------------- | ---------------------- |
| `WithOperators(ops)` | Set operators          |
| `WithVars(vars)`     | Set template variables |
| `Build()`            | Create ability         |

**Removed**: `RegisterFieldAccessor[S](fn)` - No longer needed!

### Rule Builder Methods

| Method                | Description            |
| --------------------- | ---------------------- |
| `Where(cond)`         | Set conditions         |
| `OnFields(fields...)` | Set field restrictions |
| `Because(reason)`     | Set denial reason      |
| `Build()`             | Create rule            |

---

## Migration Guide (v1 → v2)

### What Changed

1. **Subject Interface**: Added `GetField(field string) any` method requirement
2. **No Registration**: Removed `RegisterFieldAccessor()` function
3. **JSON Loading**: Changed `FieldAccessors` to `SubjectRegistry` in `LoadOptions`

### Migration Steps

**Step 1**: Add `GetField()` to your subjects

```go
// Before (v1)
type Article struct { ... }
func (Article) SubjectType() string { return "Article" }
func ArticleFields(a Article, field string) any { ... }

// After (v2)
type Article struct { ... }
func (Article) SubjectType() string { return "Article" }
func (a Article) GetField(field string) any { ... }  // NEW!
```

**Step 2**: Remove field accessor registration

```go
// Before (v1)
ab := casl.NewAbility()...
casl.RegisterFieldAccessor(ab, ArticleFields)
casl.RegisterFieldAccessor(ab, CommentFields)

// After (v2)
ab := casl.NewAbility()...
// No registration needed! ✅
```

**Step 3**: Update JSON loading (if used)

```go
// Before (v1)
ability, err := casl.LoadFromJSON(data, casl.LoadOptions{
    FieldAccessors: map[string]casl.FieldAccessorAny{
        "Article": func(s any, f string) any {
            return ArticleFields(s.(Article), f)
        },
    },
})

// After (v2)
ability, err := casl.LoadFromJSON(data, casl.LoadOptions{
    SubjectRegistry: map[string]func() casl.Subject{
        "Article": func() casl.Subject { return &Article{} },
    },
})
```

### Benefits of v2

- ✅ Less boilerplate code
- ✅ Cleaner API surface
- ✅ Better performance (direct method calls)
- ✅ More intuitive (field access is part of the subject)
- ✅ Easier to understand and maintain
