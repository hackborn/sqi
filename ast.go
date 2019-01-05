package sqi

import (
	"errors"
	"fmt"
	"reflect"
)

// ------------------------------------------------------------
// AST-NODE

// AstNode defines the basic interface for evaluating expressions.
type AstNode interface {
	Run(interface{}) (interface{}, error)
}

// ------------------------------------------------------------
// BINARY-NODE

// binaryNode performs binary operations on the current interface{}.
type binaryNode struct {
	Op  Token
	Lhs AstNode
	Rhs AstNode
}

func (n *binaryNode) Run(_i interface{}) (interface{}, error) {
	//	fmt.Println("Run binaryNode", n.lhs, n.rhs)
	if !(n.Op > start_binary && n.Op < end_binary) || n.Lhs == nil || n.Rhs == nil {
		return nil, errors.New("sqi: invalid binary")
	}
	// Every item in _i that has a Lhs matching Rhs is included.
	// We need to distinguish between slices, arrays, and single items
	rt := reflect.TypeOf(_i)
	switch rt.Kind() {
	case reflect.Slice:
		src := reflect.Indirect(reflect.ValueOf(_i))
		dst := reflect.MakeSlice(rt, 0, src.Len())
		for i := 0; i < src.Len(); i++ {
			item := src.Index(i)
			b, err := n.runEquals(item.Interface())
			if err != nil {
				return nil, err
			}
			if b {
				dst = reflect.Append(dst, item)
			}
		}
		return dst.Interface(), nil
	case reflect.Array:
		fmt.Println(n.Lhs, "is an array with element type", rt.Elem())
		return nil, errors.New("sqi: Unhandled binaryNode.Run() on reflect.Array")
	default:
		return n.runEquals(_i)
	}
	return _i, nil
}

func (n *binaryNode) runEquals(_i interface{}) (bool, error) {
	lhs, err := n.Lhs.Run(_i)
	if err != nil {
		return false, err
	}
	rhs, err := n.Rhs.Run(_i)
	if err != nil {
		return false, err
	}
	eq := interfacesEqual(lhs, rhs)
	return eq, nil
}

// ------------------------------------------------------------
// FIELD-NODE

// fieldNode is used to select a field from the current interface{}.
type fieldNode struct {
	Field string
}

func (n *fieldNode) Run(_i interface{}) (interface{}, error) {
	//	fmt.Println("Run fieldNode", n.field)
	if len(n.Field) < 1 {
		return nil, errors.New("Missing fieled select")
	}
	var child interface{}
	var err error
	switch t := _i.(type) {
	case map[string]interface{}:
		child = t[n.Field]
	case reflect.Value:
		return nil, errors.New("Internal error: fieldNode must not receive reflect.Value")
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
// FLOAT-NODE

// floatNode returns a constant float64.
type floatNode struct {
	Value float64
}

func (n *floatNode) Run(_i interface{}) (interface{}, error) {
	//	fmt.Println("Run floatNode", n.Value)
	return n.Value, nil
}

// ------------------------------------------------------------
// INT-NODE

// intNode returns a constant int.
type intNode struct {
	Value int
}

func (n *intNode) Run(_i interface{}) (interface{}, error) {
	//	fmt.Println("Run intNode", n.Value)
	return n.Value, nil
}

// ------------------------------------------------------------
// PATH-NODE

// pathNode combines two expressions.
type pathNode struct {
	Lhs AstNode
	Rhs AstNode
}

func (n *pathNode) Run(_i interface{}) (interface{}, error) {
	//	fmt.Println("Run pathNode", n.lhs, n.rhs)
	if n.Lhs == nil || n.Rhs == nil {
		return nil, errors.New("Invalid path")
	}
	ans, err := n.Lhs.Run(_i)
	if err != nil {
		return nil, err
	}
	return n.Rhs.Run(ans)
}

// ------------------------------------------------------------
// STRING-NODE

// stringNode returns a constant string.
type stringNode struct {
	Value string
}

func (n *stringNode) Run(_i interface{}) (interface{}, error) {
	//	fmt.Println("Run stringNode", n.Value)
	return n.Value, nil
}

// ------------------------------------------------------------
// UNARY-NODE

// unaryNode performs a unary operation on the current interface{}.
type unaryNode struct {
	Op    Token
	Child AstNode
}

func (n *unaryNode) Run(_i interface{}) (interface{}, error) {
	//	fmt.Println("Run unaryNode", n.Child)
	if n.Child == nil {
		return nil, errors.New("Invalid unary")
	}
	return n.Child.Run(_i)
}

// ----------------------------------------
// MISC

// clone() is a clever way to copy a slice, but I don't think I need or want it.
func clone(i interface{}) interface{} {
	// Wrap argument to reflect.Value, dereference it and return back as interface{}
	return reflect.Indirect(reflect.ValueOf(i)).Interface()
}
