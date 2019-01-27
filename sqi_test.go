package sqi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
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
		{`( [1] ) == "b"`, tokens(`(`, `[`, 1, `]`, `)`, `==`, `"b"`), nil},
		{`([1]) == "b"`, tokens(`(`, `[`, 1, `]`, `)`, `==`, `"b"`), nil},
		{`/a[0]`, tokens(`/`, `a`, `[`, 0, `]`), nil},
		{`/a[0]/b`, tokens(`/`, `a`, `[`, 0, `]`, `/`, `b`), nil},
		//		{`/a[ -1]`, tokens(`/`, `a`, `[`, 0, `]`), nil},
	}
	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			have_resp, have_err := scan(tc.Input)
			if !errorMatches(have_err, tc.WantErr) {
				fmt.Println("Error mismatch, have\n", have_err, "\nwant\n", tc.WantErr)
				t.Fatal()
			} else if !tokensMatch(have_resp, tc.WantResp) {
				fmt.Println("Token mismatch, have\n", toJsonString(have_resp), "\nwant\n", toJsonString(tc.WantResp))
				t.Fatal()
			}
		})
	}
}

// ------------------------------------------------------------
// TEST-PARSER

func TestParser(t *testing.T) {
	parser_want_0 := str_n(`a`)
	parser_want_1 := path_n(path_n(str_n(`a`), nil), str_n(`b`))
	parser_want_2 := path_n(path_n(path_n(str_n(`a`), nil), str_n(`b`)), str_n(`c`))
	parser_want_3 := eql_n(path_n(path_n(str_n(`a`), nil), str_n(`b`)), str_n(`c`))
	parser_want_4 := str_n(`a`)
	parser_want_6 := eql_n(path_n(path_n(str_n(`a`), nil), str_n(`b`)), str_n(`c`))
	parser_want_7 := eql_n(path_n(path_n(str_n(`a`), nil), str_n(`b`)), int_n(10))
	parser_want_8 := eql_n(path_n(path_n(str_n(`a`), nil), str_n(`b`)), float_n(5.5))
	parser_want_9 := or_n(eql_n(str_n(`a`), str_n(`b`)), eql_n(str_n(`c`), str_n(`d`)))
	parser_want_10 := or_n(eql_n(path_n(str_n(`a`), nil), path_n(str_n(`b`), nil)), eql_n(path_n(str_n(`c`), nil), path_n(str_n(`d`), nil)))
	parser_want_11 := array_n(str_n(`a`), int_n(0))
	parser_want_12 := array_n(path_n(str_n(`a`), nil), int_n(0))
	parser_want_13 := path_n(array_n(path_n(str_n(`a`), nil), int_n(0)), str_n(`b`))
	parser_want_14 := array_n(path_n(path_n(str_n(`a`), nil), str_n(`b`)), int_n(0))
	parser_want_15 := path_n(path_n(str_n(`a`), nil), eql_n(path_n(str_n(`b`), nil), str_n(`c`)))

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
		{tokens(`a`, `[`, 0, `]`), parser_want_11, nil},
		{tokens(`/`, `a`, `[`, 0, `]`), parser_want_12, nil},
		{tokens(`/`, `a`, `[`, 0, `]`, `/`, `b`), parser_want_13, nil},
		{tokens(`/`, `a`, `/`, `b`, `[`, 0, `]`), parser_want_14, nil},
		{tokens(`/`, `a`, `/`, `(`, `/`, `b`, `==`, `c`, `)`), parser_want_15, nil},
		// Errors
		{tokens(`(`, `a`, `[`, 0, `]`), nil, parseErr},
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

// ------------------------------------------------------------
// TEST-CONTEXTUALIZER

func TestContextualizer(t *testing.T) {
	ctx_input_0 := path_n(path_n(str_n(`a`), nil), eql_n(path_n(str_n(`b`), nil), str_n(`c`)))
	ctx_want_0 := path_n(path_n(str_n(`a`), nil), sel_n(eql_n(path_n(str_n(`b`), nil), str_n(`c`))))

	cases := []struct {
		Input    *node_t
		WantResp *node_t
		WantErr  error
	}{
		// A select
		{ctx_input_0, ctx_want_0, nil},
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

// ------------------------------------------------------------
// TEST-EXPR

func TestExpr(t *testing.T) {
	input_0 := &Person{Mom: Relative{Name: "Ana Belle"}}
	input_1 := &Person{Mom: Relative{Name: "Ana"}}
	input_2 := &Person{Name: "Ana Belle"}
	input_3 := &Person{Name: "Ana"}
	input_4 := &Person{Age: 22}
	input_5 := &Person{Name: "Ana", Age: 22}
	input_6 := &Person{Children: []Person{Person{Name: "a"}, Person{Name: "b"}, Person{Name: "c"}}}

	cases := []struct {
		ExprInput string
		EvalInput interface{}
		Opts      Opt
		WantResp  interface{}
		WantErr   error
	}{
		{`/Mom/Name`, input_0, Opt{}, `Ana Belle`, nil},
		// Accommodate a special syntax that will be necessary for path queries.
		{`/Mom/(/Name)`, input_0, Opt{}, `Ana Belle`, nil},
		{`/Mom/Name == Ana`, input_1, Opt{}, true, nil},
		{`(/Mom/Name) == Ana`, input_1, Opt{}, true, nil},
		// Make sure quotes are removed
		{`/Name == "Ana Belle"`, input_2, Opt{}, true, nil},
		// Strictness -- by default strict is off, and incompatibile comparisons result in false.
		{`/Name == 22`, input_3, Opt{Strict: false}, false, nil},
		// Strictness -- if strict is on, report error with incompatible comparisons.
		{`/Name == 22`, input_3, Opt{Strict: true}, false, mismatchErr},
		// Int evalation, equal and not equal.
		{`/Age == 22`, input_4, Opt{}, true, nil},
		{`/Age != 22`, input_4, Opt{}, false, nil},
		// Compound comparisons.
		{`/Name == "Ana" && /Age == 22`, input_5, Opt{}, true, nil},
		{`(/Name == "Ana") && (/Age == 22)`, input_5, Opt{}, true, nil},
		{`/Name == "Ana" || /Age == 23`, input_5, Opt{}, true, nil},
		{`(/Name == "Ana") || (/Age == 23)`, input_5, Opt{}, true, nil},
		{`(/Name == "Mana") || (/Age == 22)`, input_5, Opt{}, true, nil},
		{`(/Name == "Mana") || (/Age == 23)`, input_5, Opt{}, false, nil},
		// Path equality
		{`/Mom/Name == /Mom/Name`, input_1, Opt{}, true, nil},
		// Select
		{`/Children/(/Name == "c")`, input_6, Opt{}, []Person{Person{Name: "c"}}, nil},
		// Select, unwinding the results to a single item
		{`(/Children/(/Name == "c"))[0]`, input_6, Opt{}, Person{Name: "c"}, nil},
		// Arrays
		{`/Children[0]`, input_6, Opt{}, Person{Name: "a"}, nil},
		{`/Children[1]`, input_6, Opt{}, Person{Name: "b"}, nil},
		{`/Children[1]/Name`, input_6, Opt{}, "b", nil},
		{`[1]`, [2]string{"a", "b"}, Opt{}, "b", nil},
		{`([1]) == "b"`, [2]string{"a", "b"}, Opt{}, true, nil},
		{`/Children[0]`, input_3, Opt{}, nil, nil},
	}
	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			runTestExpr(t, tc.ExprInput, tc.EvalInput, tc.Opts, tc.WantResp, tc.WantErr)
			// Everything that works on the struct should work on the unmarshalled json.
			runTestExpr(t, tc.ExprInput, toJson(tc.EvalInput), tc.Opts, tc.WantResp, tc.WantErr)
		})
	}
}

