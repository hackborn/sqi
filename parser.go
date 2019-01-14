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
func (n *token_t) addChild(child *token_t) {
	child.Parent = n
	n.Children = append(n.Children, child)
}

func (n *token_t) wrapChild(oldchild, newchild *token_t) error {
	if newchild.Left != nil {
		return newUnhandledError("wrap child with existing left")
	}
	if n.Left == oldchild {
		newchild.Left = n.Left
		n.Left = newchild
		oldchild.Parent = newchild
		newchild.Parent = n
		return nil
	} else if n.Right == oldchild {
		newchild.Left = n.Right
		n.Right = newchild
		oldchild.Parent = newchild
		newchild.Parent = n
		return nil
	} else {
		return newUnhandledError("wrap missing child")
	}
}

func (n *token_t) replaceChild(oldchild, newchild *token_t) error {
	for i, c := range n.Children {
		if c == oldchild {
			newchild.addChild(oldchild)
			newchild.Parent = n
			n.Children[i] = newchild
			return nil
		}
	}
	return errors.New("sqi: parse missing child")
}

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

// asAst() converts this node into an AST node, including special rules like the insert.
func (n *token_t) asAst() (AstNode, error) {
	node, err := n.nodeAsAst()
	if err != nil {
		return nil, err
	}
	if n.Insert == condition_token {
		node = &conditionNode{n.Insert, node}
	}
	return node, nil
}

// nodeAsAst() returns the AST node for this tree node.
func (n *token_t) nodeAsAst() (AstNode, error) {
	fmt.Println("ast", n.Text)
	switch n.Tok {
	case eql_token, neq_token, and_token, or_token:
		lhs, rhs, err := n.makeBinary()
		if err != nil {
			return nil, err
		}
		return &binaryNode{Op: n.Tok, Lhs: lhs, Rhs: rhs}, nil
	case float_token:
		if len(n.Children) != 0 {
			return nil, errors.New("sqi: parse float has wrong number of children: " + strconv.Itoa(len(n.Children)))
		}
		f64, err := strconv.ParseFloat(n.Text, 64)
		if err != nil {
			return nil, err
		}
		return &constantNode{Value: f64}, nil
	case int_token:
		if len(n.Children) != 0 {
			return nil, errors.New("sqi: parse int has wrong number of children: " + strconv.Itoa(len(n.Children)))
		}
		i, err := strconv.ParseInt(n.Text, 10, 32)
		if err != nil {
			return nil, err
		}
		return &constantNode{Value: int(i)}, nil
	case open_token:
		child, err := n.makeUnary()
		if err != nil {
			return nil, err
		}
		return &unaryNode{Op: open_token, Child: child}, nil
	case path_token:
		lhs, rhs, err := n.makeBinary()
		if err != nil {
			return nil, err
		}
		return &pathNode{Lhs: lhs, Rhs: rhs}, nil
	case string_token:
		if len(n.Children) != 0 {
			return nil, errors.New("sqi: parse string has wrong number of children: " + strconv.Itoa(len(n.Children)))
		}
		// Unwrap quoted text, which has served its purpose of allowing special characters.
		text := strings.Trim(n.Text, `"`)
		// There are rules on strings -- based on context I can be either a field or string node
		ctx := n.getCtx()
		if ctx == string_tree_ctx {
			return &constantNode{Value: text}, nil
		}
		return &fieldNode{Field: text}, nil
	}
	return nil, errors.New("sqi: parse on unknown token: " + strconv.Itoa(int(n.Tok)))
}

// asAst() converts this node into an AST node, including special rules like the insert.
func (n *token_t) asAst_orig() (AstNode, error) {
	node, err := n.nodeAsAst_orig()
	if err != nil {
		return nil, err
	}
	if n.Insert == condition_token {
		node = &conditionNode{n.Insert, node}
	}
	return node, nil
}

