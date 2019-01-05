package sqi

import (
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
		{`Child/Name`, tokens(`Child`, `/`, `Name`), nil},
		{`Child		/	Name`, tokens(`Child`, `/`, `Name`), nil},
		{`Child/Name=="a"`, tokens(`Child`, `/`, `Name`, `==`, `"a"`), nil},
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
// BUILD

func tokens(all ...interface{}) []token_t {
	var tokens []token_t
	for _, t := range all {
		switch v := t.(type) {
		case int:
			tokens = append(tokens, token_t{int_token, strconv.Itoa(v)})
		case string:
			tokens = append(tokens, token_t{string_token, v}.reclassify())
		}
	}
	return tokens
}

// --------------------------------------------------------------------------------------
// COMPARE

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
