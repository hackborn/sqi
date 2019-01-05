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
		//		{tokens(`Child`, `/`, `Name`, `==`, `a`), parser_want_3, nil},
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
	parser_want_0 = &fieldNode{Field: `Child`}
	parser_want_1 = &pathNode{Lhs: &fieldNode{Field: `Child`}, Rhs: &fieldNode{Field: `Name`}}
	parser_want_2 = &pathNode{Lhs: &fieldNode{Field: `Child`}, Rhs: &pathNode{Lhs: &fieldNode{Field: `Arm`}, Rhs: &fieldNode{Field: `Length`}}}
	parser_want_3 = &pathNode{Lhs: &fieldNode{Field: `Child`}, Rhs: &pathNode{Lhs: &fieldNode{Field: `Name`}, Rhs: &fieldNode{Field: `a`}}}
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
