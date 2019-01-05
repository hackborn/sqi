package sqi

import (
	"fmt"
	"testing"
)

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
			} else if !astMatches(have_resp, tc.WantResp) {
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
// BUILD

// --------------------------------------------------------------------------------------
// COMPARE

func astMatches(a, b AstNode) bool {
	if a == nil && b == nil {
		return true
	} else if a == nil || b == nil {
		return false
	}
	ja := toJsonString(a)
	jb := toJsonString(b)
	return ja == jb
}
