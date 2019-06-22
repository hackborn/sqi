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
		WantResp []*nodeT
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
			haveResp, haveErr := scan(tc.Input)
			if !errorMatches(haveErr, tc.WantErr) {
				fmt.Println("Error mismatch, have\n", haveErr, "\nwant\n", tc.WantErr)
				t.Fatal()
			} else if !tokensMatch(haveResp, tc.WantResp) {
				fmt.Println("Token mismatch, have\n", toJsonString(haveResp), "\nwant\n", toJsonString(tc.WantResp))
				t.Fatal()
			}
		})
	}
}

// ------------------------------------------------------------
// TEST-PARSER

func TestParser(t *testing.T) {
	want0 := strN(`a`)
	want1 := pathN(pathN(strN(`a`), nil), strN(`b`))
	want2 := pathN(pathN(pathN(strN(`a`), nil), strN(`b`)), strN(`c`))
	want3 := eqlN(pathN(pathN(strN(`a`), nil), strN(`b`)), strN(`c`))
	want4 := strN(`a`)
	want6 := eqlN(pathN(pathN(strN(`a`), nil), strN(`b`)), strN(`c`))
	want7 := eqlN(pathN(pathN(strN(`a`), nil), strN(`b`)), intN(10))
	want8 := eqlN(pathN(pathN(strN(`a`), nil), strN(`b`)), floatN(5.5))
	want9 := orN(eqlN(strN(`a`), strN(`b`)), eqlN(strN(`c`), strN(`d`)))
	want10 := orN(eqlN(pathN(strN(`a`), nil), pathN(strN(`b`), nil)), eqlN(pathN(strN(`c`), nil), pathN(strN(`d`), nil)))
	want11 := arrayN(strN(`a`), intN(0))
	want12 := arrayN(pathN(strN(`a`), nil), intN(0))
	want13 := pathN(arrayN(pathN(strN(`a`), nil), intN(0)), strN(`b`))
	want14 := arrayN(pathN(pathN(strN(`a`), nil), strN(`b`)), intN(0))
	want15 := pathN(pathN(strN(`a`), nil), eqlN(pathN(strN(`b`), nil), strN(`c`)))

	cases := []struct {
		Input    []*nodeT
		WantResp *nodeT
		WantErr  error
	}{
		{tokens(`a`), want0, nil},
		{tokens(`/`, `a`, `/`, `b`), want1, nil},
		{tokens(`/`, `a`, `/`, `b`, `/`, `c`), want2, nil},
		{tokens(`/`, `a`, `/`, `b`, `==`, `c`), want3, nil},
		{tokens(`(`, `a`, `)`), want4, nil},
		{tokens(`(`, `/`, `a`, `/`, `b`, `)`, `==`, `c`), want6, nil},
		{tokens(`/`, `a`, `/`, `b`, `==`, 10), want7, nil},
		{tokens(`/`, `a`, `/`, `b`, `==`, 5.5), want8, nil},
		{tokens(`a`, `==`, `b`, `||`, `c`, `==`, `d`), want9, nil},
		{tokens(`(`, `a`, `==`, `b`, `)`, `||`, `(`, `c`, `==`, `d`, `)`), want9, nil},
		{tokens(`/`, `a`, `==`, `/`, `b`, `||`, `/`, `c`, `==`, `/`, `d`), want10, nil},
		{tokens(`(`, `/`, `a`, `==`, `/`, `b`, `)`, `||`, `(`, `/`, `c`, `==`, `/`, `d`, `)`), want10, nil},
		{tokens(`a`, `[`, 0, `]`), want11, nil},
		{tokens(`/`, `a`, `[`, 0, `]`), want12, nil},
		{tokens(`/`, `a`, `[`, 0, `]`, `/`, `b`), want13, nil},
		{tokens(`/`, `a`, `/`, `b`, `[`, 0, `]`), want14, nil},
		{tokens(`/`, `a`, `/`, `(`, `/`, `b`, `==`, `c`, `)`), want15, nil},
		// Errors
		{tokens(`(`, `a`, `[`, 0, `]`), nil, parseErr},
	}
	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			haveResp, haveErr := parse(tc.Input)
			if !errorMatches(haveErr, tc.WantErr) {
				fmt.Println("Error mismatch, have\n", haveErr, "\nwant\n", tc.WantErr)
				t.Fatal()
			} else if !interfaceMatches(haveResp, tc.WantResp) {
				fmt.Println("Parser mismatch, have\n", toJsonString(haveResp), "\nwant\n", toJsonString(tc.WantResp))
				t.Fatal()
			}
		})
	}
}

