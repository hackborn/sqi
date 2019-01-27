package sqi

import ()

// ------------------------------------------------------------
// TOKEN-T

// tokenT stores a single token type, and associated behaviour.
type tokenT struct {
	Symbol       symbol
	Text         string
	BindingPower int
	nud          nudFn
	led          ledFn
}

// any() answers true if my symbol is any of the supplied symbols.
func (t tokenT) any(symbols ...symbol) bool {
	for _, s := range symbols {
		if t.Symbol == s {
			return true
		}
	}
	return false
}

// inside() answers true if my symbol is after start and before end.
func (t tokenT) inside(start, end symbol) bool {
	return t.Symbol > start && t.Symbol < end
}

// ------------------------------------------------------------
// FUNC

type nudFn func(*nodeT, *parserT) (*nodeT, error)

type ledFn func(*nodeT, *parserT, *nodeT) (*nodeT, error)

// ------------------------------------------------------------
// CONST and VAR

type symbol int

const (
	// Special tokens
	illegalToken symbol = iota
	eofToken

	// Raw values
	intToken    // 12345
	floatToken  // 123.45
	stringToken // "abc"

	// Assignment
	assignToken // =

	// Negation
	negToken // -

	// Building
	pathToken // /

	// Comparison
	startComparison

	eqlToken // ==
	neqToken // !=

	endComparison

	// -- CONDITIONALS. All conditional operators must be after this
	startConditional

	andToken // &&
	orToken  // ||

	// -- END CONDITIONALS.
	endConditional

	// -- UNARIES. All unary operators must be after this
	startUnary

	// Enclosures
	openToken       // (
	openArrayToken  // [
	closeToken      // ) // All closes must be after the opens
	closeArrayToken // ]

	// True/false condition
	selectToken

	// -- END UNARIES.
	endUnary
)

var (
	token_map = map[symbol]*tokenT{
		illegalToken:    &tokenT{illegalToken, "", 0, emptyNud, emptyLed},
		intToken:        &tokenT{intToken, "", 0, emptyNud, emptyLed},
		floatToken:      &tokenT{floatToken, "", 0, emptyNud, emptyLed},
		stringToken:     &tokenT{stringToken, "", 0, emptyNud, emptyLed},
		assignToken:     &tokenT{assignToken, "=", 80, emptyNud, binaryLed},
		negToken:        &tokenT{negToken, "", 0, emptyNud, emptyLed},
		pathToken:       &tokenT{pathToken, "/", 90, pathNud, binaryLed},
		eqlToken:        &tokenT{eqlToken, "==", 70, emptyNud, binaryLed},
		neqToken:        &tokenT{neqToken, "!=", 70, emptyNud, binaryLed},
		andToken:        &tokenT{andToken, "&&", 60, emptyNud, binaryLed},
		orToken:         &tokenT{orToken, "||", 60, emptyNud, binaryLed},
		openToken:       &tokenT{openToken, "(", 0, enclosedNud, emptyLed},
		closeToken:      &tokenT{closeToken, ")", 0, emptyNud, emptyLed},
		openArrayToken:  &tokenT{openArrayToken, "[", 85, arrayNud, arrayLed},
		closeArrayToken: &tokenT{closeArrayToken, "]", 85, emptyNud, emptyLed},
		selectToken:     &tokenT{selectToken, "", 100, emptyNud, emptyLed},
	}
	keyword_map = map[string]*tokenT{
		`=`:  token_map[assignToken],
		`-`:  token_map[negToken],
		`/`:  token_map[pathToken],
		`==`: token_map[eqlToken],
		`!=`: token_map[neqToken],
		`&&`: token_map[andToken],
		`||`: token_map[orToken],
		`(`:  token_map[openToken],
		`)`:  token_map[closeToken],
		`[`:  token_map[openArrayToken],
		`]`:  token_map[closeArrayToken],
	}
)

// ------------------------------------------------------------
// TOKEN FUNCS

func emptyNud(n *nodeT, p *parserT) (*nodeT, error) {
	return n, nil
}

func emptyLed(n *nodeT, p *parserT, left *nodeT) (*nodeT, error) {
	return n, nil
}

func pathNud(n *nodeT, p *parserT) (*nodeT, error) {
	right, err := p.Expression(n.Token.BindingPower)
	if err != nil {
		return nil, err
	}
	n.addChild(right)
	return n, nil
}

func pathLed(n *nodeT, p *parserT, left *nodeT) (*nodeT, error) {
	right, err := p.Expression(n.Token.BindingPower)
	if err != nil {
		return nil, err
	}
	n.addChild(right)
	n.addChild(left)
	return n, nil
}

func binaryLed(n *nodeT, p *parserT, left *nodeT) (*nodeT, error) {
	n.addChild(left)
	right, err := p.Expression(n.Token.BindingPower)
	if err != nil {
		return nil, err
	}
	n.addChild(right)
	return n, nil
}

func enclosedNud(n *nodeT, p *parserT) (*nodeT, error) {
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
	if next.Token.Symbol != closeToken {
		return nil, newParseError("missing close for " + n.Text)
	}
	return enclosed, nil
}

func arrayNud(n *nodeT, p *parserT) (*nodeT, error) {
	right, err := p.Expression(n.Token.BindingPower)
	if err != nil {
		return nil, err
	}
	next, err := p.Next()
	if err != nil {
		return nil, err
	}
	if next.Token.Symbol != closeArrayToken {
		return nil, newParseError("missing close for " + n.Text)
	}

	n.addChild(right)
	return n, nil
}

func arrayLed(n *nodeT, p *parserT, left *nodeT) (*nodeT, error) {
	right, err := p.Expression(n.Token.BindingPower)
	if err != nil {
		return nil, err
	}
	next, err := p.Next()
	if err != nil {
		return nil, err
	}
	if next.Token.Symbol != closeArrayToken {
		return nil, newParseError("missing close for " + n.Text)
	}

	n.addChild(left)
	n.addChild(right)
	return n, nil
}
