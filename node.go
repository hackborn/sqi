package sqi

import (
	"strconv"
	"strings"
)

// ------------------------------------------------------------
// NODE-T

func newNode(s symbol, text string) *nodeT {
	token, ok := tokenMap[s]
	if !ok {
		token = tokenMap[illegalToken]
	}
	return &nodeT{Token: token, Text: text}
}

// nodeT serves two purposes: it is a token generated by the
// lexer, and it is a node in the tree constructed by the parser.
type nodeT struct {
	// Lexing
	Token *tokenT
	Text  string

	// Parsing
	Parent *nodeT `json:"-"`
	// XXX Note: This has stayed here because I really want to
	// remove the overhead of allocating a slice while building
	// the tree. For the whole time, no nodes can have more than
	// two children and I've been able to do that, but I suspect
	// future functions will require a child list.
	//	Left     *node_t
	//	Right    *node_t
	Children []*nodeT `json:"omitempty"`
}

// reclassify() converts this token into one of the defined
// keywords, if appropriate. Ideally this is done directly
// in the scanning stage, but I'm not sure how to get the
// scanner to do that.
func (n *nodeT) reclassify() *nodeT {
	if n.Token.Symbol != stringToken {
		return n
	}
	if found, ok := keywordMap[n.Text]; ok {
		return newNode(found.Symbol, n.Text)
	}
	return n
}

func (n *nodeT) addChild(child *nodeT) {
	n.Children = append(n.Children, child)
	child.Parent = n
}

// asAst() returns the AST node for this tree node.
func (n *nodeT) asAst() (AstNode, error) {
	// fmt.Println("ast", n.Text)
	switch n.Token.Symbol {
	case eqlToken, neqToken, andToken, orToken:
		lhs, rhs, err := n.makeBinary()
		if err != nil {
			return nil, err
		}
		return &binaryNode{Op: n.Token.Symbol, Lhs: lhs, Rhs: rhs}, nil
	case floatToken:
		if len(n.Children) != 0 {
			return nil, newParseError("float has wrong number of children: " + strconv.Itoa(len(n.Children)))
		}
		f64, err := strconv.ParseFloat(n.Text, 64)
		if err != nil {
			return nil, err
		}
		return &constantNode{Value: f64}, nil
	case intToken:
		if len(n.Children) != 0 {
			return nil, newParseError("int has wrong number of children: " + strconv.Itoa(len(n.Children)))
		}
		i, err := strconv.ParseInt(n.Text, 10, 32)
		if err != nil {
			return nil, err
		}
		return &constantNode{Value: int(i)}, nil
	case openToken:
		child, err := n.makeUnary()
		if err != nil {
			return nil, err
		}
		return &unaryNode{Op: openToken, Child: child}, nil
	case openArrayToken:
		return n.makeArray()
	case pathToken:
		return n.makePath()
	case stringToken:
		if len(n.Children) != 0 {
			return nil, newParseError("string has wrong number of children: " + strconv.Itoa(len(n.Children)))
		}
		// Unwrap quoted text, which has served its purpose of allowing special characters.
		text := strings.Trim(n.Text, `"`)
		return &constantNode{Value: text}, nil
	case selectToken:
		child, err := n.makeUnary()
		if err != nil {
			return nil, err
		}
		return &selectNode{Child: child}, nil
	}
	return nil, newParseError("on unknown token: " + strconv.Itoa(int(n.Token.Symbol)) + ", " + n.Token.Text)
}

func (n *nodeT) makeBinary() (AstNode, AstNode, error) {
	if len(n.Children) != 2 {
		return nil, nil, newParseError("binary has wrong number of children: " + strconv.Itoa(len(n.Children)))
	}
	lhs, err := n.Children[0].asAst()
	if err != nil {
		return nil, nil, err
	}
	rhs, err := n.Children[1].asAst()
	if err != nil {
		return nil, nil, err
	}
	return lhs, rhs, nil
}

func (n *nodeT) makeUnary() (AstNode, error) {
	if len(n.Children) != 1 {
		return nil, newParseError("unary has wrong number of children: " + strconv.Itoa(len(n.Children)))
	}
	return n.Children[0].asAst()
}

func (n *nodeT) makeArray() (AstNode, error) {
	// A single child is a special case -- it indicates we're at the top of the tree,
	// and we'll operate on whatever input I receive, instead of processing a lhs.
	var lhs AstNode
	var childidx int
	var err error

	switch len(n.Children) {
	case 1:
		childidx = 0
	case 2:
		lhs, err = n.Children[0].asAst()
		if err != nil {
			return nil, err
		}
		childidx = 1
	default:
		return nil, newParseError("array has wrong number of children: " + strconv.Itoa(len(n.Children)))
	}

	params, err := n.makeArrayParams(childidx)
	if err != nil {
		return nil, err
	}
	return &arrayNode{Lhs: lhs, Index: params}, nil
}

// makeArrayParams constructs the run params for an array node.
func (n *nodeT) makeArrayParams(childidx int) (int, error) {
	if childidx < 0 || childidx >= len(n.Children) {
		return 0, newParseError("array has missing child at " + strconv.Itoa(childidx))
	}
	child := n.Children[childidx]
	if child.Token.Symbol != intToken {
		return 0, newParseError("array must have int")
	}
	index, err := strconv.ParseInt(child.Text, 0, 32)
	if err != nil {
		return 0, err
	}
	return int(index), nil
}

func (n *nodeT) makePath() (AstNode, error) {
	// A path can have one or two children. If there are two, the first
	// must be a path and the second must be a string. If there is
	// one, the first must be a string.
	switch len(n.Children) {
	case 1:
		child0 := n.Children[0]
		// Parentheses can stack multiple paths. Consume them.
		for child0.Token.Symbol == pathToken && len(child0.Children) == 1 {
			child0 = child0.Children[0]
		}
		// Validate
		if child0.Token.Symbol != stringToken {
			return nil, newParseError("path must have string instead of " + child0.Token.Text)
		}
		text := strings.Trim(child0.Text, `"`)
		return &pathNode{Field: &fieldNode{Field: text}}, nil
	case 2:
		child0 := n.Children[0]
		child1 := n.Children[1]
		// Parentheses can stack multiple paths. Consume them.
		for child1.Token.Symbol == pathToken && len(child1.Children) == 1 {
			child1 = child1.Children[0]
		}
		// If we end in a string, we need to wrap
		var child1Ast AstNode
		if child1.Token.Symbol == stringToken {
			text := strings.Trim(child1.Text, `"`)
			child1Ast = &fieldNode{Field: text}
		} else {
			c1n, err := child1.asAst()
			if err != nil {
				return nil, err
			}
			child1Ast = c1n
		}
		cn, err := child0.asAst()
		if err != nil {
			return nil, err
		}
		return &pathNode{Child: cn, Field: child1Ast}, nil
	default:
		return nil, newParseError("path has wrong number of children: " + strconv.Itoa(len(n.Children)))
	}
}
