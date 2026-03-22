package gocasl

import (
	"maps"
	"reflect"
	"regexp"

	"github.com/infisical/gocasl/internal/compare"
)

// DefaultOperators returns a fresh copy of the built-in operators.
// Each call returns an independent copy that is safe to modify.
func DefaultOperators() Operators {
	return defaultOperators()
}

func defaultOperators() Operators {
	return Operators{
		"$eq":       opEq,
		"$ne":       opNe,
		"$gt":       opGt,
		"$gte":      opGte,
		"$lt":       opLt,
		"$lte":      opLte,
		"$in":       opIn,
		"$nin":      opNin,
		"$regex":    opRegex,
		"$contains": opContains,
		"$exists":   opExists,
		"$size":     opSize,
	}
}

// Clone returns a shallow copy of the operators map.
func (o Operators) Clone() Operators {
	newOps := make(Operators, len(o))
	maps.Copy(newOps, o)
	return newOps
}

// With adds or replaces an operator.
func (o Operators) With(name string, fn OperatorFunc) Operators {
	newOps := o.Clone()
	newOps[name] = fn
	return newOps
}

// Without removes operators by name.
func (o Operators) Without(names ...string) Operators {
	newOps := o.Clone()
	for _, name := range names {
		delete(newOps, name)
	}
	return newOps
}

// WithAll merges another set of operators into this one.
func (o Operators) WithAll(other Operators) Operators {
	newOps := o.Clone()
	maps.Copy(newOps, other)
	return newOps
}

// --- Comparison Operators ---

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

// --- Array Operators ---

func opIn(val any, constraint any) bool {
	// constraint must be a slice/array
	return compare.Contains(constraint, val)
}

func opNin(val any, constraint any) bool {
	return !compare.Contains(constraint, val)
}

// --- String Operators ---

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

// --- Other Operators ---

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
		// try float/int64 if json decoding messes up types
		valC := reflect.ValueOf(constraint)
		if compare.Equal(constraint, int(0)) { // simple check
			// fall back to reflection if needed, but int is expected from DSL
			if valC.CanConvert(reflect.TypeFor[int]()) {
				size = int(valC.Convert(reflect.TypeFor[int]()).Int())
			} else {
				return false
			}
		} else {
			// This path handles if constraint came in as float64 (common in JSON)
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

// StartsWith is a convenience DSL that uses Regex
func StartsWith(prefix string) Op {
	return Op{"$regex": "^" + regexp.QuoteMeta(prefix)}
}

// EndsWith is a convenience DSL that uses Regex
func EndsWith(suffix string) Op {
	return Op{"$regex": regexp.QuoteMeta(suffix) + "$"}
}
