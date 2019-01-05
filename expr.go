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
// FIELD-NODE

// fieldNode is used to select a field from the current interface{}.
type fieldNode struct {
	field string
}

func (n *fieldNode) Run(_i interface{}) (interface{}, error) {
	//	fmt.Println("Run fieldNode", n.field)
	if len(n.field) < 1 {
		return nil, errors.New("Missing fieled select")
	}
	var child interface{}
	var err error
	switch t := _i.(type) {
	case map[string]interface{}:
		child = t[n.field]
	case reflect.Value:
		return nil, errors.New("Internal error: fieldNode must not receive reflect.Value")
		//		child, err = n.runOnValue(t)
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
	f := v.FieldByName(n.field)
	if f.IsValid() {
		return f, nil
	}
	return nil, errors.New("No field for " + n.field)
}

// ------------------------------------------------------------
// BINARY-OP-NODE

// binaryOpNode performs binary operations on the current interface{}.
type binaryOpNode struct {
	op  string
	lhs string
	rhs string
}

func (n *binaryOpNode) Run(_i interface{}) (interface{}, error) {
	//	fmt.Println("Run binaryOpNode", n.lhs, n.rhs)
	if len(n.op) < 1 || len(n.lhs) < 1 || len(n.rhs) < 1 {
		return nil, errors.New("Invalid binary")
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
		fmt.Println(n.lhs, "is an array with element type", rt.Elem())
		return nil, errors.New("Unhandled binaryOpNode.Run() on reflect.Array")
	default:
		fmt.Println(n.lhs, "is something else entirely")
		return nil, errors.New("Unhandled binaryOpNode.Run() on default")
	}
	return _i, nil
}

func (n *binaryOpNode) runEquals(_i interface{}) (bool, error) {
	iv := reflect.Indirect(reflect.ValueOf(_i)).Interface()
	fs := &fieldNode{field: n.lhs}
	_v, err := fs.Run(iv)
	if err != nil {
		return false, err
	}
	if v, ok := _v.(string); ok {
		return v == n.rhs, nil
	}
	return false, nil
}

// ------------------------------------------------------------
// PATH-NODE

// pathNode combines two expressions.
type pathNode struct {
	lhs AstNode
	rhs AstNode
}

func (n *pathNode) Run(_i interface{}) (interface{}, error) {
	//	fmt.Println("Run pathNode", n.lhs, n.rhs)
	if n.lhs == nil || n.rhs == nil {
		return nil, errors.New("Invalid path")
	}
	ans, err := n.lhs.Run(_i)
	if err != nil {
		return nil, err
	}
	return n.rhs.Run(ans)
}

// ----------------------------------------
// MISC

// clone() is a clever way to copy a slice, but I don't think I need or want it.
func clone(i interface{}) interface{} {
	// Wrap argument to reflect.Value, dereference it and return back as interface{}
	return reflect.Indirect(reflect.ValueOf(i)).Interface()
}

// ----------------------------------------
// CONST and VAR