func runTestExpr(t *testing.T, exprinput string, evalinput interface{}, opt Opt, want_resp interface{}, want_err error) {
	expr, err := MakeExpr(exprinput)
	if err != nil {
		fmt.Println("make expr failed", err)
		printExprConstruction(exprinput)
		t.Fatal()
	}
	have_resp, have_err := expr.Eval(evalinput, &opt)
	if !errorMatches(have_err, want_err) {
		fmt.Println("Error mismatch, have\n", have_err, "\nwant\n", want_err)
		t.Fatal()
	} else if !interfaceMatches(have_resp, want_resp) {
		fmt.Println("Response mismatch, have\n", toJsonString(have_resp), "\nwant\n", toJsonString(want_resp))
		fmt.Println("Response mismatch, have\n", have_resp, "\nwant\n", want_resp)
		//		printExprConstruction(exprinput)
		t.Fatal()
	}
}

// ------------------------------------------------------------
// TEST-EVAL-FLOAT64

func TestEvalFloat64(t *testing.T) {
	type Metric struct {
		Level float64
	}

	input_0 := &Metric{Level: 46.8}

	cases := []struct {
		Term     string
		Input    interface{}
		Opt      *Opt
		WantResp float64
	}{
		{`/Level`, input_0, nil, 46.8},
		{`/NoLevel`, input_0, nil, 0},
		{`/ErrorLevel`, input_0, &Opt{OnError: 10.0}, 10.0},
	}
	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			have_resp := EvalFloat64(tc.Term, tc.Input, tc.Opt)
			if have_resp != tc.WantResp {
				fmt.Println("Response mismatch, have\n", have_resp, "\nwant\n", tc.WantResp)
				printExprConstruction(tc.Term)
				t.Fatal()
			}
		})
	}
}

// ------------------------------------------------------------
// TEST-EVAL-INT

