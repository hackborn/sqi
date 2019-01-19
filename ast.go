package sqi

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

// ------------------------------------------------------------
// AST-NODE

// AstNode defines the basic interface for evaluating expressions.
type AstNode interface {
	// Evaluate the interface, returning the results. This is intended
	// to  be only an internal API: the Opt* is a pointer but can not be nil.
	Eval(interface{}, *Opt) (interface{}, error)
}

// ------------------------------------------------------------
// BINARY-NODE

// binaryNode performs binary operations on the current interface{}.
type binaryNode struct {
	Op  symbol
	Lhs AstNode
	Rhs AstNode
}

func (n *binaryNode) Eval(_i interface{}, opt *Opt) (interface{}, error) {
	// fmt.Println("Eval binaryNode", n.Lhs, n.Rhs)
	if !(n.Op > start_binary && n.Op < end_binary) || n.Lhs == nil || n.Rhs == nil {
		return nil, newMalformedError("binary node")
	}
	switch n.Op {
	case eql_token:
		return n.evalEql(_i, opt)
	case neq_token:
		resp, err := n.evalEql(_i, opt)
		if err != nil {
			return false, err
		}
		return !resp, err
	case and_token:
		return n.evalAnd(_i, opt)
	case or_token:
		return n.evalOr(_i, opt)
	default:
		return nil, newUnhandledError("binary " + strconv.Itoa(int(n.Op)))
	}
}

func (n *binaryNode) evalEql(_i interface{}, opt *Opt) (bool, error) {
	lhs, rhs, err := n.evalBinary(_i, opt)
	if err != nil {
		return false, err
	}
	eq, err := interfacesEqual(lhs, rhs)
	if err != nil && opt != nil && opt.Strict {
		return false, err
	}
	return eq, nil
}

func (n *binaryNode) evalAnd(_i interface{}, opt *Opt) (bool, error) {
	lhs, rhs, err := n.evalBinary(_i, opt)
	if err != nil {
		return false, err
	}
	ls, lok := lhs.(bool)
	rs, rok := rhs.(bool)
	if !lok || !rok {
		return false, newConditionError("&& must evaluate to boolean")
	}
	return ls && rs, nil
}

func (n *binaryNode) evalOr(_i interface{}, opt *Opt) (bool, error) {
	lhs, rhs, err := n.evalBinary(_i, opt)
	if err != nil {
		return false, err
	}
	ls, lok := lhs.(bool)
	rs, rok := rhs.(bool)
	if !lok || !rok {
		return false, newConditionError("|| must evaluate to boolean")
	}
	return ls || rs, nil
}

func (n *binaryNode) evalBinary(_i interface{}, opt *Opt) (interface{}, interface{}, error) {
	lhs, err := n.Lhs.Eval(_i, opt)
	if err != nil {
		return nil, nil, err
	}
	rhs, err := n.Rhs.Eval(_i, opt)
	if err != nil {
		return nil, nil, err
	}
	return lhs, rhs, nil
}

// ------------------------------------------------------------
// CONDITION-NODE

// conditionNode is a unary that filters the incoming interface
// by a boolean condition. The node it contains must respond with true or false.
type conditionNode struct {
	Op    symbol
	Child AstNode
}

func (n *conditionNode) Eval(_i interface{}, opt *Opt) (interface{}, error) {
	//	fmt.Println("Eval conditionNode", n.lhs, n.rhs)
	if !(n.Op > start_unary && n.Op < end_unary) || n.Child == nil {
		return nil, newMalformedError("condition node")
	}
	// Every item in _i is evaluated to true or false.
	// We need to distinguish between slices, arrays, and single items
	rt := reflect.TypeOf(_i)
	switch rt.Kind() {
	case reflect.Slice:
		src := reflect.Indirect(reflect.ValueOf(_i))
		dst := reflect.MakeSlice(rt, 0, src.Len())
		for i := 0; i < src.Len(); i++ {
			item := src.Index(i)
			b, err := n.isTrue(item.Interface(), opt)
			if err != nil {
				return nil, err
			}
			if b {
				dst = reflect.Append(dst, item)
			}
		}
		return dst.Interface(), nil
	case reflect.Array:
		fmt.Println("condition is an array with element type", rt.Elem())
		return nil, newUnhandledError("conditionNode.Eval() on reflect.Array")
	default:
		// When working on collections, we return a new one, but when working
		// on single objects, we return the results of the evaluation.
		return n.isTrue(_i, opt)
	}
}

