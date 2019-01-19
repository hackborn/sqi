package sqi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"
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
	return

	cases := []struct {
		Input    []*node_t
		WantResp *node_t
		WantErr  error
	}{
		{tokens(`a`), parser_want_0, nil},
		{tokens(`/`, `a`, `/`, `b`), parser_want_1, nil},
		{tokens(`/`, `a`, `/`, `b`, `/`, `c`), parser_want_2, nil},
		{tokens(`/`, `a`, `/`, `b`, `==`, `c`), parser_want_3, nil},
		{tokens(`(`, `a`, `)`), parser_want_4, nil},
		{tokens(`(`, `/`, `a`, `/`, `b`, `)`, `==`, `c`), parser_want_6, nil},
		{tokens(`/`, `a`, `/`, `b`, `==`, 10), parser_want_7, nil},
		{tokens(`/`, `a`, `/`, `b`, `==`, 5.5), parser_want_8, nil},
		{tokens(`a`, `==`, `b`, `||`, `c`, `==`, `d`), parser_want_9, nil},
		{tokens(`(`, `a`, `==`, `b`, `)`, `||`, `(`, `c`, `==`, `d`, `)`), parser_want_9, nil},
		{tokens(`/`, `a`, `==`, `/`, `b`, `||`, `/`, `c`, `==`, `/`, `d`), parser_want_10, nil},
		{tokens(`(`, `/`, `a`, `==`, `/`, `b`, `)`, `||`, `(`, `/`, `c`, `==`, `/`, `d`, `)`), parser_want_10, nil},
	}
	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			have_resp, have_err := parse(tc.Input)
			if !errorMatches(have_err, tc.WantErr) {
				fmt.Println("Error mismatch, have\n", have_err, "\nwant\n", tc.WantErr)
				t.Fatal()
			} else if !interfaceMatches(have_resp, tc.WantResp) {
				fmt.Println("Parser mismatch, have\n", toJsonString(have_resp), "\nwant\n", toJsonString(tc.WantResp))
				t.Fatal()
			}
		})
	}
}

var (
	parser_want_0  = str_n(`a`)
	parser_want_1  = path_n(path_n(str_n(`b`), nil), str_n(`a`))
	parser_want_2  = path_n(path_n(path_n(str_n(`a`), nil), str_n(`b`)), str_n(`c`))
	parser_want_3  = eql_n(path_n(path_n(str_n(`a`), nil), str_n(`b`)), str_n(`c`))
	parser_want_4  = str_n(`a`)
	parser_want_6  = eql_n(path_n(path_n(str_n(`b`), nil), str_n(`a`)), str_n(`c`))
	parser_want_7  = eql_n(path_n(path_n(str_n(`b`), nil), str_n(`a`)), int_n(10))
	parser_want_8  = eql_n(path_n(path_n(str_n(`b`), nil), str_n(`a`)), float_n(5.5))
	parser_want_9  = or_n(eql_n(str_n(`a`), str_n(`b`)), eql_n(str_n(`c`), str_n(`d`)))
	parser_want_10 = or_n(eql_n(path_n(str_n(`a`), nil), path_n(str_n(`b`), nil)), eql_n(path_n(str_n(`c`), nil), path_n(str_n(`d`), nil)))
)

// ------------------------------------------------------------
// TEST-CONTEXTUALIZER

