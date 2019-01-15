package sqi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"reflect"
	"testing"
	"strings"
)

// ------------------------------------------------------------
// TEST-LEXER

func TestLexer(t *testing.T) {
	cases := []struct {
		Input    string
		WantResp []*node_t
		WantErr  error
	}{
		{`a`, tokens(`a`), nil},
		{`(a)`, tokens(`(`, `a`, `)`), nil},
		{`a/b`, tokens(`a`, `/`, `b`), nil},
		{`a		/	b`, tokens(`a`, `/`, `b`), nil},
		{`a/b=="c"`, tokens(`a`, `/`, `b`, `==`, `"c"`), nil},
		{`a/b=="c d"`, tokens(`a`, `/`, `b`, `==`, `"c d"`), nil},
		{`a / b == "c"`, tokens(`a`, `/`, `b`, `==`, `"c"`), nil},
		{`a/(b=="c"||d==10)`, tokens(`a`, `/`, `(`, `b`, `==`, `"c"`, `||`, `d`, `==`, 10, `)`), nil},
		{`a / (b == "c" || d == 10)`, tokens(`a`, `/`, `(`, `b`, `==`, `"c"`, `||`, `d`, `==`, 10, `)`), nil},
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

// ------------------------------------------------------------
// TEST-PARSER

func TestParser(t *testing.T) {
	cases := []struct {
		Input    []*node_t
		WantResp AstNode
		WantErr  error
	}{
		/*
		{tokens(`a`), parser_want_0, nil},
		{tokens(`a`, `/`, `b`), parser_want_1, nil},
		{tokens(`a`, `/`, `b`, `/`, `c`), parser_want_2, nil},
		{tokens(`a`, `/`, `b`, `==`, `c`), parser_want_3, nil},
		{tokens(`(`, `a`, `)`), parser_want_4, nil},
		{tokens(`a`, `/`, `(`, `b`, `==`, `c`, `)`), parser_want_5, nil},
		{tokens(`(`, `a`, `/`, `b`, `)`, `==`, `c`), parser_want_6, nil},
		{tokens(`a`, `/`, `b`, `==`, 10), parser_want_7, nil},
		{tokens(`a`, `/`, `b`, `==`, 5.5), parser_want_8, nil},
		{tokens(`a`, `==`, `b`, `||`, `c`, `==`, `d`), parser_want_9, nil},
		{tokens(`(`, `a`, `==`, `b`, `)`, `||`, `(`, `c`, `==`, `d`, `)`), parser_want_9, nil},
		*/
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
	parser_want_0 = string_n(`a`)
	parser_want_1 = path_n(string_n(`a`), string_n(`b`))
	parser_want_2 = path_n(path_n(string_n(`a`), string_n(`b`)), string_n(`c`))
	parser_want_3 = path_n(string_n(`a`), eql_n(string_n(`b`), string_n(`c`)))
	parser_want_4 = string_n(`a`)
	parser_want_5 = path_n(string_n(`a`), eql_n(string_n(`b`), string_n(`c`)))
	parser_want_6 = eql_n(path_n(string_n(`a`), string_n(`b`)), string_n(`c`))
	parser_want_7 = path_n(string_n(`a`), eql_n(string_n(`b`), int_n(10)))
	parser_want_8 = path_n(string_n(`a`), eql_n(string_n(`b`), float_n(5.5)))
	parser_want_9 = or_n(eql_n(string_n(`a`), string_n(`b`)), eql_n(string_n(`c`), string_n(`d`)))
)

// ------------------------------------------------------------
// TEST-CONTEXTUALIZER

func TestContextualizer(t *testing.T) {
	cases := []struct {
		Input    *node_t
		WantResp *node_t
		WantErr  error
	}{
		{ctx_input_0, ctx_want_0, nil},
		{ctx_input_1, ctx_want_1, nil},
	}
	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			have_resp, have_err := contextualize(tc.Input)
			if !errorMatches(have_err, tc.WantErr) {
				fmt.Println("Error mismatch, have\n", have_err, "\nwant\n", tc.WantErr)
				t.Fatal()
			} else if !interfaceMatches(have_resp, tc.WantResp) {
				fmt.Println("Token mismatch, have\n", toJsonString(have_resp), "\nwant\n", toJsonString(tc.WantResp))
				t.Fatal()
			}
		})
	}
}

