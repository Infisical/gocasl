# Go CASL Implementation Plan

## Project Overview
Implement a complete Go CASL authorization library from scratch based on the high-level-api.md specification. The library will provide type-safe, high-performance authorization with a unified DSL for Go and JSON.

## Implementation Strategy
- **Bottom-up approach**: Build foundational components first, then compose them
- **Test-driven**: Each phase includes comprehensive tests before moving forward
- **Incremental**: Each phase produces working, testable code
- **Type-safe**: Leverage Go generics for compile-time safety

---

## Phase 1: Project Setup & Core Types

### Objectives
- Initialize Go module
- Define foundational types and interfaces
- Establish project structure

### Tasks
1. **Initialize Go Module**
   - Create `go.mod` with module name `github.com/yourorg/gocasl`
   - Set Go version to 1.21+ (for generics support)

2. **Create Project Structure**
   ```
   gocasl/
   ├── go.mod
   ├── core.go              # Core types and interfaces
   ├── core_test.go         # Core types tests
   ├── action.go            # ActionFor[S] implementation
   ├── action_test.go
   ├── condition.go         # Cond and condition helpers
   ├── condition_test.go
   └── README.md
   ```

3. **Implement Core Types** (`core.go`)
   - `Subject` interface with `SubjectType()` and `GetField(field string) any`
   - `Cond` type as `map[string]any`
   - `Op` type as `map[string]any`
   - `VarRef` type as `string`
   - `OperatorFunc` type signature
   - `Operators` type as `map[string]OperatorFunc`

4. **Implement ActionFor[S]** (`action.go`)
   - Generic `ActionFor[S Subject]` struct with `name` field
   - `DefineAction[S Subject](name string) ActionFor[S]` function
   - `Name() string` method

5. **Implement Condition Helpers** (`condition.go`)
   - `Var(name string) VarRef` - template variable reference
   - Logical operators: `And(conds ...Cond) Cond`, `Or(conds ...Cond) Cond`, `Not(cond Cond) Cond`

### Testing
- **Core Types Tests** (`core_test.go`)
  - Test Subject interface implementation with sample struct
  - Verify GetField() returns correct values and nil for missing fields

- **Action Tests** (`action_test.go`)
  - Test `DefineAction[S]()` creates actions with correct names
  - Test compile-time type safety (different actions for different subjects)
  - Test `Name()` method returns correct action name

- **Condition Helpers Tests** (`condition_test.go`)
  - Test `Var()` creates VarRef
  - Test `And()` creates `{"$and": [...]}`
  - Test `Or()` creates `{"$or": [...]}`
  - Test `Not()` creates `{"$not": {...}}`

### Success Criteria
- ✅ All core types defined and documented
- ✅ Actions are type-safe at compile time
- ✅ All tests passing (>90% coverage)

**Phase 1 Completed** ✅

---

## Phase 2: Operators System

### Objectives
- Implement all built-in operators
- Create operator DSL functions
- Implement custom operator support

### Tasks
1. **Create Operators File** (`operators.go`)
   - Define `DefaultOperators` with all built-in operators
   - Implement helper function for type comparisons

2. **Implement Comparison Operators**
   - `opEq`, `opNe`, `opGt`, `opGte`, `opLt`, `opLte`
   - DSL functions: `Eq()`, `Ne()`, `Gt()`, `Gte()`, `Lt()`, `Lte()`
   - Type-aware comparison (numbers, strings, booleans)

3. **Implement Array Operators**
   - `opIn`, `opNin`
   - DSL functions: `In(values ...any) Op`, `Nin(values ...any) Op`

4. **Implement String Operators**
   - `opRegex`, `opContains`, `opStartsWith`, `opEndsWith`
   - DSL functions: `Regex()`, `Contains()`, `StartsWith()`, `EndsWith()`

5. **Implement Other Operators**
   - `opExists`, `opSize`
   - DSL functions: `Exists(bool)`, `Size(n int)`
   - `All(ops ...Op) Op` for combining operators

