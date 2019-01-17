package sqi

import (
	"fmt"
)

// parse() converts a flat list of tokens into a tree.
func parse(tokens []*node_t) (*node_t, error) {
	p := newParser(tokens)
	return p.Expression(0)
}

// ------------------------------------------------------------
// PARSER-T

type parser interface {
	// Next answers the next node, or nil if we're finished.
	// Note that a finished condition is both the node and error being nil;
	// any error response is always an actual error.
	Next() (*node_t, error)
	// Peek the next value. Note that this is never nil; an illegal is returned if we're at the end.
	Peek() *node_t
	Expression(rbp int) (*node_t, error)
}

type parser_t struct {
	tokens   []*node_t
	position int
	illegal  *node_t
}

func newParser(tokens []*node_t) parser {
	illegal := &node_t{Token: token_map[illegal_token]}
	return &parser_t{tokens: tokens, position: 0, illegal: illegal}
}

func (p *parser_t) Next() (*node_t, error) {
	if p.position >= len(p.tokens) {
		return nil, nil
	}
	pos := p.position
	p.position++
	return p.tokens[pos], nil
}

func (p *parser_t) Peek() *node_t {
	if p.position >= len(p.tokens) {
		return p.illegal
	}
	return p.tokens[p.position]
}

func (p *parser_t) Expression(rbp int) (*node_t, error) {
	n, err := p.Next()
	if err != nil {
		return nil, err
	}
	if n == nil {
		return nil, newParseError("premature stop")
	}
	//	fmt.Println("Expression on rbp", rbp, "next \"", n.Text, "\"", n.Token)
	left, err := n.Token.nud(n, p)
	if err != nil {
		return nil, err
	}
	//	fmt.Println("peek binding", p.Peek().Token.BindingPower)
	for rbp < p.Peek().Token.BindingPower {
		n, err = p.Next()
		//		fmt.Println("\tloop rbp", rbp, "next \"", n.Text, "\"", n.Token)
		if err != nil {
			return nil, err
		}
		if n == nil {
			return nil, newParseError("premature stop")
		}
		left, err = n.Token.led(n, p, left)
		if err != nil {
			return nil, err
		}
	}
	return left, nil
}

// ------------------------------------------------------------
// BOILERPLATE

func parserFakeFmt() {
	fmt.Println()
}
