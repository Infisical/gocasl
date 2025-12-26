package compare

import (
	"reflect"
	"strings"
)

// Compare returns -1 if a < b, 0 if a == b, 1 if a > b.
// It returns -2 if values are not comparable.
// It handles numeric type coercion (e.g. comparing int and float).
func Compare(a, b any) int {
	if a == nil || b == nil {
		if a == b {
			return 0
		}
		return -2
	}

	// Fast path for string
	if as, ok := a.(string); ok {
		if bs, ok := b.(string); ok {
			if as < bs { return -1 }
			if as > bs { return 1 }
			return 0
		}
	}

	// Fast path for int
	if ai, ok := a.(int); ok {
		if bi, ok := b.(int); ok {
			if ai < bi { return -1 }
			if ai > bi { return 1 }
			return 0
		}
	}

	valA := reflect.ValueOf(a)
	valB := reflect.ValueOf(b)

	// Handle Strings (if not caught by fast path, e.g. custom string type? unlikely for 'any')
	if valA.Kind() == reflect.String && valB.Kind() == reflect.String {
		sa, sb := valA.String(), valB.String()
		if sa < sb {
			return -1
		}
		if sa > sb {
			return 1
		}
		return 0
	}

	// Handle Bools
	if valA.Kind() == reflect.Bool && valB.Kind() == reflect.Bool {
		ba, bb := valA.Bool(), valB.Bool()
		if ba == bb {
			return 0
		}
		if !ba && bb {
			return -1
		}
		return 1
	}

	// Handle Numerics
	if isNumeric(valA.Kind()) && isNumeric(valB.Kind()) {
		fa, fb := toFloat64(valA), toFloat64(valB)
		if fa < fb {
			return -1
		}
		if fa > fb {
			return 1
		}
		return 0
	}

	return -2
}

// Equal checks if a and b are equal, handling numeric coercion.
func Equal(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	
	// Fast path for same type equality (basic types)
	if a == b {
		return true
	}

	// Fast path for int vs ...
	if ai, ok := a.(int); ok {
		if bi, ok := b.(int); ok {
			return ai == bi
		}
	}
	
	valA := reflect.ValueOf(a)
	valB := reflect.ValueOf(b)

	if isNumeric(valA.Kind()) && isNumeric(valB.Kind()) {
		return toFloat64(valA) == toFloat64(valB)
	}

	return reflect.DeepEqual(a, b)
}

func isNumeric(k reflect.Kind) bool {
	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}

func toFloat64(v reflect.Value) float64 {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(v.Uint())
	case reflect.Float32, reflect.Float64:
		return v.Float()
	default:
		return 0
	}
}

// Contains check for strings and arrays/slices
func Contains(collection any, item any) bool {
    val := reflect.ValueOf(collection)
    switch val.Kind() {
    case reflect.String:
        itemStr, ok := item.(string)
        if !ok {
            return false
        }
        return strings.Contains(val.String(), itemStr)
    case reflect.Slice, reflect.Array:
        for i := 0; i < val.Len(); i++ {
            if Equal(val.Index(i).Interface(), item) {
                return true
            }
        }
    }
    return false
}