6. **Implement Operators Methods** (`operators.go`)
   - `Clone() Operators`
   - `With(name string, fn OperatorFunc) Operators`
   - `WithAll(ops Operators) Operators`
   - `Without(names ...string) Operators`

### Testing
- **Operator Tests** (`operators_test.go`)
  - Test each operator function (eq, ne, gt, etc.) with various types
  - Test `$in` with arrays, empty arrays, single values
  - Test `$regex` with valid and invalid patterns
  - Test `$contains`, `$startsWith`, `$endsWith` with strings
  - Test `$exists` with nil and non-nil values
  - Test `$size` with arrays of different lengths
  - Test operator DSL functions create correct Op structures
  - Test `All()` combines multiple operators
  - Test Operators methods: Clone, With, WithAll, Without
  - Test custom operator registration and usage

### Success Criteria
- ✅ All 14 built-in operators implemented
- ✅ All operator DSL functions working
- ✅ Operators immutability verified
- ✅ All tests passing (>95% coverage)

**Phase 2 Completed** ✅

---

## Phase 3: Conditions Engine

### Objectives
- Implement condition compilation and evaluation
- Support nested conditions and logical operators
- Handle template variables

### Tasks
1. **Create Condition Evaluator** (`evaluator.go`)
   - `condCompiler` struct with operators and variables
   - `compile(cond Cond) Condition` - compiles Cond to evaluator function
   - `Condition` type as `func(subject Subject) bool`

2. **Implement Field Evaluation**
   - Extract field value using `subject.GetField(field)`
   - Apply operators to field values
   - Handle missing fields (nil values)

3. **Implement Logical Operators Evaluation**
   - `$and` - all conditions must match
   - `$or` - at least one condition must match
   - `$not` - condition must not match

4. **Implement Template Variable Resolution**
   - Detect `VarRef` in condition values
   - Replace with actual values from variables map
   - Handle missing variables (error or false)

5. **Implement Implicit Equality**
   - Bare values (no operator) treated as `$eq`
   - Example: `{"status": "draft"}` → `{"status": {"$eq": "draft"}}`

### Testing
- **Evaluator Tests** (`evaluator_test.go`)
  - Test simple field equality: `{"status": "draft"}`
  - Test all operators with various field types
  - Test nested conditions: `{"$and": [{...}, {...}]}`
  - Test `$or` with multiple branches
  - Test `$not` negation
  - Test template variables: `{"authorID": Var("UserID")}`
  - Test missing fields return false
  - Test missing variables handling
  - Test complex nested conditions
  - Test lazy compilation with sync.Once

### Success Criteria
- ✅ Condition compilation working correctly
- ✅ All operators integrated
- ✅ Template variables resolved properly
- ✅ All tests passing (>95% coverage)

**Phase 3 Completed** ✅

---

## Phase 4: Rule System

### Objectives
- Implement Rule structure
- Create RuleBuilder with fluent API
- Support allow/forbid rules

### Tasks
1. **Create Rule Types** (`rule.go`)
   - `Rule[S Subject]` struct with Action, Inverted, Cond, Fields, Reason
   - `RuleBuilder[S Subject]` struct

2. **Implement Rule Builder**
   - `Allow[S Subject](action ActionFor[S]) *RuleBuilder[S]`
   - `Forbid[S Subject](action ActionFor[S]) *RuleBuilder[S]`
   - `Where(cond Cond) *RuleBuilder[S]` - set conditions
   - `OnFields(fields ...string) *RuleBuilder[S]` - field restrictions
   - `Because(reason string) *RuleBuilder[S]` - set reason
   - `Build() Rule[S]` - construct final rule

3. **Implement Rule Validation**
   - Ensure action is not nil
   - Validate field names are non-empty
   - Ensure reason is set for forbid rules (optional but recommended)

