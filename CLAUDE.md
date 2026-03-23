# CLAUDE.md

## Project Overview

gocasl (`github.com/infisical/gocasl`) is a type-safe, isomorphic authorization library for Go, inspired by CASL.js. It uses a MongoDB-style query language for defining permission conditions.

## Build & Test

```bash
# Build
go build ./...

# Run all tests
go test ./...

# Run tests verbose
go test -v ./...

# Run benchmarks
go test -bench=. -benchmem ./...

# Run a specific test
go test -run TestName ./...
```

No Makefile or special scripts — standard Go tooling only. Zero external dependencies.

## Project Structure

```
├── core.go           # Core types: Subject, FieldOp, CondOp, CompileCtx, MapSubject, Cond, Op
├── ability.go        # AbilityBuilder and Ability
├── action.go         # Type-safe action definitions (DefineAction)
├── rule.go           # Rule builder (Allow/Forbid)
├── condition.go      # Condition DSL helpers (And, Or, Not, Var)
├── operators.go      # All operators: comparison funcs, FieldOps, CondOps, DSL helpers
├── evaluator.go      # Condition compiler (condCompiler) — compiles Cond → Condition closures
├── permissions.go    # Permission checking (Can, Cannot, WhyNot)
├── introspection.go  # Rule introspection API
├── index.go          # Rule indexing (SubjectType → Action → compiledRules)
├── json.go           # JSON rule loading
├── pack.go           # Rule serialization
├── internal/compare/ # Comparison utilities (Compare, Equal, Contains)
└── examples/         # Usage examples (basic, blog, json, custom_operator)
```

## Architecture

### Two-phase evaluation

1. **Build time** — `AbilityBuilder.Build()` validates rules and creates the `Ability`. Conditions are lazily compiled into closures on first use via `sync.Once`.
2. **Eval time** — `Can(ability, action, subject)` looks up indexed rules and evaluates compiled conditions.

### Operator system

Two compile-time operator types (defined in `core.go`):

- **`FieldOp`** `func(cc *CompileCtx, field string, constraint any) Condition` — field-level operators (`$eq`, `$gt`, `$elemMatch`, `$all`, etc.). Simple comparison operators are wrapped with `Compare(fn)`.
- **`CondOp`** `func(cc *CompileCtx, value any) Condition` — condition-level logical operators (`$and`, `$or`, `$not`).

Both receive `CompileCtx` with `Compile()` (recursive sub-condition compilation) and `Resolve()` (VarRef resolution). This means any operator — including user-defined custom ones — can compile nested conditions.

Registration: `DefaultFieldOps()` / `DefaultCondOps()` return the built-in sets. Override via `AbilityBuilder.WithFieldOps()` / `WithCondOps()`.

### Key conventions

- Operators are pre-compiled at build time, not evaluated at runtime — no caching needed.
- `$and`/`$or`/`$not` are registered `CondOp`s, not hardcoded in the compiler.
- Forbid rules always take precedence over allow rules.
- Rules are indexed by `SubjectType → Action` for O(1) lookup.
- All `FieldOps`/`CondOps` methods (`Clone`, `With`, `Without`, `WithAll`) are immutable — they return new maps.

## Code Style

- Go 1.26, uses generics for type-safe actions (`ActionFor[S Subject]`)
- No external dependencies
- Tests use standard `testing` package, no test framework
- Test files mirror source: `foo.go` → `foo_test.go`
- Package name: `gocasl` (single package, no sub-packages except `internal/compare`)
