package sqi

import (
	"fmt"
	"testing"
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
)