### Testing
- **Rule Tests** (`rule_test.go`)
  - Test `Allow()` creates non-inverted rule
  - Test `Forbid()` creates inverted rule
  - Test `Where()` sets conditions
  - Test `OnFields()` sets field restrictions
  - Test `Because()` sets reason
  - Test method chaining works correctly
  - Test `Build()` produces correct Rule struct
  - Test unconditional rules (no Where clause)
  - Test complex condition rules
  - Test field-restricted rules

### Success Criteria
- ✅ Rule builder API is fluent and intuitive
- ✅ Allow/Forbid rules created correctly
- ✅ All tests passing (>90% coverage)

**Phase 4 Completed** ✅

---

## Phase 5: Ability Builder & Indexing

### Objectives
- Implement AbilityBuilder with configuration
- Create Ability with rule indexing
- Implement performance optimizations

### Tasks
1. **Create Ability Types** (`ability.go`)
   - `AbilityBuilder` struct with operators, vars, rules
   - `Ability` struct with index, operators, compiler

2. **Implement Ability Builder**
   - `NewAbility() *AbilityBuilder` - creates builder with DefaultOperators
   - `WithOperators(ops Operators) *AbilityBuilder`
   - `WithVars(vars map[string]any) *AbilityBuilder`
   - `Build() *Ability` - constructs Ability with indexing

3. **Implement Rule Addition**
   - `AddRule[S Subject](ab *AbilityBuilder, rule Rule[S]) *AbilityBuilder`
   - `AddRules[S Subject](ab *AbilityBuilder, rules ...Rule[S]) *AbilityBuilder`
   - Store rules internally as type-erased interface

4. **Implement Rule Indexing** (`index.go`)
   - `ruleIndex` struct: `map[subjectType]map[action][]compiledRule`
   - Index rules by subject type and action at build time
   - `compiledRule` struct with rule, compiled condition, sync.Once

5. **Implement Lazy Compilation**
   - Compile conditions on first evaluation
   - Use sync.Once for thread-safe lazy initialization
   - Cache compiled conditions for reuse

6. **Implement Merged Rule Caching**
   - Cache merged rule lists per subject/action pair
   - Invalidate on ability rebuild (immutable after Build())

### Testing
- **Ability Builder Tests** (`ability_test.go`)
  - Test `NewAbility()` creates builder with DefaultOperators
  - Test `WithOperators()` replaces operators
  - Test `WithVars()` sets template variables
  - Test `AddRule()` adds single rule
  - Test `AddRules()` adds multiple rules
  - Test `Build()` creates Ability with indexes
  - Test builder immutability (methods return new instances)

- **Indexing Tests** (`index_test.go`)
  - Test rules indexed by subject type
  - Test rules indexed by action within subject
  - Test multiple rules for same subject/action
  - Test lazy compilation on first use
  - Test merged rule caching
  - Test index lookup performance (O(1))

### Success Criteria
- ✅ AbilityBuilder API is clean and functional
- ✅ Rule indexing provides O(1) lookup
- ✅ Lazy compilation working correctly
- ✅ All tests passing (>90% coverage)

**Phase 5 Completed** ✅

---

## Phase 6: Permission Checking

### Objectives
- Implement core permission checking functions
- Handle allow/forbid rule precedence
- Support field-level permissions

### Tasks
1. **Implement Basic Permission Checks** (`permissions.go`)
   - `Can[S Subject](a *Ability, action ActionFor[S], subject S) bool`
   - `Cannot[S Subject](a *Ability, action ActionFor[S], subject S) bool`

2. **Implement Permission Logic**
   - Retrieve rules from index for subject type and action
   - Evaluate conditions for each rule
   - Apply rule precedence: forbid rules override allow rules
   - Return true if any allow rule matches and no forbid rules match

3. **Implement Field-Level Checks**
   - `CanWithField[S Subject](a *Ability, action ActionFor[S], subject S, field string) bool`
   - Check if field is in allowed fields list
   - Check if field is not in forbidden fields list
   - Handle rules with no field restrictions (all fields)

4. **Implement Multi-Action Checks**
   - `CanAll[S Subject](a *Ability, subject S, actions ...ActionFor[S]) bool`
   - `CanAny[S Subject](a *Ability, subject S, actions ...ActionFor[S]) bool`