var (
	ctx_input_0 = newToken(string_token, `a`)
	ctx_want_0 = newToken(field_token, `a`)

	ctx_input_1 = mk_binary(eql_token, newToken(string_token, `a`), newToken(string_token, `b`))
	ctx_want_1 = mk_binary(eql_token, newToken(field_token, `a`), newToken(string_token, `b`))
)

// ------------------------------------------------------------
// TEST-AST-GET

func TestAstGet(t *testing.T) {
	cases := []struct {
		Input    interface{}
		Expr     AstNode
		Opts     Opt
		WantResp interface{}
		WantErr  error
	}{
		/*
			{ast_get_input_0, ast_get_expr_0, Opt{}, "a", nil},
			{ast_get_input_1, ast_get_expr_1, Opt{}, Relative{Name: "ca"}, nil},
			{ast_get_input_2, ast_get_expr_2, Opt{}, []Person{Person{Name: "ca"}, Person{Name: "cb"}}, nil},
			{ast_get_input_3, ast_get_expr_3, Opt{}, []Person{Person{Name: "cb"}}, nil},
		*/
	}
	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			runTestAstGet(t, tc.Input, tc.Expr, tc.Opts, tc.WantResp, tc.WantErr)
			// Everything that works on the struct variant should work on the unmarshalled json.
			runTestAstGet(t, toJson(tc.Input), tc.Expr, tc.Opts, tc.WantResp, tc.WantErr)
		})
	}
}

func runTestAstGet(t *testing.T, input interface{}, expr AstNode, opt Opt, want_resp interface{}, want_err error) {
	have_resp, have_err := expr.Eval(input, &opt)
	if !errorMatches(have_err, want_err) {
		fmt.Println("Error mismatch, have\n", have_err, "\nwant\n", want_err)
		t.Fatal()
	} else if !interfaceMatches(have_resp, want_resp) {
		fmt.Println("Response mismatch, have\n", toJsonString(have_resp), "\nwant\n", toJsonString(want_resp))
		t.Fatal()
	}
}

var (
	ast_get_input_0 = Person{Name: "a"}
	ast_get_expr_0  = field_n("Name")

	ast_get_input_1 = &Person{Mom: Relative{Name: "ca"}}
	ast_get_expr_1  = field_n("Mom")

	ast_get_input_2 = &Person{Children: []Person{Person{Name: "ca"}, Person{Name: "cb"}}}
	ast_get_expr_2  = field_n("Children")

	ast_get_input_3 = &Person{Children: []Person{Person{Name: "ca"}, Person{Name: "cb"}, Person{Name: "cc"}}}
	ast_get_expr_3  = path_n(field_n("Children"), cnd_n(eql_n(field_n("Name"), string_n("cb"))))

	children_node    = field_n("Children")
	get_name_cb_node = eql_n(field_n("Name"), string_n("cb"))
)

// ------------------------------------------------------------
// TEST-EXPR

// TestExpr() runs tests on the full system. It is technically
// unneeded, since the individual pieces have all been tested,
// but I use it for easier-to-read high-level tests.
func TestExpr(t *testing.T) {
	cases := []struct {
		TermInput string
		EvalInput interface{}
		Opts      Opt
		WantResp  interface{}
		WantErr   error
	}{
		/*
			{`Mom/Name`, expr_eval_input_0, Opt{}, `Ana Belle`, nil},
			{`(Mom/Name) == Ana`, expr_eval_input_1, Opt{}, true, nil},
			// Make sure quotes are removed
			{`Name == "Ana Belle"`, expr_eval_input_2, Opt{}, true, nil},
			// Test strictness -- by default strict is off, and incompatibile comparisons result in false.
			{`Name == 22`, expr_eval_input_3, Opt{Strict: false}, false, nil},
			// Test strictness -- if strict is on, report error with incompatible comparisons.
			{`Name == 22`, expr_eval_input_3, Opt{Strict: true}, false, mismatchErr},
			// Test int evalation, equal and not equal.
			{`Age == 22`, expr_eval_input_4, Opt{}, true, nil},
			{`Age != 22`, expr_eval_input_4, Opt{}, false, nil},
			// Test compound comparisons.
			{`(Name == "Ana") && (Age == 22)`, expr_eval_input_5, Opt{}, true, nil},
			{`(Name == "Ana") || (Age == 23)`, expr_eval_input_5, Opt{}, true, nil},
			{`(Name == "Mana") || (Age == 22)`, expr_eval_input_5, Opt{}, true, nil},
			{`(Name == "Mana") || (Age == 23)`, expr_eval_input_5, Opt{}, false, nil},
		*/
	}
	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			expr, err := MakeExpr(tc.TermInput)
			if err != nil {
				panic(err)
			}
			have_resp, have_err := expr.Eval(tc.EvalInput, &tc.Opts)
			if !errorMatches(have_err, tc.WantErr) {
				fmt.Println("Error mismatch, have\n", have_err, "\nwant\n", tc.WantErr)
				t.Fatal()
			} else if !interfaceMatches(have_resp, tc.WantResp) {
				fmt.Println("Response mismatch, have\n", toJsonString(have_resp), "\nwant\n", toJsonString(tc.WantResp))
				t.Fatal()
			}
		})
	}
}

