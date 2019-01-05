package sqi

// interfacesEqual() answers true if both interfaces are the same underlying data.
func interfacesEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	} else if a == nil || b == nil {
		return false
	}
	switch at := a.(type) {
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
			return at == bt
		}
	}
	return false
}