5. **Handle Edge Cases**
   - No rules defined (default deny)
   - Unconditional rules (always match)
   - Empty field restrictions (all fields allowed)
   - Conflicting rules (forbid wins)

### Testing
- **Permission Tests** (`permissions_test.go`)
  - Test `Can()` with matching allow rule returns true
  - Test `Can()` with no rules returns false
  - Test `Cannot()` is inverse of Can()
  - Test forbid rule overrides allow rule
  - Test multiple allow rules (any match = true)
  - Test conditions evaluated correctly
  - Test template variables in conditions
  - Test unconditional allow/forbid rules
  - Test `CanWithField()` with field restrictions
  - Test `CanWithField()` with no field restrictions
  - Test `CanAll()` with multiple actions
  - Test `CanAny()` with multiple actions
  - Test type safety at compile time

### Success Criteria
- ✅ Permission checking logic correct
- ✅ Rule precedence handled properly
- ✅ Field-level permissions working
- ✅ All tests passing (>95% coverage)

**Phase 6 Completed** ✅

---

## Phase 7: Introspection Functions

### Objectives
- Implement introspection for debugging and UI
- Provide reasons for denials
- Return allowed/forbidden fields

### Tasks
1. **Implement WhyNot** (`introspection.go`)
   - `WhyNot[S Subject](a *Ability, action ActionFor[S], subject S) string`
   - Return reason from first matching forbid rule
   - Return generic message if denied but no forbid rule
   - Return empty string if allowed

2. **Implement RulesFor**
   - `RulesFor[S Subject](a *Ability, action ActionFor[S], subject S) []RuleInfo`
   - `RuleInfo` struct with Action, Subject, Inverted, Fields, Reason, Matched
   - Return all rules for action/subject with match status

3. **Implement Field Introspection**
   - `AllowedFields[S Subject](a *Ability, action ActionFor[S], subject S) []string`
   - `ForbiddenFields[S Subject](a *Ability, action ActionFor[S], subject S) []string`
   - Compute allowed fields from matching allow rules
   - Compute forbidden fields from matching forbid rules
   - Handle rules with no field restrictions (all fields)

### Testing
- **Introspection Tests** (`introspection_test.go`)
  - Test `WhyNot()` returns reason from forbid rule
  - Test `WhyNot()` returns generic message when denied without forbid
  - Test `WhyNot()` returns empty string when allowed
  - Test `RulesFor()` returns all matching rules
  - Test `RulesFor()` sets Matched=true for matching conditions
  - Test `RulesFor()` sets Matched=false for non-matching conditions
  - Test `AllowedFields()` returns correct field list
  - Test `AllowedFields()` with no field restrictions
  - Test `ForbiddenFields()` returns correct field list
  - Test field introspection with multiple rules

### Success Criteria
- ✅ Introspection functions provide useful debugging info
- ✅ WhyNot gives clear denial reasons
- ✅ Field introspection accurate
- ✅ All tests passing (>90% coverage)

**Phase 7 Completed** ✅

---

## Phase 8: JSON Loading

### Objectives
- Load rules from JSON files
- Support template variable syntax in JSON
- Map JSON to type-safe rules

### Tasks
1. **Create JSON Types** (`json.go`)
   - `JSONRule` struct with Action (StringOrSlice), Subject, Inverted, Conditions, Fields, Reason
   - `JSONRuleSet` struct with Rules array
   - `StringOrSlice` custom type with UnmarshalJSON
   - `LoadOptions` struct with Operators, Vars, SubjectRegistry, ActionValidators

2. **Implement Template Variable Parsing**
   - Detect `{{ .VarName }}` syntax in JSON string values
   - Convert to `VarRef` during parsing
   - Support nested template variables in conditions

3. **Implement JSON Loading**
   - `LoadFromJSON(data []byte, opts LoadOptions) (*Ability, error)`
   - `LoadFromFile(path string, opts LoadOptions) (*Ability, error)`
   - `LoadFromReader(r io.Reader, opts LoadOptions) (*Ability, error)`