var (
	expr_eval_input_0 = &Person{Mom: Relative{Name: "Ana Belle"}}
	expr_eval_input_1 = &Person{Mom: Relative{Name: "Ana"}}
	expr_eval_input_2 = &Person{Name: "Ana Belle"}
	expr_eval_input_3 = &Person{Name: "Ana"}
	expr_eval_input_4 = &Person{Age: 22}
	expr_eval_input_5 = &Person{Name: "Ana", Age: 22}
)

// ------------------------------------------------------------
// MODEL

type Person struct {
	Name     string    `json:"Name,omitempty"`     // Test a single value string
	Age      int       `json:"Age,omitempty"`      // Test a single value int
	Mom      Relative  `json:"Mom,omitempty"`      // Test a value field
	Children []Person  `json:"Children,omitempty"` // Test a value collection
	Friends  []*Person `json:"Friends,omitempty"`  // Test a pointer collection
}

type Relative struct {
	Name string `json:"Name,omitempty"`
}

// ------------------------------------------------------------
// MODEL BOILERPLATE

// MarshalJSON() is only necessary because go randomizes the fields.
func (p *Person) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer

	type pair_t struct {
		Key   interface{}
		Value interface{}
	}
	// XXX I'm not actually using the json metadata names but I should be.
	ordered := []pair_t{}
	if p.Name != "" {
		ordered = append(ordered, pair_t{"Name", p.Name})
	}
	if p.Age != 0 {
		ordered = append(ordered, pair_t{"Age", p.Age})
	}
	if !p.Mom.Empty() {
		ordered = append(ordered, pair_t{"Mom", p.Mom})
	}
	if len(p.Children) > 0 {
		ordered = append(ordered, pair_t{"Children", p.Children})
	}
	if len(p.Friends) > 0 {
		ordered = append(ordered, pair_t{"Friends", p.Friends})
	}

	buf.WriteString("{")
	for i, kv := range ordered {
		if i != 0 {
			buf.WriteString(",")
		}
		// marshal key
		key, err := json.Marshal(kv.Key)
		if err != nil {
			return nil, err
		}
		buf.Write(key)
		buf.WriteString(":")
		// marshal value
		val, err := json.Marshal(kv.Value)
		if err != nil {
			return nil, err
		}
		buf.Write(val)
	}

	buf.WriteString("}")
	return buf.Bytes(), nil
}

func (r Relative) Empty() bool {
	return r.Name == ""
}

// ------------------------------------------------------------
// BUILD (tokens)

// tokens() constructs a flat list of tokens
func tokens(all ...interface{}) []*node_t {
	var tokens []*node_t
	for _, t := range all {
		switch v := t.(type) {
		case float32:
			val := strconv.FormatFloat(float64(v), 'f', 8, 64)
			tokens = append(tokens, newToken(float_token, val))
		case float64:
			val := strconv.FormatFloat(v, 'f', 8, 64)
			tokens = append(tokens, newToken(float_token, val))
		case int:
			tokens = append(tokens, newToken(int_token, strconv.Itoa(v)))
		case string:
			tokens = append(tokens, newToken(string_token, v).reclassify())
		}
	}
	return tokens
}

