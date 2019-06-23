package sqi

import (
	"reflect"
)

// ------------------------------------------------------------
// EVAL

// Eval runs term against input, returning the result.
func Eval(term string, input interface{}, opt *Opt) (interface{}, error) {
	expr, err := MakeExpr(term)
	if err != nil {
		return nil, err
	}
	return expr.Eval(input, opt)
}

// EvalBool runs term against input, returning the boolean result.
// If an error occurs, the opt.OnError is returned.
func EvalBool(term string, input interface{}, opt *Opt) bool {
	resp, err := Eval(term, input, opt)
	if err == nil && resp != nil {
		if v, ok := resp.(bool); ok {
			return v
		}
	}
	if opt == nil {
		return false
	}
	return opt.onErrorBool()
}

// EvalFloat64 runs term against input, returning the float64 result.
// If an error occurs, the opt.OnError is returned.
func EvalFloat64(term string, input interface{}, opt *Opt) float64 {
	resp, err := Eval(term, input, opt)
	if err == nil && resp != nil {
		if v, ok := resp.(float64); ok {
			return v
		}
	}
	if opt == nil {
		return 0.0
	}
	return opt.onErrorFloat64()
}

// EvalInt runs term against input, returning the int result.
// If an error occurs, the opt.OnError is returned.
func EvalInt(term string, input interface{}, opt *Opt) int {
	resp, err := Eval(term, input, opt)
	if err == nil && resp != nil {
		if v, ok := resp.(int); ok {
			return v
		}
	}
	if opt == nil {
		return 0
	}
	return opt.onErrorInt()
}

// EvalString runs term against input, returning the string result.
// If an error occurs, the opt.OnError is returned.
func EvalString(term string, input interface{}, opt *Opt) string {
	resp, err := Eval(term, input, opt)
	if err == nil && resp != nil {
		if v, ok := resp.(string); ok {
			return v
		}
	}
	if opt == nil {
		return ""
	}
	return opt.onErrorString()
}

// EvalStringInterfaceMap runs term against input, returning the string map result.
// If an error occurs, the opt.OnError is returned or nil.
func EvalStringInterfaceMap(term string, input interface{}, opt *Opt) map[string]interface{} {
	resp, err := Eval(term, input, opt)
	if err == nil && resp != nil {
		switch v := resp.(type) {
		case map[string]interface{}:
			return v
		default:
			// Hand convert remaining cases. I really wish go had a way
			// to cast to a interface{} but I don't think it does.
			rt := reflect.TypeOf(resp)
			rv := reflect.ValueOf(resp)
			if rv.Kind() == reflect.Map && rt.Key().Kind() == reflect.String {
				m := make(map[string]interface{})
				iter := rv.MapRange()
				for iter.Next() {
					m[iter.Key().Interface().(string)] = iter.Value().Interface()
				}
				return m
			}
		}
	}
	if opt == nil {
		return nil
	}
	return opt.onErrorStringInterfaceMap()
}

// EvalStringStringMap runs term against input, returning the string map result.
// If an error occurs, the opt.OnError is returned or nil.
func EvalStringStringMap(term string, input interface{}, opt *Opt) map[string]string {
	resp, err := Eval(term, input, opt)
	if err == nil && resp != nil {
		switch v := resp.(type) {
		case map[string]string:
			return v
		default:
			// Hand convert remaining cases.
			rv := reflect.ValueOf(resp)
			if rv.Kind() == reflect.Map {
				m := make(map[string]string)
				iter := rv.MapRange()
				for iter.Next() {
					if s, ok := iter.Value().Interface().(string); ok {
						m[iter.Key().Interface().(string)] = s
					}
				}
				return m
			}
		}
	}
	if opt == nil {
		return nil
	}
	return opt.onErrorStringStringMap()
}

// EvalStringSlice runs term against input, returning the string slice result.
// If an error occurs, the opt.OnError is returned or nil.
func EvalStringSlice(term string, input interface{}, opt *Opt) []string {
	resp, err := Eval(term, input, opt)
	if err == nil && resp != nil {
		switch v := resp.(type) {
		case []string:
			return v
		default:
			// Hand convert remaining cases. I really wish go had a way
			// to cast to a interface{} but I don't think it does.
			rv := reflect.ValueOf(resp)
			if rv.Kind() == reflect.Slice {
				s := make([]string, 0, rv.Len())
				for i := 0; i < rv.Len(); i++ {
					intf := rv.Index(i).Interface()
					if intfs, ok := intf.(string); ok {
						s = append(s, intfs)
					}
				}
				return s
			}
		}
	}
	if opt == nil {
		return nil
	}
	return opt.onErrorStringSlice()
}

// ------------------------------------------------------------
// OPT

// Opt contains options for evaluation.
type Opt struct {
	// Strict causes sloppy conditions to become errors. For example, comparing a
	// number to a string is false if strict is off, but error if it's on.
	Strict bool
	// OnError is a value returned when one of the typed Eval() statements returns an error.
	// Must match the type. For example, the value must be assigend a string if using EvalString().
	OnError interface{}
}

func (o Opt) onErrorBool() bool {
	if v, ok := o.OnError.(bool); ok {
		return v
	}
	return false
}

func (o Opt) onErrorFloat64() float64 {
	if v, ok := o.OnError.(float64); ok {
		return v
	}
	return 0.0
}

func (o Opt) onErrorInt() int {
	if v, ok := o.OnError.(int); ok {
		return v
	}
	return 0
}

func (o Opt) onErrorString() string {
	if v, ok := o.OnError.(string); ok {
		return v
	}
	return ""
}

func (o Opt) onErrorStringInterfaceMap() map[string]interface{} {
	if v, ok := o.OnError.(map[string]interface{}); ok {
		return v
	}
	return nil
}

func (o Opt) onErrorStringStringMap() map[string]string {
	if v, ok := o.OnError.(map[string]string); ok {
		return v
	}
	return nil
}

func (o Opt) onErrorStringSlice() []string {
	if v, ok := o.OnError.([]string); ok {
		return v
	}
	return nil
}
