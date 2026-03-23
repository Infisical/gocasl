package gocasl

import (
	"reflect"
	"regexp"

	"github.com/infisical/gocasl/internal/compare"
)

// DefaultFieldOps returns a fresh copy of the built-in field operators.
// Each call returns an independent copy that is safe to modify.
func DefaultFieldOps() FieldOps {
	return defaultFieldOps()
}

// DefaultCondOps returns a fresh copy of the built-in condition operators.
// Each call returns an independent copy that is safe to modify.
func DefaultCondOps() CondOps {
	return defaultCondOps()
}

func defaultFieldOps() FieldOps {
	return FieldOps{
		"$eq":        Compare(opEq),
		"$ne":        Compare(opNe),
		"$gt":        Compare(opGt),
		"$gte":       Compare(opGte),
		"$lt":        Compare(opLt),
		"$lte":       Compare(opLte),
		"$in":        Compare(opIn),
		"$nin":       Compare(opNin),
		"$regex":     Compare(opRegex),
		"$contains":  Compare(opContains),
		"$exists":    Compare(opExists),
		"$size":      Compare(opSize),
		"$elemMatch": opElemMatchFieldOp,
		"$all":       opAllFieldOp,
	}
}

func defaultCondOps() CondOps {
	return CondOps{
		"$and": opAndCondOp,
		"$or":  opOrCondOp,
		"$not": opNotCondOp,
	}
}

// --- Comparison Functions ---
// These are pure comparison functions used with Compare() wrapper.

func opEq(val any, constraint any) bool {
	return compare.Equal(val, constraint)
}

func opNe(val any, constraint any) bool {
	return !compare.Equal(val, constraint)
}

func opGt(val any, constraint any) bool {
	return compare.Compare(val, constraint) == 1
}

func opGte(val any, constraint any) bool {
	res := compare.Compare(val, constraint)
	return res == 1 || res == 0
}

func opLt(val any, constraint any) bool {
	return compare.Compare(val, constraint) == -1
}

func opLte(val any, constraint any) bool {
	res := compare.Compare(val, constraint)
	return res == -1 || res == 0
}

// --- Array Comparison Functions ---

func opIn(val any, constraint any) bool {
	return compare.Contains(constraint, val)
}

func opNin(val any, constraint any) bool {
	return !compare.Contains(constraint, val)
}

// --- String Comparison Functions ---

func opRegex(val any, constraint any) bool {
	valStr, ok1 := val.(string)
	pattern, ok2 := constraint.(string)
	if !ok1 || !ok2 {
		return false
	}
	matched, _ := regexp.MatchString(pattern, valStr)
	return matched
}

func opContains(val any, constraint any) bool {
	return compare.Contains(val, constraint)
}

// --- Other Comparison Functions ---

func opExists(val any, constraint any) bool {
	shouldExist, ok := constraint.(bool)
	if !ok {
		return false
	}
	if shouldExist {
		return val != nil
	}
	return val == nil
}

func opSize(val any, constraint any) bool {
	size, ok := constraint.(int)
	if !ok {
		valC := reflect.ValueOf(constraint)
		if compare.Equal(constraint, int(0)) {
			if valC.CanConvert(reflect.TypeFor[int]()) {
				size = int(valC.Convert(reflect.TypeFor[int]()).Int())
			} else {
				return false
			}
		} else {
			if f, ok := constraint.(float64); ok {
				size = int(f)
			} else {
				return false
			}
		}
	}

	v := reflect.ValueOf(val)
	switch v.Kind() {
	case reflect.String, reflect.Slice, reflect.Array, reflect.Map:
		return v.Len() == size
	}
	return false
}

// --- Compound Field Operators ---

// opElemMatchFieldOp checks if any element in an array field matches all the sub-conditions.
func opElemMatchFieldOp(cc *CompileCtx, field string, constraint any) Condition {
	subCond := cc.Compile(toCond(constraint))

	return func(s Subject) bool {
		fieldVal := s.GetField(field)
		return arrayAny(fieldVal, subCond)
	}
}

