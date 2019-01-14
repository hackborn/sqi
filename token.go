package sqi

import ()

// ------------------------------------------------------------
// TOKEN_T

// token_t stores a single token type, and associated behaviour.
type token_t struct {
	Symbol       symbol
	Text         string
	BindingPower int
	nud          nudFn
	led          ledFn
}

// ------------------------------------------------------------
// FUNC

type nudFn func(*node_t, *parser_t) (*node_t, error)

type ledFn func(*node_t, *parser_t, *node_t) (*node_t, error)

// ------------------------------------------------------------
// CONST and VAR

type symbol int

const (
	// Special tokens
	illegal_token symbol = iota
	eof_token

	// Raw values
	int_token    // 12345
	float_token  // 123.45
	string_token // "abc"

	// -- BINARIES. All binary operators must be after this
	start_binary

	// Assignment
	assign_token // =

	// Building
	path_token // /

	// Comparison
	eql_token // ==
	neq_token // !=

	// -- CONDITIONALS. All conditional operators must be after this
	start_conditional

	and_token // &&
	or_token  // ||

	// -- END CONDITIONALS.
	end_conditional

	// -- END BINARIES.
	end_binary

	// -- UNARIES. All unary operators must be after this
	start_unary

	// Enclosures
	open_token  // (
	open_array  // [
	close_token // ) // All closes must be after the opens
	close_array // ]

	// True/false condition
	condition_token

	// -- END UNARIES.
	end_unary
)

var (
	token_map = map[symbol]*token_t{
		illegal_token:   &token_t{illegal_token, "", 0, emptyNud, emptyLed},
		int_token:       &token_t{int_token, "", 0, emptyNud, emptyLed},
		float_token:     &token_t{float_token, "", 0, emptyNud, emptyLed},
		string_token:    &token_t{string_token, "", 0, emptyNud, emptyLed},
		assign_token:    &token_t{assign_token, "=", 80, emptyNud, binaryLed},
		path_token:      &token_t{path_token, "/", 50, emptyNud, binaryLed},
		eql_token:       &token_t{eql_token, "==", 70, emptyNud, binaryLed},
		neq_token:       &token_t{neq_token, "!=", 70, emptyNud, binaryLed},
		and_token:       &token_t{and_token, "&&", 60, emptyNud, binaryLed},
		or_token:        &token_t{or_token, "||", 60, emptyNud, binaryLed},
		open_token:      &token_t{open_token, "(", 0, enclosedNud, emptyLed},
		close_token:     &token_t{close_token, ")", 0, emptyNud, emptyLed},
		condition_token: &token_t{condition_token, "", 100, emptyNud, emptyLed},
	}
	keyword_map = map[string]*token_t{
		`=`:  token_map[assign_token],
		`/`:  token_map[path_token],
		`==`: token_map[eql_token],
		`!=`: token_map[neq_token],
		`&&`: token_map[and_token],
		`||`: token_map[or_token],
		`(`:  token_map[open_token],
		`)`:  token_map[close_token],
	}
)

// ------------------------------------------------------------
// TOKEN FUNCS

func emptyNud(n *node_t, p *parser_t) (*node_t, error) {
	return n, nil
}

func emptyLed(n *node_t, p *parser_t, left *node_t) (*node_t, error) {
	return n, nil
}

func binaryLed(n *node_t, p *parser_t, left *node_t) (*node_t, error) {
	n.Children = append(n.Children, left)
	right, err := p.Expression(n.Token.BindingPower)
	if err != nil {
		return nil, err
	}
	n.Children = append(n.Children, right)
	return n, nil
}

func enclosedNud(n *node_t, p *parser_t) (*node_t, error) {
	enclosed, err := p.Expression(n.Token.BindingPower)
	if err != nil {
		return nil, err
	}
	next, err := p.Next()
	if err != nil {
		return nil, err
	}
	if next.Token.Symbol != close_token {
		return nil, newParseError("missing close for " + n.Text)
	}
	return enclosed, nil
}
