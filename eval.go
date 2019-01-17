package sqi

// ------------------------------------------------------------
// EVAL

// Eval() runs term against input, returning the result.
func Eval(term string, input interface{}) (interface{}, error) {
	return EvalWith(term, input, Opt{})
}

// EvalWith() runs term against input and options, returning the result.
func EvalWith(term string, input interface{}, opt Opt) (interface{}, error) {
	expr, err := MakeExpr(term)
	if err != nil {
		return nil, err
	}
	return expr.Eval(input, &opt)
}

// EvalString() runs term against input, returning the string result.
// If an error occurs, the errorval is returned.
func EvalString(term string, input interface{}, errorval string) string {
	return EvalStringWith(term, input, errorval, Opt{})
}

// EvalString() runs term against input and options, returning the string result.
// If an error occurs, the errorval is returned.
func EvalStringWith(term string, input interface{}, errorval string, opt Opt) string {
	resp, err := EvalWith(term, input, opt)
	if resp == nil || err != nil {
		return errorval
	}
	if s, ok := resp.(string); ok {
		return s
	}
	return errorval
}

// ------------------------------------------------------------
// OPT

// Opt contains options for evaluation.
type Opt struct {
	// Strict causes sloppy conditions to become errors. For example, comparing a
	// number to a string is false if strict is off, but error if it's on.
	Strict bool
}
