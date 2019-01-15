package sqi

import (
	"fmt"
)

// parse() converts tokens into an AST.
func parse(tokens []*node_t) (AstNode, error) {
	tree, err := make_tree(tokens)
	if err != nil {
		return nil, err
	}
	//	if tree == nil || len(tree.Children) != 1 {
	//		return nil, errors.New("sqi: parse created empty tree")
	//	}
	// fmt.Println("tree:", toJson(tree))
	return tree.asAst()
}

// make_tree() creates the tree structure. It is solely concerned about
// the structure -- i.e. it cares that a token is a binary, but does not
// care what type of binary it is.
func make_tree(tokens []*node_t) (*node_t, error) {
	p := newParser(tokens)
	return p.Expression(0)
}
/*
// make_tree() creates the tree structure. It is solely concerned about
// the structure -- i.e. it cares that a token is a binary, but does not
// care what type of binary it is.
func make_tree_no(tokens []*node_t) (*node_t, error) {
	root := &node_t{}
	cur := root
	for _, t := range tokens {
		fmt.Println("token " + t.Text)
		if t.isBinary() {
			if cur.Parent == nil {
				return nil, newParseError("missing parent for " + t.Text)
			}
			// Find the token to wrap: the highest item with <= my binding power.
			wrap := cur

			// Three conditions:
			// If our binding power is greater than the parent, wrap the current.
			// If it's ==, go to the top.
			// If it's <, that's an error
			if t.Token.BindingPower > wrap.Parent.Token.BindingPower {

			} else if t.Token.BindingPower == wrap.Parent.Token.BindingPower {
				fmt.Println("bind ==")
				for wrap.Parent != nil && wrap.Parent.Parent != nil && wrap.Parent.Token.BindingPower == t.Token.BindingPower {
					fmt.Println("bind UP")
					wrap = wrap.Parent
				}
			} else {
				//				return nil, newParseError("lower binding for " + t.Text + " to " + wrap.Parent.Text)
			}
			//			print := t.Text == "=="
			fmt.Println("1 on", wrap, "with parent", wrap.Parent)
			print := false
			if print {
				//				fmt.Println("\tbefore", toJson(root))
			}
			err := wrap.Parent.wrapChild(wrap, t)
			fmt.Println("2")
			if print {
				//				fmt.Println("\tafter", toJson(root))
			}
			if err != nil {
				return nil, err
			}
			cur = t

			// Special insertion rules: Conditionals wrap condition nodes like equality, transforming
			// to the desired output.
			if cur.needsCondition() {
				err := cur.setCondition()
				if err != nil {
					return nil, err
				}
			}

		} else {
			if t.Text == "Name" {
				//			fmt.Println("\tcur is", cur)
			}
			print := false
			//			print := t.Text == "a"
			if print {
				//				fmt.Println("\tA before", toJson(root))
			}
			// If I'm a closing token, just set my cur accordingly.
			opener := t.isCloseFor()
			if opener != illegal_token {
				for cur.Parent != nil && cur.Token.Symbol != opener {
					cur = cur.Parent
				}
			} else if cur.Left == nil {
				t.Parent = cur
				cur.Left = t
				cur = t
			} else if cur.Right == nil {
				t.Parent = cur
				cur.Right = t
				cur = t
			} else {
				return nil, newParseError("invalid " + t.Text)
			}
			if print {
				//				fmt.Println("\tA after", toJson(root))
			}
		}
	}
	// fmt.Println("tree:", toJson(root))
	if root.Left == nil && root.Right != nil {
		return nil, newParseError("broken root")
	}
	return root.Left, nil
}
*/
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
// PARSE-NODE
// Additional behaviour on tokens so they can be assembled into a tree.

/*
// setCondition() finds the proper node to insert a condtion node. This is used
// by boolean conditions: Every subgraph that needs to evaluate to true/false
// must be wrapped in a condition. Currently that means any comparison booleans,
// and the conditionals that can contain them.
func (n *token_t) setCondition() error {
	if n.Parent == nil || !n.Parent.canHaveCondition() {
		return n.setInsert(condition_token)
	}
	return n.Parent.setCondition()
}

// setInsert() sets the insert value for this node. A node can only have a single
// insert type set -- any change will result in an error.
func (n *token_t) setInsert(t Token) error {
	if n.Insert == illegal_token || n.Insert == t {
		n.Insert = t
		return nil
	}
	return newMismatchError("tree insert " + strconv.Itoa(int(n.Insert)) + " and " + strconv.Itoa(int(t)))
}
*/

// ------------------------------------------------------------
// BOILERPLATE

func parserFakeFmt() {
	fmt.Println()
}
