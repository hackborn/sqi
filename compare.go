package sqi

import (
	"math"
	"reflect"
)

// interfacesEqual() answers true if both interfaces are the same underlying data.
// If strict is true, numbers must be of the same type. If it's false,
// numbers will attempt to convert for a comparison.
func interfacesEqual(a, b interface{}, strict bool) (bool, error) {
	if a == nil && b == nil {
		return true, nil
	} else if a == nil || b == nil {
		return false, nil
	}
	switch at := a.(type) {
	case bool:
		if bt, ok := b.(bool); ok {
			return at == bt, nil
		}
		return false, newMismatchError("types " + reflect.TypeOf(a).Name() + " and " + reflect.TypeOf(b).Name())
	case string:
		if bt, ok := b.(string); ok {
			return at == bt, nil
		}
		return false, newMismatchError("types " + reflect.TypeOf(a).Name() + " and " + reflect.TypeOf(b).Name())
	case int:
		if bt, ok := b.(int); ok {
			return at == bt, nil
		}
		if !strict {
			return numbersEqual(float64(at), b)
		}
		return false, newMismatchError("types " + reflect.TypeOf(a).Name() + " and " + reflect.TypeOf(b).Name())
	case float32:
		if bt, ok := b.(float32); ok {
			return float64Equal(float64(at), float64(bt)), nil
		}
		if !strict {
			return numbersEqual(float64(at), b)
		}
		return false, newMismatchError("types " + reflect.TypeOf(a).Name() + " and " + reflect.TypeOf(b).Name())
	case float64:
		if bt, ok := b.(float64); ok {
			return float64Equal(at, bt), nil
		}
		if !strict {
			return numbersEqual(at, b)
		}
		return false, newMismatchError("types " + reflect.TypeOf(a).Name() + " and " + reflect.TypeOf(b).Name())
	default:
		return false, newUnhandledError("type " + reflect.TypeOf(a).Name())
	}
}

// ------------------------------------------------------------
// MISC

func numbersEqual(a float64, b interface{}) (bool, error) {
	switch bt := b.(type) {
	case int:
		return float64Equal(a, float64(bt)), nil
	case float32:
		return float64Equal(a, float64(bt)), nil
	case float64:
		return float64Equal(a, bt), nil
	default:
		return false, newUnhandledError("type " + reflect.TypeOf(a).Name())
	}
}

func float64Equal(a, b float64) bool {
	return math.Abs(a-b) <= float64EqualityThreshold
}

// ------------------------------------------------------------
// CONST and VAR

const (
	float64EqualityThreshold = 1e-9
)