// nodeAsAst() returns the AST node for this tree node.
func (n *token_t) nodeAsAst_orig() (AstNode, error) {
	switch n.Tok {
	case eql_token, neq_token, and_token, or_token:
		lhs, rhs, err := n.makeBinary()
		if err != nil {
			return nil, err
		}
		return &binaryNode{Op: n.Tok, Lhs: lhs, Rhs: rhs}, nil
	case float_token:
		if n.Left != nil || n.Right != nil {
			return nil, newParseError("float " + n.Text + " has wrong number of children")
		}
		f64, err := strconv.ParseFloat(n.Text, 64)
		if err != nil {
			return nil, err
		}
		return &constantNode{Value: f64}, nil
	case int_token:
		if n.Left != nil || n.Right != nil {
			return nil, newParseError("int " + n.Text + " has wrong number of children")
		}
		i, err := strconv.ParseInt(n.Text, 10, 32)
		if err != nil {
			return nil, err
		}
		return &constantNode{Value: int(i)}, nil
	case open_token:
		child, err := n.makeUnary()
		if err != nil {
			return nil, err
		}
		return &unaryNode{Op: open_token, Child: child}, nil
	case path_token:
		lhs, rhs, err := n.makeBinary()
		if err != nil {
			return nil, err
		}
		return &pathNode{Lhs: lhs, Rhs: rhs}, nil
	case string_token:
		if n.Left != nil || n.Right != nil {
			return nil, newParseError("string " + n.Text + " has wrong number of children")
		}
		// Unwrap quoted text, which has served its purpose of allowing special characters.
		text := strings.Trim(n.Text, `"`)
		// There are rules on strings -- based on context I can be either a field or string node
		ctx := n.getCtx_orig()
		if ctx == string_tree_ctx {
			return &constantNode{Value: text}, nil
		}
		return &fieldNode{Field: text}, nil
	}
	return nil, errors.New("sqi: parse on unknown token: " + strconv.Itoa(int(n.Tok)))
}

func (n *token_t) makeBinary() (AstNode, AstNode, error) {
	if n.Left == nil {
		return nil, nil, newParseError("binary " + n.Text + " missing left")
	}
	if n.Right == nil {
		return nil, nil, newParseError("binary " + n.Text + " missing right")
	}
	lhs, err := n.Left.asAst()
	if err != nil {
		return nil, nil, err
	}
	rhs, err := n.Right.asAst()
	if err != nil {
		return nil, nil, err
	}
	return lhs, rhs, nil
}

func (n *token_t) makeUnary() (AstNode, error) {
	if n.Left == nil || n.Right != nil {
		return nil, newParseError("unary " + n.Text + " must have single term")
	}
	return n.Left.asAst()
}

