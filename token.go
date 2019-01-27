package sqi

import ()

// ------------------------------------------------------------
// TOKEN-T

// token_t stores a single token type, and associated behaviour.
type token_t struct {
	Symbol       symbol
	Text         string
	BindingPower int
	nud          nudFn
	led          ledFn
}

// any() answers true if my symbol is any of the supplied symbols.
func (t token_t) any(symbols ...symbol) bool {
	for _, s := range symbols {
		if t.Symbol == s {
			return true
		}
	}
	return false
}

// inside() answers true if my symbol is after start and before end.
func (t token_t) inside(start, end symbol) bool {
	return t.Symbol > start && t.Symbol < end
}

// ------------------------------------------------------------
// FUNC

type nudFn func(*node_t, *parserT) (*node_t, error)

type ledFn func(*node_t, *parserT, *node_t) (*node_t, error)

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

	// Assignment
	assign_token // =

	// Negation
	neg_token // -

	// Building
	path_token // /

	// Comparison
	start_comparison

	eql_token // ==
	neq_token // !=

	end_comparison

	// -- CONDITIONALS. All conditional operators must be after this
	start_conditional

	and_token // &&
	or_token  // ||

	// -- END CONDITIONALS.
	end_conditional

	// -- UNARIES. All unary operators must be after this
	start_unary

	// Enclosures
	open_token  // (
	open_array  // [
	close_token // ) // All closes must be after the opens
	close_array // ]

	// True/false condition
	select_token

	// -- END UNARIES.
	end_unary
)

var (
	token_map = map[symbol]*token_t{
		illegal_token: &token_t{illegal_token, "", 0, emptyNud, emptyLed},
		int_token:     &token_t{int_token, "", 0, emptyNud, emptyLed},
		float_token:   &token_t{float_token, "", 0, emptyNud, emptyLed},
		string_token:  &token_t{string_token, "", 0, emptyNud, emptyLed},
		assign_token:  &token_t{assign_token, "=", 80, emptyNud, binaryLed},
		neg_token:     &token_t{neg_token, "", 0, emptyNud, emptyLed},
		path_token:    &token_t{path_token, "/", 90, pathNud, binaryLed},
		eql_token:     &token_t{eql_token, "==", 70, emptyNud, binaryLed},
		neq_token:     &token_t{neq_token, "!=", 70, emptyNud, binaryLed},
		and_token:     &token_t{and_token, "&&", 60, emptyNud, binaryLed},
		or_token:      &token_t{or_token, "||", 60, emptyNud, binaryLed},
		open_token:    &token_t{open_token, "(", 0, enclosedNud, emptyLed},
		close_token:   &token_t{close_token, ")", 0, emptyNud, emptyLed},
		open_array:    &token_t{open_array, "[", 85, arrayNud, arrayLed},
		close_array:   &token_t{close_array, "]", 85, emptyNud, emptyLed},
		select_token:  &token_t{select_token, "", 100, emptyNud, emptyLed},
	}
	keyword_map = map[string]*token_t{
		`=`:  token_map[assign_token],
		`-`:  token_map[neg_token],
		`/`:  token_map[path_token],
		`==`: token_map[eql_token],
		`!=`: token_map[neq_token],
		`&&`: token_map[and_token],
		`||`: token_map[or_token],
		`(`:  token_map[open_token],
		`)`:  token_map[close_token],
		`[`:  token_map[open_array],
		`]`:  token_map[close_array],
	}
)

// ------------------------------------------------------------
// TOKEN FUNCS

func emptyNud(n *node_t, p *parserT) (*node_t, error) {
	return n, nil
}

func emptyLed(n *node_t, p *parserT, left *node_t) (*node_t, error) {
	return n, nil
}

func pathNud(n *node_t, p *parserT) (*node_t, error) {
	right, err := p.Expression(n.Token.BindingPower)
	if err != nil {
		return nil, err
	}
	n.addChild(right)
	return n, nil
}

func pathLed(n *node_t, p *parserT, left *node_t) (*node_t, error) {
	right, err := p.Expression(n.Token.BindingPower)
	if err != nil {
		return nil, err
	}
	n.addChild(right)
	n.addChild(left)
	return n, nil
}

func binaryLed(n *node_t, p *parserT, left *node_t) (*node_t, error) {
	n.addChild(left)
	right, err := p.Expression(n.Token.BindingPower)
	if err != nil {
		return nil, err
	}
	n.addChild(right)
	return n, nil
}

func enclosedNud(n *node_t, p *parserT) (*node_t, error) {
	enclosed, err := p.Expression(n.Token.BindingPower)
	if err != nil {
		return nil, err
	}
	next, err := p.Next()
	if err != nil {
		return nil, err
	}
	if next == nil {
		return nil, newParseError("missing next for " + n.Text)
	}
	if next.Token.Symbol != close_token {
		return nil, newParseError("missing close for " + n.Text)
	}
	return enclosed, nil
}

func arrayNud(n *node_t, p *parserT) (*node_t, error) {
	right, err := p.Expression(n.Token.BindingPower)
	if err != nil {
		return nil, err
	}
	next, err := p.Next()
	if err != nil {
		return nil, err
	}
	if next.Token.Symbol != close_array {
		return nil, newParseError("missing close for " + n.Text)
	}

	n.addChild(right)
	return n, nil
}

func arrayLed(n *node_t, p *parserT, left *node_t) (*node_t, error) {
	right, err := p.Expression(n.Token.BindingPower)
	if err != nil {
		return nil, err
	}
	next, err := p.Next()
	if err != nil {
		return nil, err
	}
	if next.Token.Symbol != close_array {
		return nil, newParseError("missing close for " + n.Text)
	}

	n.addChild(left)
	n.addChild(right)
	return n, nil
}
