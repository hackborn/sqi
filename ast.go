package sqi

import (
	"errors"
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
// ARRAY-NODE

// arrayNode performs an array indexing. Currently it supports
// a single int index.
type arrayNode struct {
	Lhs   AstNode // Optional -- if missing then I just use my input directly
	Index int
}

func (n *arrayNode) Eval(_i interface{}, opt *Opt) (interface{}, error) {
	// fmt.Println("Eval arrayNode", n.Lhs, n.Index)

	lhs := _i
	if n.Lhs != nil {
		var err error
		lhs, err = n.Lhs.Eval(_i, opt)
		if err != nil {
			return nil, err
		}
	}

	// We need to decide on how to handle invalid node input:
	// I'm currently inclined to just returned no result, since this
	// is a search.
	if lhs == nil {
		return nil, nil
	}

	// We need to distinguish between slices, arrays, and single items
	rt := reflect.TypeOf(lhs)
	switch rt.Kind() {
	case reflect.Array, reflect.Slice:
		src := reflect.Indirect(reflect.ValueOf(lhs))
		// This matches what I expect of a single-index array access
		// (that it returns a single requested value)
		if src.Len() < 1 {
			return nil, nil
		}
		if n.Index >= 0 && n.Index < src.Len() {
			item := src.Index(n.Index)
			return item.Interface(), nil
		}
	}

	// When not in strict mode, invalid arrays are pass-throughs.
	if opt != nil && opt.Strict {
		return nil, newEvalError("operator [] must have array or slice")
	}
	return lhs, nil
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
	if n.Lhs == nil || n.Rhs == nil {
		return nil, newMalformedError("binary node")
	}
	switch n.Op {
	case eqlToken:
		return n.evalEql(_i, opt)
	case neqToken:
		resp, err := n.evalEql(_i, opt)
		if err != nil {
			return false, err
		}
		return !resp, err
	case andToken:
		return n.evalAnd(_i, opt)
	case orToken:
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
	strict := false
	if opt != nil {
		strict = opt.Strict
	}
	eq, err := interfacesEqual(lhs, rhs, strict)
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
	if _i == nil {
		return nil, nil
	}
	// Collectinons with special handling.
	rt := reflect.TypeOf(_i)
	ismap := false
	switch rt.Kind() {
	case reflect.Array:
		return nil, newConditionError("fieldNode must not receive reflect.Array")
	case reflect.Slice:
		return nil, newConditionError("fieldNode must not receive reflect.Slice")
	case reflect.Map:
		ismap = true
	}

	var child interface{}
	var err error
	switch t := _i.(type) {
	case map[string]interface{}:
		child = t[n.Field]
	case reflect.Value:
		return nil, newConditionError("fieldNode must not receive reflect.Value")
	default:
		// Special condition, this is a specific type of map, but we don't
		// know what kind.
		if ismap {
			key := reflect.ValueOf(n.Field)
			child = reflect.Indirect(reflect.ValueOf(_i)).MapIndex(key)
		} else {
			child, err = n.runOnValue(reflect.Indirect(reflect.ValueOf(_i)))
		}
	}
	if child == nil || err != nil {
		return nil, err
	}
	return n.getInterface(child)
}

func (n *fieldNode) runOnValue(v reflect.Value) (interface{}, error) {
	f := v.FieldByName(n.Field)
	if f.IsValid() {
		return f, nil
	}
	return nil, errors.New("No field for " + n.Field)
}

// getInterface() calls reflect.Value.Interface() "safely" by handling
// the panic. The reflect package doesn't seem to define any way to
// determine if caling Interface() is valid, and as far as I can see,
// rolling your own is fairly heavyweight.
func (n *fieldNode) getInterface(_i interface{}) (interface{}, error) {
	defer func() {
		if r := recover(); r != nil {
		}
	}()
	if i, ok := _i.(reflect.Value); ok {
		return i.Interface(), nil
	}
	return _i, nil
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
// SELECT-NODE

// selectnNode is a select statement: It expects a collection
// of items, and a child that will resolve each item to
// true or false. It answers the result of all true evaluations.
type selectNode struct {
	Child AstNode
}

func (n *selectNode) Eval(_i interface{}, opt *Opt) (interface{}, error) {
	// fmt.Println("Eval selectNode", n.Child)
	if n.Child == nil {
		return nil, newMalformedError("select node")
	}
	rt := reflect.TypeOf(_i)
	switch rt.Kind() {
	case reflect.Array, reflect.Slice:
		src := reflect.Indirect(reflect.ValueOf(_i))
		collectiontype := rt.Elem()
		dst := reflect.MakeSlice(reflect.SliceOf(collectiontype), 0, src.Len())
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
	default:
		// It's an open question what to do when operating selects on
		// non-collections. I'm inclined to think of this as a search,
		// and no search results are returned, but I could see a mode
		// that made this an error condition.
		return nil, nil
	}
}

// isTrue() determines if my child evaluates to true based on the input.
func (n *selectNode) isTrue(_i interface{}, opt *Opt) (bool, error) {
	resp, err := n.Child.Eval(_i, opt)
	if err != nil {
		return false, err
	}
	if b, ok := resp.(bool); ok {
		return b, nil
	}
	return false, newEvalError("must result ine boolean")
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