func TestEvalInt(t *testing.T) {
	input_0 := &Person{Age: 32}

	cases := []struct {
		Term     string
		Input    interface{}
		Opt      *Opt
		WantResp int
	}{
		{`/Age`, input_0, nil, 32},
		{`/NoAge`, input_0, nil, 0},
		{`/ErrorAge`, input_0, &Opt{OnError: 10}, 10},
	}
	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			have_resp := EvalInt(tc.Term, tc.Input, tc.Opt)
			if have_resp != tc.WantResp {
				fmt.Println("Response mismatch, have\n", strconv.Itoa(have_resp), "\nwant\n", strconv.Itoa(tc.WantResp))
				printExprConstruction(tc.Term)
				t.Fatal()
			}
		})
	}
}

// ------------------------------------------------------------
// TEST-EVAL-STRING

func TestEvalString(t *testing.T) {
	input_0 := &Person{Name: "Ana"}

	cases := []struct {
		Term     string
		Input    interface{}
		Opt      *Opt
		WantResp string
	}{
		{`/Name`, input_0, nil, `Ana`},
		{`/NoName`, input_0, nil, ``},
		{`/ErrorName`, input_0, &Opt{OnError: `zip`}, `zip`},
	}
	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			have_resp := EvalString(tc.Term, tc.Input, tc.Opt)
			if have_resp != tc.WantResp {
				fmt.Println("Response mismatch, have\n", have_resp, "\nwant\n", tc.WantResp)
				printExprConstruction(tc.Term)
				t.Fatal()
			}
		})
	}
}

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
func (p Person) MarshalJSON() ([]byte, error) {
	return orderedMarshalJSON(p)
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

func array_n(left, right *node_t) *node_t {
	return bin_n(open_array, left, right)
}

// bin_n() constructs a binary token from the symbol
func bin_n(sym symbol, left, right *node_t) *node_t {
	b := newNode(sym, token_map[sym].Text)
	b.addChild(left)
	b.addChild(right)
	return b
}

func eql_n(left, right *node_t) *node_t {
	return bin_n(eql_token, left, right)
}

func float_n(v float64) *node_t {
	text := strconv.FormatFloat(v, 'f', 8, 64)
	return newNode(float_token, text)
}

func int_n(v int) *node_t {
	return newNode(int_token, strconv.Itoa(v))
}

func or_n(left, right *node_t) *node_t {
	return bin_n(or_token, left, right)
}

func path_n(left, right *node_t) *node_t {
	b := newNode(path_token, token_map[path_token].Text)
	b.addChild(left)
	if right != nil {
		b.addChild(right)
	}
	return b
}

func sel_n(child *node_t) *node_t {
	return mk_unary(select_token, child)
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
// REPORT

func printExprConstruction(exprinput string) {
	tokens, _ := scan(exprinput)
	fmt.Println("after lexing\n", toJsonString(tokens))
	tree, _ := parse(tokens)
	fmt.Println("after parsing\n", toJsonString(tree))
	tree, _ = contextualize(tree)
	fmt.Println("after contextualizing\n", toJsonString(tree))
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

func (n node_t) MarshalJSON() ([]byte, error) {
	return orderedMarshalJSON(n)
}

func (t token_t) MarshalJSON() ([]byte, error) {
	return orderedMarshalJSON(t)
}

// orderedMarshalJSON() is a generic function for writing ordered json data.
func orderedMarshalJSON(src interface{}) ([]byte, error) {
	var buf bytes.Buffer
	v := reflect.Indirect(reflect.ValueOf(src))
	var pairs []pair_t
	for i := 0; i < v.NumField(); i++ {
		val := v.Field(i)
		field := v.Type().Field(i)
		if wantsJsonField(val, field) {
			pairs = append(pairs, pair_t{field.Name, val.Interface()})
		}
	}

	sort.Sort(ByPair(pairs))

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

func wantsJsonField(val reflect.Value, field reflect.StructField) bool {
	tag := field.Tag.Get("json")
	if tag == "-" {
		return false
	}
	// Don't try and encode unexported fields. What's the official way to determine this?
	if len(field.Name) < 1 || field.Name[0] != strings.ToUpper(field.Name)[0] {
		return false
	}
	if tag == "omitempty" {
		x := val.Interface()
		return !(reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface()))
	}
	return true
}

type pair_t struct {
	Key   string
	Value interface{}
}

type ByPair []pair_t

func (a ByPair) Len() int           { return len(a) }
func (a ByPair) Less(i, j int) bool { return strings.Compare(a[i].Key, a[j].Key) < 0 }
func (a ByPair) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

// ------------------------------------------------------------
// CONVERT

func toJson(i interface{}) interface{} {
	pbytes, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}

	rt := reflect.TypeOf(i)
	switch rt.Kind() {
	case reflect.Slice, reflect.Array:
		i2 := make([]interface{}, 0, 0)
		err = json.Unmarshal(pbytes, &i2)
		if err != nil {
			panic(err)
		}
		return i2
	default:
		i2 := make(map[string]interface{})
		err = json.Unmarshal(pbytes, &i2)
		if err != nil {
			panic(err)
		}
		return i2
	}
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
