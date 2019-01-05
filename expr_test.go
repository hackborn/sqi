package sqi

import (
	"encoding/json"
	"fmt"
	"testing"
)

// --------------------------------------------------------------------------------------
// TEST-EXPR-GET

func TestExprGet(t *testing.T) {
	cases := []struct {
		Input    interface{}
		Expr     AstNode
		WantResp interface{}
		WantErr  error
	}{
		{expr_get_input_0, expr_get_expr_0, "a", nil},
		{expr_get_input_1, expr_get_expr_1, Child{"ca"}, nil},
		{expr_get_input_2, expr_get_expr_2, []Child{Child{"ca"}, Child{"cb"}}, nil},
		{expr_get_input_3, expr_get_expr_3, []Child{Child{"cb"}}, nil},
	}
	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			runTestExprGet(t, tc.Input, tc.Expr, tc.WantResp, tc.WantErr)
			// Everything that works on the struct variant should work on the unmarshalled json.
			runTestExprGet(t, toJson(tc.Input), tc.Expr, tc.WantResp, tc.WantErr)
		})
	}
}

func runTestExprGet(t *testing.T, input interface{}, expr AstNode, want_resp interface{}, want_err error) {
	have_resp, have_err := expr.Run(input)
	// fmt.Println("have_resp", have_resp, "have_err", have_err)
	if !errorMatches(have_err, want_err) {
		fmt.Println("Error mismatch, have\n", have_err, "\nwant\n", want_err)
		t.Fatal()
	} else if !interfaceMatches(have_resp, want_resp) {
		fmt.Println("Response mismatch, have\n", toJsonString(have_resp), "\nwant\n", toJsonString(want_resp))
		t.Fatal()
	}
}

var (
	expr_get_input_0 = Child{Val: "a"}
	expr_get_expr_0  = &fieldNode{Field: "Val"}

	expr_get_input_1 = &Parent{Child: Child{"ca"}}
	expr_get_expr_1  = &fieldNode{Field: "Child"}

	expr_get_input_2 = &MultiParent{Children: []Child{Child{"ca"}, Child{"cb"}}}
	expr_get_expr_2  = &fieldNode{Field: "Children"}

	expr_get_input_3 = &MultiParent{Children: []Child{Child{"ca"}, Child{"cb"}, Child{"cc"}}}
	expr_get_expr_3  = &pathNode{Lhs: children_node, Rhs: get_val_cb_node}

	children_node   = &fieldNode{Field: "Children"}
	get_val_cb_node = &binaryOpNode{Op: "==", Lhs: "Val", Rhs: "cb"}
)

// --------------------------------------------------------------------------------------
// TEST-EXPR-GET MODEL

type Child struct {
	Val string `json:"Val,omitempty"`
}

type Parent struct {
	Child Child `json:"Child,omitempty"`
}

type ParentPtr struct {
	Child *Child `json:"Child,omitempty"`
}

type MultiParent struct {
	Children []Child `json:"Children,omitempty"`
}

type MultiParentPtr struct {
	Children []*Child `json:"Children,omitempty"`
}

// --------------------------------------------------------------------------------------
// COMPARE

func errorMatches(a, b error) bool {
	if a == nil && b == nil {
		return true
	} else if a == nil || b == nil {
		return false
	}
	return a.Error() == b.Error()
}

func interfaceMatches(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	} else if a == nil || b == nil {
		return false
	}
	ja := toJsonString(a)
	jb := toJsonString(b)
	return ja == jb
}

func toJson(i interface{}) interface{} {
	pbytes, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}
	i2 := make(map[string]interface{})
	err = json.Unmarshal(pbytes, &i2)
	if err != nil {
		panic(err)
	}
	return i2
}

func toJsonString(i interface{}) string {
	// The item might already be a string -- this makes tests easier to read
	if s, ok := i.(string); ok {
		return s
	}
	pbytes, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}
	return string(pbytes)
}