func (n *token_t) makeBinary_orig() (AstNode, AstNode, error) {
	if len(n.Children) != 2 {
		return nil, nil, errors.New("sqi: parse binary has wrong number of children: " + strconv.Itoa(len(n.Children)))
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

func (n *token_t) makeUnary_orig() (AstNode, error) {
	if len(n.Children) != 1 {
		return nil, errors.New("sqi: parse unary has wrong number of children: " + strconv.Itoa(len(n.Children)))
	}
	return n.Children[0].asAst()
}

// getCtx() answers the context for this token, based on its position in the syntax tree.
func (n *token_t) getCtx() tree_ctx {
	// Not sure how this will evolve, but currently only strings that are rhs of binaries have meaning.
	if n.Tok != string_token || n.Parent == nil || n.Parent.Left == nil || n.Parent.Right == nil {
		return empty_tree_ctx
	}
	if !n.Parent.isToken(string_capable_rhs...) {
		return empty_tree_ctx
	}
	if n.Parent.Right == n {
		return string_tree_ctx
	}
	return empty_tree_ctx
}

func (n *token_t) getCtx_orig() tree_ctx {
	// Not sure how this will evolve, but currently only strings that are rhs of binaries have meaning.
	if n.Tok != string_token || n.Parent == nil || len(n.Parent.Children) != 2 {
		return empty_tree_ctx
	}
	if !n.Parent.isToken(string_capable_rhs...) {
		return empty_tree_ctx
	}
	if n.Parent.Children[1] == n {
		return string_tree_ctx
	}
	return empty_tree_ctx
}

func (n *token_t) isToken(tokens ...Token) bool {
	for _, t := range tokens {
		if n.Tok == t {
			return true
		}
	}
	return false
}
*/

// ------------------------------------------------------------
// TREE-NODE
/*
// tree_node is used to assemble the tokens into a tree.
type tree_node struct {
	Parent   *tree_node `json:"-"`
	Children []*tree_node
	T        token_t
	Insert   Token // A command to replace this node with a unary that contains it.
}

func (n *tree_node) addChild(child *tree_node) {
	child.Parent = n
	n.Children = append(n.Children, child)
}

func (n *tree_node) replaceChild(oldchild, newchild *tree_node) error {
	for i, c := range n.Children {
		if c == oldchild {
			newchild.addChild(oldchild)
			newchild.Parent = n
			n.Children[i] = newchild
			return nil
		}
	}
	return errors.New("sqi: parse missing child")
}

// setCondition() finds the proper node to insert a condtion node. This is used
// by boolean conditions: Every subgraph that needs to evaluate to true/false
// must be wrapped in a condition. Currently that means any comparison booleans,
// and the conditionals that can contain them.
func (n *tree_node) setCondition() error {
	if n.Parent == nil || !n.Parent.T.canHaveCondition() {
		return n.setInsert(condition_token)
	}
	return n.Parent.setCondition()
}

// setInsert() sets the insert value for this node. A node can only have a single
// insert type set -- any change will result in an error.
func (n *tree_node) setInsert(t Token) error {
	if n.Insert == illegal_token || n.Insert == t {
		n.Insert = t
		return nil
	}
	return newMismatchError("tree insert " + strconv.Itoa(int(n.Insert)) + " and " + strconv.Itoa(int(t)))
}

// asAst() converts this node into an AST node, including special rules like the insert.
func (n *tree_node) asAst() (AstNode, error) {
	node, err := n.nodeAsAst()
	if err != nil {
		return nil, err
	}
	if n.Insert == condition_token {
		node = &conditionNode{n.Insert, node}
	}
	return node, nil
}

// nodeAsAst() returns the AST node for this tree node.
func (n *tree_node) nodeAsAst() (AstNode, error) {
	switch n.T.Tok {
	case eql_token, neq_token, and_token, or_token:
		lhs, rhs, err := n.makeBinary()
		if err != nil {
			return nil, err
		}
		return &binaryNode{Op: n.T.Tok, Lhs: lhs, Rhs: rhs}, nil
	case float_token:
		if len(n.Children) != 0 {
			return nil, errors.New("sqi: parse float has wrong number of children: " + strconv.Itoa(len(n.Children)))
		}
		f64, err := strconv.ParseFloat(n.T.Text, 64)
		if err != nil {
			return nil, err
		}
		return &constantNode{Value: f64}, nil
	case int_token:
		if len(n.Children) != 0 {
			return nil, errors.New("sqi: parse int has wrong number of children: " + strconv.Itoa(len(n.Children)))
		}
		i, err := strconv.ParseInt(n.T.Text, 10, 32)
		if err != nil {
			return nil, err
		}
		return &constantNode{Value: int(i)}, nil
	case open_token:
		child, err := n.makeUnary()
		if err != nil {
			return nil, err
		}
		return &unaryNode{Op: open_token, Child: child}, nil
	case path_token:
		lhs, rhs, err := n.makeBinary()
		if err != nil {
			return nil, err
		}
		return &pathNode{Lhs: lhs, Rhs: rhs}, nil
	case string_token:
		if len(n.Children) != 0 {
			return nil, errors.New("sqi: parse string has wrong number of children: " + strconv.Itoa(len(n.Children)))
		}
		// Unwrap quoted text, which has served its purpose of allowing special characters.
		text := strings.Trim(n.T.Text, `"`)
		// There are rules on strings -- based on context I can be either a field or string node
		ctx := n.getCtx()
		if ctx == string_tree_ctx {
			return &constantNode{Value: text}, nil
		}
		return &fieldNode{Field: text}, nil
	}
	return nil, errors.New("sqi: parse on unknown token: " + strconv.Itoa(int(n.T.Tok)))
}

func (n *tree_node) makeBinary() (AstNode, AstNode, error) {
	if len(n.Children) != 2 {
		return nil, nil, errors.New("sqi: parse binary has wrong number of children: " + strconv.Itoa(len(n.Children)))
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

func (n *tree_node) makeUnary() (AstNode, error) {
	if len(n.Children) != 1 {
		return nil, errors.New("sqi: parse unary has wrong number of children: " + strconv.Itoa(len(n.Children)))
	}
	return n.Children[0].asAst()
}

// getCtx() answers the context for this token, based on its position in the syntax tree.
func (n *tree_node) getCtx() tree_ctx {
	// Not sure how this will evolve, but currently only strings that are rhs of binaries have meaning.
	if n.T.Tok != string_token || n.Parent == nil || len(n.Parent.Children) != 2 {
		return empty_tree_ctx
	}
	if !n.Parent.isToken(string_capable_rhs...) {
		return empty_tree_ctx
	}
	if n.Parent.Children[1] == n {
		return string_tree_ctx
	}
	return empty_tree_ctx
}

func (n *tree_node) isToken(tokens ...Token) bool {
	for _, t := range tokens {
		if n.T.Tok == t {
			return true
		}
	}
	return false
}
*/

// ------------------------------------------------------------
// CONST and VAR

type tree_ctx int

const (
	// Special tokens
	empty_tree_ctx tree_ctx = iota

	field_tree_ctx
	string_tree_ctx
)

var (
	string_capable_rhs = []symbol{assign_token, eql_token, neq_token}
)

// ------------------------------------------------------------
// BOILERPLATE

func parserFakeFmt() {
	fmt.Println()
}