func TestContextualizer(t *testing.T) {
	return

	cases := []struct {
		Input    *node_t
		WantResp *node_t
		WantErr  error
	}{
		// A string
		{ctx_input_0, ctx_want_0, nil},
		// A field
		{ctx_input_1, ctx_want_1, nil},
		// A conditional and field/string
		{ctx_input_2, ctx_want_2, nil},
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
	ctx_input_0 = str_n(`a`)
	ctx_want_0  = str_n(`a`)

	ctx_input_1 = path_n(str_n(`a`), str_n(`b`))
	ctx_want_1  = path_n(str_n(`a`), str_n(`b`))

	ctx_input_2 = eql_n(str_n(`a`), str_n(`b`))
	ctx_want_2  = mk_unary(condition_token, eql_n(str_n(`a`), str_n(`b`)))
)

// ------------------------------------------------------------
// TEST-AST-GET

// I can't decide if I want these tests -- all the pieces and
// the full system is tested, and this is feeling like more
// overhead than it's worth.

/*
func TestAstGet(t *testing.T) {
	cases := []struct {
		Input    interface{}
		Expr     AstNode
		Opts     Opt
		WantResp interface{}
		WantErr  error
	}{
		{ast_get_input_0, ast_get_expr_0, Opt{}, "a", nil},
		{ast_get_input_1, ast_get_expr_1, Opt{}, Relative{Name: "ca"}, nil},
		{ast_get_input_2, ast_get_expr_2, Opt{}, []Person{Person{Name: "ca"}, Person{Name: "cb"}}, nil},
		{ast_get_input_3, ast_get_expr_3, Opt{}, []Person{Person{Name: "cb"}}, nil},
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
	//	ast_get_expr_3  = path_n(field_n("Children"), cnd_n(eql_n(field_n("Name"), string_n("cb"))))
	ast_get_expr_3 = eql_n(field_n("Name"), string_n("cb"))

	children_node    = field_n("Children")
	get_name_cb_node = eql_n(field_n("Name"), string_n("cb"))
)
*/

// ------------------------------------------------------------
// TEST-EXPR

func TestExpr(t *testing.T) {
	cases := []struct {
		TermInput string
		EvalInput interface{}
		Opts      Opt
		WantResp  interface{}
		WantErr   error
	}{
		{`/Mom/Name`, expr_eval_input_0, Opt{}, `Ana Belle`, nil},
		{`/Mom/Name == Ana`, expr_eval_input_1, Opt{}, true, nil},
		{`(/Mom/Name) == Ana`, expr_eval_input_1, Opt{}, true, nil},
		// Make sure quotes are removed
		{`/Name == "Ana Belle"`, expr_eval_input_2, Opt{}, true, nil},
		// Test strictness -- by default strict is off, and incompatibile comparisons result in false.
		{`/Name == 22`, expr_eval_input_3, Opt{Strict: false}, false, nil},
		// Test strictness -- if strict is on, report error with incompatible comparisons.
		{`/Name == 22`, expr_eval_input_3, Opt{Strict: true}, false, mismatchErr},
		// Test int evalation, equal and not equal.
		{`/Age == 22`, expr_eval_input_4, Opt{}, true, nil},
		{`/Age != 22`, expr_eval_input_4, Opt{}, false, nil},
		// Test compound comparisons.
		{`/Name == "Ana" && /Age == 22`, expr_eval_input_5, Opt{}, true, nil},
		{`(/Name == "Ana") && (/Age == 22)`, expr_eval_input_5, Opt{}, true, nil},
		{`/Name == "Ana" || /Age == 23`, expr_eval_input_5, Opt{}, true, nil},
		{`(/Name == "Ana") || (/Age == 23)`, expr_eval_input_5, Opt{}, true, nil},
		{`(/Name == "Mana") || (/Age == 22)`, expr_eval_input_5, Opt{}, true, nil},
		{`(/Name == "Mana") || (/Age == 23)`, expr_eval_input_5, Opt{}, false, nil},
		// Test path equality
		{`/Mom/Name == /Mom/Name`, expr_eval_input_1, Opt{}, true, nil},
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
				tokens, _ := scan(tc.TermInput)
				tree, _ := parse(tokens)
				tree, _ = contextualize(tree)
				fmt.Println("tree is\n", toJsonString(tree))
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
			tokens = append(tokens, newNode(float_token, val))
		case float64:
			val := strconv.FormatFloat(v, 'f', 8, 64)
			tokens = append(tokens, newNode(float_token, val))
		case int:
			tokens = append(tokens, newNode(int_token, strconv.Itoa(v)))
		case string:
			tokens = append(tokens, newNode(string_token, v).reclassify())
		}
	}
	return tokens
}

// bin_n() constructs a binary token from the symbol
func bin_n(sym symbol, left, right *node_t) *node_t {
	b := newNode(sym, "")
	b.addChild(left)
	b.addChild(right)
	return b
}

func eql_n(left, right *node_t) *node_t {
	return bin_n(eql_token, left, right)
}

func float_n(v float64) *node_t {
	text := strconv.FormatFloat(v, 'f', 6, 64)
	return newNode(int_token, text)
}

func int_n(v int) *node_t {
	return newNode(int_token, strconv.Itoa(v))
}

func or_n(left, right *node_t) *node_t {
	return bin_n(or_token, left, right)
}

func path_n(left, right *node_t) *node_t {
	b := newNode(path_token, "")
	b.addChild(left)
	if right != nil {
		b.addChild(right)
	}
	return b
}

func str_n(text string) *node_t {
	return newNode(string_token, text)
}

// mk_unary() constructs a unary token from the symbol
func mk_unary(sym symbol, child *node_t) *node_t {
	n := newNode(sym, "")
	n.addChild(child)
	return n
}

// ------------------------------------------------------------
// BUILD (ast)

func and_a(lhs, rhs interface{}) AstNode {
	return &binaryNode{Op: and_token, Lhs: wrap_field_a(lhs), Rhs: wrap_field_a(rhs)}
}

func cnd_a(child AstNode) AstNode {
	return &conditionNode{Op: condition_token, Child: child}
}

func eql_a(lhs, rhs interface{}) AstNode {
	return &binaryNode{Op: eql_token, Lhs: wrap_field_a(lhs), Rhs: wrap_field_a(rhs)}
}

func field_a(name string) AstNode {
	return &fieldNode{Field: name}
}

func float_a(value float64) AstNode {
	return &constantNode{Value: value}
}

func int_a(value int) AstNode {
	return &constantNode{Value: value}
}

func or_a(lhs, rhs interface{}) AstNode {
	return &binaryNode{Op: or_token, Lhs: wrap_field_a(lhs), Rhs: wrap_field_a(rhs)}
}

func paren_a(child AstNode) AstNode {
	return &unaryNode{Op: open_token, Child: child}
}

func path_a(child, field AstNode) AstNode {
	return &pathNode{Child: child, Field: field}
}

func string_a(value string) AstNode {
	return &constantNode{Value: value}
}

func wrap_field_a(a interface{}) AstNode {
	switch t := a.(type) {
	case string:
		return field_a(t)
	case AstNode:
		return t
	default:
		panic(a)
	}
}

func wrap_string_n(a interface{}) AstNode {
	switch t := a.(type) {
	case string:
		return string_a(t)
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
