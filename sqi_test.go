package sqi

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
)

// --------------------------------------------------------------------------------------
// TEST-LEXER

func TestLexer(t *testing.T) {
	cases := []struct {
		Input    string
		WantResp []token_t
		WantErr  error
	}{
		{`(Child)`, tokens(`(`, `Child`, `)`), nil},
		{`(Child)`, tokens(`(`, `Child`, `)`), nil},
		{`Child/Name`, tokens(`Child`, `/`, `Name`), nil},
		{`Child		/	Name`, tokens(`Child`, `/`, `Name`), nil},
		{`Child/Name=="a"`, tokens(`Child`, `/`, `Name`, `==`, `"a"`), nil},
		{`Child/Name=="a b"`, tokens(`Child`, `/`, `Name`, `==`, `"a b"`), nil},
		{`Child / Name == "a"`, tokens(`Child`, `/`, `Name`, `==`, `"a"`), nil},
		{`Child/(Name=="a"||Age==10)`, tokens(`Child`, `/`, `(`, `Name`, `==`, `"a"`, `||`, `Age`, `==`, 10, `)`), nil},
		{`Child / (Name == "a" || Age == 10)`, tokens(`Child`, `/`, `(`, `Name`, `==`, `"a"`, `||`, `Age`, `==`, 10, `)`), nil},
	}
	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			have_resp, have_err := scan(tc.Input)
			if !errorMatches(have_err, tc.WantErr) {
				fmt.Println("Error mismatch, have\n", have_err, "\nwant\n", tc.WantErr)
				t.Fatal()
			} else if !tokensMatch(have_resp, tc.WantResp) {
				fmt.Println("Token mismatch, have\n", have_resp, "\nwant\n", tc.WantResp)
				t.Fatal()
			}
		})
	}
}

// --------------------------------------------------------------------------------------
// TEST-PARSER

func TestParser(t *testing.T) {
	cases := []struct {
		Input    []token_t
		WantResp AstNode
		WantErr  error
	}{
		{tokens(`Child`), parser_want_0, nil},
		{tokens(`Child`, `/`, `Name`), parser_want_1, nil},
		{tokens(`Child`, `/`, `Arm`, `/`, `Length`), parser_want_2, nil},
		{tokens(`Child`, `/`, `Name`, `==`, `a`), parser_want_3, nil},
		{tokens(`(`, `Child`, `)`), parser_want_4, nil},
		{tokens(`Child`, `/`, `(`, `Name`, `==`, `a`, `)`), parser_want_5, nil},
		{tokens(`(`, `Child`, `/`, `Name`, `)`, `==`, `a`), parser_want_6, nil},
		{tokens(`Child`, `/`, `Age`, `==`, 10), parser_want_7, nil},
		{tokens(`Child`, `/`, `Height`, `==`, 5.5), parser_want_8, nil},
	}
	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			have_resp, have_err := parse(tc.Input)
			if !errorMatches(have_err, tc.WantErr) {
				fmt.Println("Error mismatch, have\n", have_err, "\nwant\n", tc.WantErr)
				t.Fatal()
			} else if !interfaceMatches(have_resp, tc.WantResp) {
				fmt.Println("Ast mismatch, have\n", toJsonString(have_resp), "\nwant\n", toJsonString(tc.WantResp))
				t.Fatal()
			}
		})
	}
}

var (
	parser_want_0 = field_n(`Child`)
	parser_want_1 = path_n(field_n(`Child`), field_n(`Name`))
	parser_want_2 = path_n(field_n(`Child`), path_n(field_n(`Arm`), field_n(`Length`)))
	parser_want_3 = path_n(field_n(`Child`), eql_n(field_n(`Name`), string_n(`a`)))
	parser_want_4 = paren_n(field_n(`Child`))
	parser_want_5 = path_n(field_n(`Child`), paren_n(eql_n(field_n(`Name`), string_n(`a`))))
	parser_want_6 = eql_n(paren_n(path_n(field_n(`Child`), field_n(`Name`))), string_n(`a`))
	parser_want_7 = path_n(field_n(`Child`), eql_n(field_n(`Age`), int_n(10)))
	parser_want_8 = path_n(field_n(`Child`), eql_n(field_n(`Height`), float_n(5.5)))
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
// TEST-EXPR

func TestExpr(t *testing.T) {
	cases := []struct {
		TermInput string
		EvalInput interface{}
		WantResp  interface{}
		WantErr   error
	}{
		{`Child/Val`, expr_eval_input_0, `ca`, nil},
		{`(Child/Val) == ca`, expr_eval_input_0, true, nil},
	}
	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			expr, err := MakeExpr(tc.TermInput)
			if err != nil {
				panic(err)
			}
			have_resp, have_err := expr.Eval(tc.EvalInput)
			if !errorMatches(have_err, tc.WantErr) {
				fmt.Println("Error mismatch, have\n", have_err, "\nwant\n", tc.WantErr)
				t.Fatal()
			} else if !interfacesEqual(have_resp, tc.WantResp) {
				fmt.Println("Response mismatch, have\n", toJsonString(have_resp), "\nwant\n", toJsonString(tc.WantResp))
				t.Fatal()
			}
		})
	}
}

var (
	expr_eval_input_0 = &Parent{Child: Child{"ca"}}
	expr_eval_input_1 = &Parent{Child: Child{"ca"}}
)

// --------------------------------------------------------------------------------------
// MODEL

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
// BUILD

func tokens(all ...interface{}) []token_t {
	var tokens []token_t
	for _, t := range all {
		switch v := t.(type) {
		case float32:
			val := strconv.FormatFloat(float64(v), 'f', 8, 64)
			tokens = append(tokens, token_t{float_token, val})
		case float64:
			val := strconv.FormatFloat(v, 'f', 8, 64)
			tokens = append(tokens, token_t{float_token, val})
		case int:
			tokens = append(tokens, token_t{int_token, strconv.Itoa(v)})
		case string:
			tokens = append(tokens, token_t{string_token, v}.reclassify())
		}
	}
	return tokens
}

func eql_n(lhs, rhs AstNode) AstNode {
	return &binaryNode{Op: eql_token, Lhs: lhs, Rhs: rhs}
}

func field_n(name string) AstNode {
	return &fieldNode{Field: name}
}

func float_n(value float64) AstNode {
	return &floatNode{Value: value}
}

func int_n(value int) AstNode {
	return &intNode{Value: value}
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

func tokensMatch(a, b []token_t) bool {
	if len(a) != len(b) {
		return false
	} else if len(a) < 1 {
		return true
	}
	for i, t := range a {
		if b[i] != t {
			return false
		}
	}
	return true
}

// --------------------------------------------------------------------------------------
// CONVERT

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