// mk_binary() constructs a binary token from the symbol
func mk_binary(sym symbol, left, right *node_t) *node_t {
	b := newToken(sym, "")
	b.addChild(left)
	b.addChild(right)
	return b
}

// ------------------------------------------------------------
// BUILD (ast)

func and_n(lhs, rhs interface{}) AstNode {
	return &binaryNode{Op: and_token, Lhs: wrap_field_n(lhs), Rhs: wrap_field_n(rhs)}
}

func cnd_n(child AstNode) AstNode {
	return &conditionNode{Op: condition_token, Child: child}
}

func eql_n(lhs, rhs interface{}) AstNode {
	return &binaryNode{Op: eql_token, Lhs: wrap_field_n(lhs), Rhs: wrap_string_n(rhs)}
}

func field_n(name string) AstNode {
	return &fieldNode{Field: name}
}

func float_n(value float64) AstNode {
	return &constantNode{Value: value}
}

func int_n(value int) AstNode {
	return &constantNode{Value: value}
}

func or_n(lhs, rhs interface{}) AstNode {
	return &binaryNode{Op: or_token, Lhs: wrap_field_n(lhs), Rhs: wrap_field_n(rhs)}
}

func paren_n(child AstNode) AstNode {
	return &unaryNode{Op: open_token, Child: child}
}

func path_n(lhs, rhs AstNode) AstNode {
	return &pathNode{Lhs: lhs, Rhs: rhs}
}

func string_n(value string) AstNode {
	return &constantNode{Value: value}
}

func wrap_field_n(a interface{}) AstNode {
	switch t := a.(type) {
	case string:
		return field_n(t)
	case AstNode:
		return t
	default:
		panic(a)
	}
}

func wrap_string_n(a interface{}) AstNode {
	switch t := a.(type) {
	case string:
		return string_n(t)
	case AstNode:
		return t
	default:
		panic(a)
	}
}

// ------------------------------------------------------------
// COMPARE

func errorMatches(a, b error) bool {
	if a == nil && b == nil {
		return true
	} else if a == nil || b == nil {
		return false
	}
	// Internal error class only needs to match to the type
	aerr, aok := a.(*sqi_err_t)
	berr, bok := b.(*sqi_err_t)
	if aok && bok {
		return aerr.code == berr.code
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

func tokensMatch(a, b []*node_t) bool {
	if len(a) != len(b) {
		return false
	} else if len(a) < 1 {
		return true
	}
	for i, t := range a {
		if !tokenMatches(t, b[i]) {
			return false
		}
	}
	return true
}

// tokenMatches() compares only the lexer portion of the token, not the parsing.
func tokenMatches(a, b *node_t) bool {
	if a == nil && b == nil {
		return true
	} else if a == nil || b == nil {
		return false
	}
	return a.Token.Symbol == b.Token.Symbol && a.Text == b.Text
}

// ------------------------------------------------------------
// COMPARE BOILERPLATE

func (n *node_t) MarshalJSON() ([]byte, error) {
	return orderedMarshalJSON(n)
}

func (t *token_t) MarshalJSON() ([]byte, error) {
	return orderedMarshalJSON(t)
}

// orderedMarshalJSON() is a generic function for writing ordered json data.
func orderedMarshalJSON(src interface{}) ([]byte, error) {
	var buf bytes.Buffer
	type pair_t struct {
		Key   interface{}
		Value interface{}
	}

	v := reflect.Indirect(reflect.ValueOf(src))
	var pairs []pair_t
	for i := 0; i < v.NumField(); i++ {
		val := v.Field(i)
		field := v.Type().Field(i)
		tag := field.Tag.Get("json")
		if tag != "-" && len(field.Name) > 0 && field.Name[0] == strings.ToUpper(field.Name)[0] {
			pairs = append(pairs, pair_t{field.Name, val.Interface()})
		}
	}

	buf.WriteString("{")
	for i, kv := range pairs {
		if i != 0 {
			buf.WriteString(",")
		}
		// marshal key
		key, err := json.Marshal(kv.Key)
		if err != nil {
			return nil, err
		}
		buf.Write(key)
		buf.WriteString(":")
		// marshal value
		val, err := json.Marshal(kv.Value)
		if err != nil {
			return nil, err
		}
		buf.Write(val)
	}

	buf.WriteString("}")
	return buf.Bytes(), nil
}

// ------------------------------------------------------------
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