// isTrue() determines if my child evaluates to true based on the input.
func (n *conditionNode) isTrue(_i interface{}, opt *Opt) (bool, error) {
	resp, err := n.Child.Eval(_i, opt)
	if err != nil {
		return false, err
	}
	if b, ok := resp.(bool); ok {
		return b, nil
	}
	return false, newConditionError("must be boolean")
}

// ------------------------------------------------------------
// CONSTANT-NODE

// constantNode returns a constant value (string, float, etc.).
type constantNode struct {
	Value interface{}
}

func (n *constantNode) Eval(_i interface{}, opt *Opt) (interface{}, error) {
	//	fmt.Println("Evak constantNode", n.Value)
	return n.Value, nil
}

// ------------------------------------------------------------
// FIELD-NODE

// fieldNode is used to select a field from the current interface{}.
type fieldNode struct {
	Field string
}

func (n *fieldNode) Eval(_i interface{}, opt *Opt) (interface{}, error) {
	// fmt.Println("Eval fieldNode", n.Field)
	if len(n.Field) < 1 {
		return nil, newMalformedError("field node")
	}
	// Errors
	rt := reflect.TypeOf(_i)
	switch rt.Kind() {
	case reflect.Array:
		return nil, newConditionError("fieldNode must not receive reflect.Array")
	case reflect.Slice:
		return nil, newConditionError("fieldNode must not receive reflect.Slice")
	}

	var child interface{}
	var err error
	switch t := _i.(type) {
	case map[string]interface{}:
		child = t[n.Field]
	case reflect.Value:
		return nil, newConditionError("fieldNode must not receive reflect.Value")
	default:
		child, err = n.runOnValue(reflect.Indirect(reflect.ValueOf(_i)))
	}
	if child == nil || err != nil {
		return nil, err
	}
	if tt, ok := child.(reflect.Value); ok {
		return tt.Interface(), nil
	}
	return child, err
}

func (n *fieldNode) runOnValue(v reflect.Value) (interface{}, error) {
	f := v.FieldByName(n.Field)
	if f.IsValid() {
		return f, nil
	}
	return nil, errors.New("No field for " + n.Field)
}

// ------------------------------------------------------------
// PATH-NODE

// pathNode combines two expressions.
type pathNode struct {
	Child AstNode `json:"child,omitempty"`
	Field AstNode `json:"field,omitempty"`
}

func (n *pathNode) Eval(i interface{}, opt *Opt) (interface{}, error) {
	// fmt.Println("Eval pathNode", n.Child, n.Field)
	if n.Field == nil {
		return nil, newMalformedError("path node")
	}
	if n.Child != nil {
		var err error
		i, err = n.Child.Eval(i, opt)
		if err != nil {
			return nil, err
		}
	}
	return n.Field.Eval(i, opt)
}

// ------------------------------------------------------------
// UNARY-NODE

// unaryNode performs a unary operation on the current interface{}.
type unaryNode struct {
	Op    symbol
	Child AstNode
}

func (n *unaryNode) Eval(_i interface{}, opt *Opt) (interface{}, error) {
	//	fmt.Println("Eval unaryNode", n.Child)
	if n.Child == nil {
		return nil, newMalformedError("unary node")
	}
	return n.Child.Eval(_i, opt)
}

// ----------------------------------------
// MISC

// clone() is a clever way to copy a slice, but I don't think I need or want it.
func clone(i interface{}) interface{} {
	// Wrap argument to reflect.Value, dereference it and return back as interface{}
	return reflect.Indirect(reflect.ValueOf(i)).Interface()
}
