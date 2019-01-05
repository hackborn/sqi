package sqi

import (
	"math"
)

// interfacesEqual() answers true if both interfaces are the same underlying data.
func interfacesEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	} else if a == nil || b == nil {
		return false
	}
	switch at := a.(type) {
	case bool:
		if bt, ok := b.(bool); ok {
			return at == bt
		}
	case string:
		if bt, ok := b.(string); ok {
			return at == bt
		}
	case int:
		if bt, ok := b.(int); ok {
			return at == bt
		}
	case float32:
		if bt, ok := b.(float32); ok {
			return float64Equal(float64(at), float64(bt))
		}
	case float64:
		if bt, ok := b.(float64); ok {
			return float64Equal(at, bt)
		}
	}
	return false
}

// ------------------------------------------------------------
// MISC

func float64Equal(a, b float64) bool {
	return math.Abs(a-b) <= float64EqualityThreshold
}

// ------------------------------------------------------------
// CONST and VAR

const (
	float64EqualityThreshold = 1e-9
)