// opAllFieldOp checks if every element in an array field matches all the sub-conditions.
func opAllFieldOp(cc *CompileCtx, field string, constraint any) Condition {
	subCond := cc.Compile(toCond(constraint))

	return func(s Subject) bool {
		fieldVal := s.GetField(field)
		return arrayAll(fieldVal, subCond)
	}
}

// arrayAny returns true if any element in the slice/array satisfies the condition.
func arrayAny(fieldVal any, subCond Condition) bool {
	rv := reflect.ValueOf(fieldVal)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return false
	}
	for i := 0; i < rv.Len(); i++ {
		elem := rv.Index(i).Interface()
		if subj := toSubject(elem); subj != nil && subCond(subj) {
			return true
		}
	}
	return false
}

// arrayAll returns true if every element in the slice/array satisfies the condition.
// Returns false for empty or non-array values.
func arrayAll(fieldVal any, subCond Condition) bool {
	rv := reflect.ValueOf(fieldVal)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return false
	}
	if rv.Len() == 0 {
		return false
	}
	for i := 0; i < rv.Len(); i++ {
		elem := rv.Index(i).Interface()
		if subj := toSubject(elem); subj == nil || !subCond(subj) {
			return false
		}
	}
	return true
}

// toSubject wraps a value as a Subject for sub-condition evaluation.
func toSubject(elem any) Subject {
	if subj, ok := elem.(Subject); ok {
		return subj
	}
	if m, ok := elem.(map[string]any); ok {
		return MapSubject(m)
	}
	return nil
}

// toCond converts a value to a Cond for sub-condition compilation.
func toCond(val any) Cond {
	switch v := val.(type) {
	case Cond:
		return v
	case map[string]any:
		return Cond(v)
	}
	return nil
}

// --- Condition-Level Operators ---

func opAndCondOp(cc *CompileCtx, value any) Condition {
	list, ok := value.([]any)
	if !ok {
		return func(s Subject) bool { return false }
	}

	var conds []Condition
	for _, item := range list {
		conds = append(conds, cc.Compile(toCond(item)))
	}

	return func(s Subject) bool {
		for _, c := range conds {
			if !c(s) {
				return false
			}
		}
		return true
	}
}

func opOrCondOp(cc *CompileCtx, value any) Condition {
	list, ok := value.([]any)
	if !ok {
		return func(s Subject) bool { return false }
	}

	var conds []Condition
	for _, item := range list {
		conds = append(conds, cc.Compile(toCond(item)))
	}

	return func(s Subject) bool {
		if len(conds) == 0 {
			return false
		}
		for _, c := range conds {
			if c(s) {
				return true
			}
		}
		return false
	}
}

func opNotCondOp(cc *CompileCtx, value any) Condition {
	cond := cc.Compile(toCond(value))
	return func(s Subject) bool {
		return !cond(s)
	}
}

// --- DSL Functions ---

func Eq(val any) Op           { return Op{"$eq": val} }
func Ne(val any) Op           { return Op{"$ne": val} }
func Gt(val any) Op           { return Op{"$gt": val} }
func Gte(val any) Op          { return Op{"$gte": val} }
func Lt(val any) Op           { return Op{"$lt": val} }
func Lte(val any) Op          { return Op{"$lte": val} }
func In(vals ...any) Op       { return Op{"$in": vals} }
func Nin(vals ...any) Op      { return Op{"$nin": vals} }
func Regex(pattern string) Op { return Op{"$regex": pattern} }
func Contains(val any) Op     { return Op{"$contains": val} }
func Exists(exists bool) Op   { return Op{"$exists": exists} }
func Size(size int) Op        { return Op{"$size": size} }

// ElemMatch checks if any element in an array field matches all the sub-conditions.
func ElemMatch(cond Cond) Op { return Op{"$elemMatch": cond} }

// All checks if every element in an array field matches all the sub-conditions.
func All(cond Cond) Op { return Op{"$all": cond} }

// StartsWith is a convenience DSL that uses Regex
func StartsWith(prefix string) Op {
	return Op{"$regex": "^" + regexp.QuoteMeta(prefix)}
}

// EndsWith is a convenience DSL that uses Regex
func EndsWith(suffix string) Op {
	return Op{"$regex": regexp.QuoteMeta(suffix) + "$"}
}