4. **Implement Subject Registry**
   - Use SubjectRegistry to validate subject types
   - Validate actions using ActionValidators (if provided)
   - Handle multiple actions in single rule (expand to multiple rules)

5. **Handle JSON Validation**
   - Validate required fields (action, subject)
   - Validate operator syntax in conditions
   - Return descriptive errors for invalid JSON

### Testing
- **JSON Tests** (`json_test.go`)
  - Test `StringOrSlice` unmarshals single string
  - Test `StringOrSlice` unmarshals array of strings
  - Test template variable detection: `{{ .UserID }}`
  - Test `LoadFromJSON()` with valid JSON
  - Test `LoadFromJSON()` with template variables
  - Test `LoadFromFile()` loads from file
  - Test subject registry validation
  - Test action validator validation
  - Test multiple actions expand to multiple rules
  - Test invalid JSON returns error
  - Test missing required fields return error
  - Test loaded rules work with Can()
  - Test complex nested conditions in JSON

### Success Criteria
- ✅ JSON loading working correctly
- ✅ Template variables parsed and resolved
- ✅ Validation prevents invalid rules
- ✅ All tests passing (>90% coverage)

**Phase 8 Completed** ✅

---

## Phase 9: Performance Optimizations

### Objectives
- Benchmark current implementation
- Optimize hot paths
- Validate performance characteristics

### Tasks
1. **Create Benchmarks** (`benchmark_test.go`)
   - Benchmark `Can()` with no conditions (should be O(1))
   - Benchmark `Can()` with simple conditions
   - Benchmark `Can()` with complex nested conditions
   - Benchmark rule indexing performance
   - Benchmark condition compilation
   - Benchmark with 100, 1000, 10000 rules

2. **Optimize Index Lookups**
   - Profile index access patterns
   - Optimize map lookups
   - Verify O(1) lookup complexity

3. **Optimize Condition Evaluation**
   - Profile condition evaluation
   - Optimize hot operator functions
   - Cache compiled conditions effectively

4. **Optimize Field Access**
   - Verify GetField() is efficiently called
   - Minimize allocations in field evaluation
   - Profile reflection if used

5. **Memory Optimizations**
   - Profile memory allocations
   - Reduce unnecessary copying
   - Optimize rule storage

### Testing
- **Performance Tests** (`benchmark_test.go`)
  - Run all benchmarks
  - Compare with performance targets from spec
  - Verify O(1) for indexed operations
  - Verify O(n) only for condition evaluation where n = matching rules
  - Profile memory allocations

### Success Criteria
- ✅ Can() with unconditional rules: O(1)
- ✅ Can() with conditions: O(n) where n = matching rules
- ✅ Rule indexing: O(1) lookup
- ✅ Memory usage reasonable (no leaks)
- ✅ Benchmarks documented in README

---

## Phase 10: Complete Integration Testing & Examples

### Objectives
- Write comprehensive integration tests
- Create complete working examples
- Write documentation and README

### Tasks
1. **Create Integration Tests** (`integration_test.go`)
   - Test complete Article/Comment example from spec
   - Test multi-subject authorization
   - Test JSON loading + permission checking
   - Test all operators in real scenarios
   - Test field-level permissions end-to-end
   - Test template variables in complex scenarios

2. **Create Examples** (`examples/`)
   - `examples/basic/` - Simple getting started example
   - `examples/blog/` - Article/Comment system from spec
   - `examples/json/` - Loading rules from JSON
   - `examples/custom_operators/` - Custom operator example
   - Each with working `main.go` and `rules.json`

3. **Write Documentation**
   - Complete README.md with:
     - Installation instructions
     - Quick start guide
     - Full API documentation
     - Examples
     - Performance characteristics
     - Comparison with CASL.js
   - Add godoc comments to all public APIs
   - Create CONTRIBUTING.md
   - Create LICENSE file

