package sqi

import (
	"fmt"
)

// parse() converts a flat list of tokens into a tree.
func parse(tokens []*nodeT) (*nodeT, error) {
	p := newParser(tokens)
	return p.Expression(0)
}

// ------------------------------------------------------------
// PARSER-T

type parser interface {
	// Next answers the next node, or nil if we're finished.
	// Note that a finished condition is both the node and error being nil;
	// any error response is always an actual error.
	Next() (*nodeT, error)
	// Peek the next value. Note that this is never nil; an illegal is returned if we're at the end.
	Peek() *nodeT
	Expression(rbp int) (*nodeT, error)
}

type parserT struct {
	tokens   []*nodeT
	position int
	illegal  *nodeT
}

func newParser(tokens []*nodeT) parser {
	illegal := &nodeT{Token: tokenMap[illegalToken]}
	return &parserT{tokens: tokens, position: 0, illegal: illegal}
}

func (p *parserT) Next() (*nodeT, error) {
	if p.position >= len(p.tokens) {
		return nil, nil
	}
	pos := p.position
	p.position++
	return p.tokens[pos], nil
}

func (p *parserT) Peek() *nodeT {
	if p.position >= len(p.tokens) {
		return p.illegal
	}
	return p.tokens[p.position]
}

func (p *parserT) Expression(rbp int) (*nodeT, error) {
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
