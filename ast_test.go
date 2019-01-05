package sqi

import (
	"encoding/json"
	"fmt"
	"testing"
)

// --------------------------------------------------------------------------------------
// TEST-AST-GET

func TestAstGet(t *testing.T) {
	cases := []struct {
		Input    interface{}
		Expr     AstNode
		WantResp interface{}
		WantErr  error
	}{
		{ast_get_input_0, ast_get_expr_0, "a", nil},
		{ast_get_input_1, ast_get_expr_1, Child{"ca"}, nil},
		{ast_get_input_2, ast_get_expr_2, []Child{Child{"ca"}, Child{"cb"}}, nil},
		{ast_get_input_3, ast_get_expr_3, []Child{Child{"cb"}}, nil},
	}
	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			runTestAstGet(t, tc.Input, tc.Expr, tc.WantResp, tc.WantErr)
			// Everything that works on the struct variant should work on the unmarshalled json.
			runTestAstGet(t, toJson(tc.Input), tc.Expr, tc.WantResp, tc.WantErr)
		})
	}
}

func runTestAstGet(t *testing.T, input interface{}, expr AstNode, want_resp interface{}, want_err error) {
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
	ast_get_input_0 = Child{Val: "a"}
	ast_get_expr_0  = field_n("Val")

	ast_get_input_1 = &Parent{Child: Child{"ca"}}
	ast_get_expr_1  = field_n("Child")

	ast_get_input_2 = &MultiParent{Children: []Child{Child{"ca"}, Child{"cb"}}}
	ast_get_expr_2  = field_n("Children")

	ast_get_input_3 = &MultiParent{Children: []Child{Child{"ca"}, Child{"cb"}, Child{"cc"}}}
	ast_get_expr_3  = path_n(children_node, get_val_cb_node)

	children_node   = field_n("Children")
	get_val_cb_node = eql_n(field_n("Val"), string_n("cb"))
)

// --------------------------------------------------------------------------------------
// BUILD

func eql_n(lhs, rhs AstNode) AstNode {
	return &binaryNode{Op: eql_token, Lhs: lhs, Rhs: rhs}
}

func field_n(name string) AstNode {
	return &fieldNode{Field: name}
}

func paren_n(child AstNode) AstNode {
	return &unaryNode{Op: open_token, Child: child}
}

func path_n(lhs, rhs AstNode) AstNode {
	return &pathNode{Lhs: lhs, Rhs: rhs}
}

func string_n(value string) AstNode {
	return &stringNode{Value: value}
}

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
