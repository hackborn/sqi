package sqi

// ------------------------------------------------------------
// EVAL

// Eval() runs term against input, returning the result.
func Eval(term string, input interface{}) (interface{}, error) {
	expr, err := MakeExpr(term)
	if err != nil {
		return nil, err
	}
	return expr.Eval(input)
}

// EvalString() runs term against input, returning the string result.
// If an error occurs, the errorval is returned.
func EvalString(term string, input interface{}, errorval string) string {
	resp, err := Eval(term, input)
	if resp == nil || err != nil {
		return errorval
	}
	if s, ok := resp.(string); ok {
		return s
	}
	return errorval
}