// ------------------------------------------------------------
// TEST-CONTEXTUALIZER

func TestContextualizer(t *testing.T) {
	input0 := pathN(pathN(strN(`a`), nil), eqlN(pathN(strN(`b`), nil), strN(`c`)))
	want0 := pathN(pathN(strN(`a`), nil), selN(eqlN(pathN(strN(`b`), nil), strN(`c`))))

	cases := []struct {
		Input    *nodeT
		WantResp *nodeT
		WantErr  error
	}{
		// A select
		{input0, want0, nil},
	}
	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			haveResp, haveErr := contextualize(tc.Input)
			if !errorMatches(haveErr, tc.WantErr) {
				fmt.Println("Error mismatch, have\n", haveErr, "\nwant\n", tc.WantErr)
				t.Fatal()
			} else if !interfaceMatches(haveResp, tc.WantResp) {
				fmt.Println("Token mismatch, have\n", toJsonString(haveResp), "\nwant\n", toJsonString(tc.WantResp))
				t.Fatal()
			}
		})
	}
}

// ------------------------------------------------------------
// TEST-EXPR

func TestExpr(t *testing.T) {
	input0 := &Person{Mom: Relative{Name: "Ana Belle"}}
	input1 := &Person{Mom: Relative{Name: "Ana"}}
	input2 := &Person{Name: "Ana Belle"}
	input3 := &Person{Name: "Ana"}
	input4 := &Person{Age: 22}
	input5 := &Person{Name: "Ana", Age: 22}
	input6 := &Person{Children: []Person{Person{Name: "a"}, Person{Name: "b"}, Person{Name: "c"}}}

	cases := []struct {
		ExprInput string
		EvalInput interface{}
		Opts      Opt
		WantResp  interface{}
		WantErr   error
	}{
		{`/Mom/Name`, input0, Opt{}, `Ana Belle`, nil},
		// Accommodate a special syntax that will be necessary for path queries.
		{`/Mom/(/Name)`, input0, Opt{}, `Ana Belle`, nil},
		{`/Mom/Name == Ana`, input1, Opt{}, true, nil},
		{`(/Mom/Name) == Ana`, input1, Opt{}, true, nil},
		// Make sure quotes are removed
		{`/Name == "Ana Belle"`, input2, Opt{}, true, nil},
		// Strictness -- by default strict is off, and incompatible comparisons result in false.
		{`/Name == 22`, input3, Opt{Strict: false}, false, nil},
		// Strictness -- if strict is on, report error with incompatible comparisons.
		{`/Name == 22`, input3, Opt{Strict: true}, false, mismatchErr},
		// Int evalation, equal and not equal.
		{`/Age == 22`, input4, Opt{}, true, nil},
		{`/Age != 22`, input4, Opt{}, false, nil},
		// Compound comparisons.
		{`/Name == "Ana" && /Age == 22`, input5, Opt{}, true, nil},
		{`(/Name == "Ana") && (/Age == 22)`, input5, Opt{}, true, nil},
		{`/Name == "Ana" || /Age == 23`, input5, Opt{}, true, nil},
		{`(/Name == "Ana") || (/Age == 23)`, input5, Opt{}, true, nil},
		{`(/Name == "Mana") || (/Age == 22)`, input5, Opt{}, true, nil},
		{`(/Name == "Mana") || (/Age == 23)`, input5, Opt{}, false, nil},
		// Path equality
		{`/Mom/Name == /Mom/Name`, input1, Opt{}, true, nil},
		// Select
		{`/Children/(/Name == "c")`, input6, Opt{}, []Person{Person{Name: "c"}}, nil},
		// Select, unwinding the results to a single item
		{`(/Children/(/Name == "c"))[0]`, input6, Opt{}, Person{Name: "c"}, nil},
		// Arrays
		{`/Children[0]`, input6, Opt{}, Person{Name: "a"}, nil},
		{`/Children[1]`, input6, Opt{}, Person{Name: "b"}, nil},
		{`/Children[1]/Name`, input6, Opt{}, "b", nil},
		{`[1]`, [2]string{"a", "b"}, Opt{}, "b", nil},
		{`([1]) == "b"`, [2]string{"a", "b"}, Opt{}, true, nil},
		{`/Children[0]`, input3, Opt{}, nil, nil},
		// Maps
		{`/a`, map[string]string{`a`: `a1`}, Opt{}, "a1", nil},
		// Special paths
		{`/a/b`, map[string]string{`a/b`: `a1`}, Opt{}, nil, nil},
		{`/"a/b"`, map[string]string{`a/b`: `a1`}, Opt{}, "a1", nil},
	}
	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			runTestExpr(t, tc.ExprInput, tc.EvalInput, tc.Opts, tc.WantResp, tc.WantErr)
			// Everything that works on the struct should work on the unmarshalled json.
			runTestExpr(t, tc.ExprInput, toJson(tc.EvalInput), tc.Opts, tc.WantResp, tc.WantErr)
		})
	}
}