4. **Create Migration Guide**
   - Document differences from CASL.js
   - Provide Go idioms and best practices
   - Include common patterns

5. **Final Validation**
   - Run all tests with race detector: `go test -race ./...`
   - Run tests with coverage: `go test -cover ./...`
   - Ensure coverage >90%
   - Lint code: `golangci-lint run`
   - Format code: `go fmt ./...`

### Testing
- **Integration Tests**
  - Full end-to-end user workflow
  - Multi-subject authorization scenarios
  - Complex permission hierarchies
  - JSON loading + execution
  - Error handling paths
  - Thread safety (race detector)

### Success Criteria
- ✅ All integration tests passing
- ✅ Examples run successfully
- ✅ README is comprehensive
- ✅ Code coverage >90%
- ✅ No race conditions
- ✅ All linters passing

**Phase 10 Completed** ✅

---

## Implementation Timeline

| Phase | Estimated Complexity | Dependencies |
|-------|---------------------|--------------|
| 1. Core Types | Low | None |
| 2. Operators | Medium | Phase 1 |
| 3. Conditions Engine | High | Phases 1-2 |
| 4. Rule System | Medium | Phase 1 |
| 5. Ability Builder | High | Phases 1-4 |
| 6. Permission Checking | High | Phases 1-5 |
| 7. Introspection | Medium | Phases 1-6 |
| 8. JSON Loading | Medium | Phases 1-7 |
| 9. Performance | Medium | Phases 1-8 |
| 10. Integration | Medium | Phases 1-9 |

---

## Key Files to Create

```
gocasl/
├── go.mod
├── go.sum
├── README.md
├── LICENSE
├── CONTRIBUTING.md
├── core.go              # Subject, Cond, Op, VarRef, Operators types
├── core_test.go
├── action.go            # ActionFor[S], DefineAction()
├── action_test.go
├── condition.go         # And(), Or(), Not(), Var()
├── condition_test.go
├── operators.go         # Built-in operators, DSL functions
├── operators_test.go
├── evaluator.go         # Condition compilation and evaluation
├── evaluator_test.go
├── rule.go              # Rule[S], RuleBuilder[S], Allow(), Forbid()
├── rule_test.go
├── ability.go           # AbilityBuilder, Ability, AddRule()
├── ability_test.go
├── index.go             # Rule indexing internals
├── index_test.go
├── permissions.go       # Can(), Cannot(), CanWithField(), etc.
├── permissions_test.go
├── introspection.go     # WhyNot(), RulesFor(), AllowedFields(), etc.
├── introspection_test.go
├── json.go              # JSON loading, LoadOptions
├── json_test.go
├── benchmark_test.go    # Performance benchmarks
├── integration_test.go  # End-to-end tests
├── examples/
│   ├── basic/
│   │   └── main.go
│   ├── blog/
│   │   ├── main.go
│   │   └── rules.json
│   ├── json/
│   │   ├── main.go
│   │   └── rules.json
│   └── custom_operators/
│       └── main.go
└── internal/            # Internal helpers if needed
    └── compare.go       # Type comparison utilities
```

---

## Testing Strategy

### Unit Tests
- Each public function has dedicated tests
- Edge cases covered (nil, empty, invalid inputs)
- Error paths tested
- Type safety validated at compile time

### Integration Tests
- End-to-end workflows
- Multi-component interaction
- Real-world scenarios from spec

### Benchmarks
- Performance regression prevention
- Validate O(1) and O(n) complexity claims
- Memory allocation tracking

### Coverage Goals
- Overall: >90%
- Core logic (evaluator, permissions): >95%
- Error handling paths: 100%

---

## Notes

1. **Go Version**: Requires Go 1.21+ for generics support
2. **Dependencies**: Minimize external dependencies (standard library preferred)
3. **Thread Safety**: Ability must be thread-safe after Build()
4. **Immutability**: Builders create new instances on method calls
5. **Type Safety**: Leverage generics to prevent runtime type errors
6. **Performance**: Index-based lookups, lazy compilation, merged caching
