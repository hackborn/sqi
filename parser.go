package sqi

import (
	"errors"
	"strconv"
)

// parse() converts tokens into an AST.
func parse(tokens []token_t) (AstNode, error) {
	var resp AstNode
	var err error
	for i := 0; i < len(tokens); i++ {
		resp, i, err = new_node(resp, i, tokens)
		if err != nil {
			return nil, err
		}
	}
	return resp, nil
}

func new_node(prv AstNode, pos int, tokens []token_t) (AstNode, int, error) {
	if pos >= len(tokens) {
		return nil, 0, errors.New("sqi: parse out of bounds")
	}
	token := tokens[pos]
	switch token.tok {
	case string_token:
		// The only time this is valid is when there's no previous
		if prv != nil {
			return nil, 0, errors.New("sqi: parse string node has a previous")
		}
		return &fieldNode{Field: token.text}, pos, nil
	case path_token:
		if prv == nil {
			return nil, 0, errors.New("sqi: parse path node has no previous")
		}
		rhs, pos, err := new_node(nil, pos+1, tokens)
		if err != nil {
			return nil, 0, err
		}
		return &pathNode{Lhs: prv, Rhs: rhs}, pos, nil
	}
	return nil, 0, errors.New("sqi: unknown token " + strconv.Itoa(int(token.tok)))
}