func runTestExpr(t *testing.T, exprinput string, evalinput interface{}, opt Opt, wantResp interface{}, wantErr error) {
	expr, err := MakeExpr(exprinput)
	if err != nil {
		fmt.Println("make expr failed", err)
		printExprConstruction(exprinput)
		t.Fatal()
	}
	haveResp, haveErr := expr.Eval(evalinput, &opt)
	if !errorMatches(haveErr, wantErr) {
		fmt.Println("Error mismatch, have\n", haveErr, "\nwant\n", wantErr)
		t.Fatal()
	} else if !interfaceMatches(haveResp, wantResp) {
		fmt.Println("Response mismatch, have\n", toJsonString(haveResp), "\nwant\n", toJsonString(wantResp))
		fmt.Println("Response mismatch, have\n", haveResp, "\nwant\n", wantResp)
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

	input0 := &Metric{Level: 46.8}

	cases := []struct {
		Term     string
		Input    interface{}
		Opt      *Opt
		WantResp float64
	}{
		{`/Level`, input0, nil, 46.8},
		{`/NoLevel`, input0, nil, 0},
		{`/ErrorLevel`, input0, &Opt{OnError: 10.0}, 10.0},
	}
	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			haveResp := EvalFloat64(tc.Term, tc.Input, tc.Opt)
			if haveResp != tc.WantResp {
				fmt.Println("Response mismatch, have\n", haveResp, "\nwant\n", tc.WantResp)
				printExprConstruction(tc.Term)
				t.Fatal()
			}
		})
	}
}

// ------------------------------------------------------------
// TEST-EVAL-INT

func TestEvalInt(t *testing.T) {
	input0 := &Person{Age: 32}

	cases := []struct {
		Term     string
		Input    interface{}
		Opt      *Opt
		WantResp int
	}{
		{`/Age`, input0, nil, 32},
		{`/NoAge`, input0, nil, 0},
		{`/ErrorAge`, input0, &Opt{OnError: 10}, 10},
	}
	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			haveResp := EvalInt(tc.Term, tc.Input, tc.Opt)
			if haveResp != tc.WantResp {
				fmt.Println("Response mismatch, have\n", strconv.Itoa(haveResp), "\nwant\n", strconv.Itoa(tc.WantResp))
				printExprConstruction(tc.Term)
				t.Fatal()
			}
		})
	}
}

// ------------------------------------------------------------
// TEST-EVAL-STRING

func TestEvalString(t *testing.T) {
	input0 := &Person{Name: "Ana"}

	cases := []struct {
		Term     string
		Input    interface{}
		Opt      *Opt
		WantResp string
	}{
		{`/Name`, input0, nil, `Ana`},
		{`/NoName`, input0, nil, ``},
		{`/ErrorName`, input0, &Opt{OnError: `zip`}, `zip`},
	}
	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			haveResp := EvalString(tc.Term, tc.Input, tc.Opt)
			if haveResp != tc.WantResp {
				fmt.Println("Response mismatch, have\n", haveResp, "\nwant\n", tc.WantResp)
				printExprConstruction(tc.Term)
				t.Fatal()
			}
		})
	}
}

// ------------------------------------------------------------
// TEST-EVAL-STRING-MAP

func TestEvalStringMap(t *testing.T) {
	sub0 := map[string]string { "Ana": "Belle", "Cera": "Sayed" }
	input0 := map[string]interface{} { "Names": sub0 }
	sub1 := map[string]interface{} { "Ana": "Belle", "Cera": "Sayed" }
	input1 := map[string]interface{} { "Names": sub1 }

	cases := []struct {
		Term     string
		Input    interface{}
		Opt      *Opt
		WantResp interface{}
	}{
		{`/Names`, input0, nil, sub0},
		{`/Names`, input1, nil, sub1},
	}
	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			haveResp := EvalStringMap(tc.Term, tc.Input, tc.Opt)
			if !interfaceMatches(haveResp, tc.WantResp) {
				fmt.Println("Response mismatch, have\n", haveResp, "\nwant\n", tc.WantResp)
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
func tokens(all ...interface{}) []*nodeT {
	var tokens []*nodeT
	for _, t := range all {
		switch v := t.(type) {
		case float32:
			val := strconv.FormatFloat(float64(v), 'f', 8, 64)
			tokens = append(tokens, newNode(floatToken, val))
		case float64:
			val := strconv.FormatFloat(v, 'f', 8, 64)
			tokens = append(tokens, newNode(floatToken, val))
		case int:
			tokens = append(tokens, newNode(intToken, strconv.Itoa(v)))
		case string:
			tokens = append(tokens, newNode(stringToken, v).reclassify())
		}
	}
	return tokens
}

func arrayN(left, right *nodeT) *nodeT {
	return binN(openArrayToken, left, right)
}

// binN constructs a binary token from the symbol
func binN(sym symbol, left, right *nodeT) *nodeT {
	b := newNode(sym, tokenMap[sym].Text)
	b.addChild(left)
	b.addChild(right)
	return b
}

func eqlN(left, right *nodeT) *nodeT {
	return binN(eqlToken, left, right)
}

func floatN(v float64) *nodeT {
	text := strconv.FormatFloat(v, 'f', 8, 64)
	return newNode(floatToken, text)
}

func intN(v int) *nodeT {
	return newNode(intToken, strconv.Itoa(v))
}

func orN(left, right *nodeT) *nodeT {
	return binN(orToken, left, right)
}

func pathN(left, right *nodeT) *nodeT {
	b := newNode(pathToken, tokenMap[pathToken].Text)
	b.addChild(left)
	if right != nil {
		b.addChild(right)
	}
	return b
}

func selN(child *nodeT) *nodeT {
	return mkUnary(selectToken, child)
}

func strN(text string) *nodeT {
	return newNode(stringToken, text)
}

// mkUnary constructs a unary token from the symbol
func mkUnary(sym symbol, child *nodeT) *nodeT {
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
	aerr, aok := a.(*sqiErr)
	berr, bok := b.(*sqiErr)
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

func tokensMatch(a, b []*nodeT) bool {
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
func tokenMatches(a, b *nodeT) bool {
	if a == nil && b == nil {
		return true
	} else if a == nil || b == nil {
		return false
	}
	return a.Token.Symbol == b.Token.Symbol && a.Text == b.Text
}

// ------------------------------------------------------------
// COMPARE BOILERPLATE

func (n nodeT) MarshalJSON() ([]byte, error) {
	return orderedMarshalJSON(n)
}

func (t tokenT) MarshalJSON() ([]byte, error) {
	return orderedMarshalJSON(t)
}

// orderedMarshalJSON() is a generic function for writing ordered json data.
func orderedMarshalJSON(src interface{}) ([]byte, error) {
	var buf bytes.Buffer
	v := reflect.Indirect(reflect.ValueOf(src))
	var pairs []pairT
	for i := 0; i < v.NumField(); i++ {
		val := v.Field(i)
		field := v.Type().Field(i)
		if wantsJsonField(val, field) {
			pairs = append(pairs, pairT{field.Name, val.Interface()})
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

type pairT struct {
	Key   string
	Value interface{}
}

type ByPair []pairT

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
