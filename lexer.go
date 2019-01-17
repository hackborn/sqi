package sqi

import (
	"strings"
	"text/scanner"
	"unicode"
)

// scan() converts a string into a flat list of tokens.
func scan(input string) ([]*node_t, error) {
	var lexer scanner.Scanner
	lexer.Init(strings.NewReader(input))
	// lexer.Whitespace = 1<<'\r' | 1<<'\t'
	lexer.Whitespace = 0
	lexer.Mode = scanner.ScanChars | scanner.ScanComments | scanner.ScanFloats | scanner.ScanIdents | scanner.ScanInts | scanner.ScanRawStrings | scanner.ScanStrings

	runer := &ident_runer{}
	lexer.IsIdentRune = runer.isIdentRune

	for tok := lexer.Scan(); tok != scanner.EOF; tok = lexer.Scan() {
		// fmt.Println("TOK", tok, "text", lexer.TokenText())
		switch tok {
		case scanner.Float:
			runer.flush()
			runer.addToken(newNode(float_token, lexer.TokenText()))
		case scanner.Int:
			runer.flush()
			runer.addToken(newNode(int_token, lexer.TokenText()))
		case scanner.Ident:
			runer.flush()
			runer.addString(lexer.TokenText())
		case scanner.String:
			runer.flush()
			runer.addString(lexer.TokenText())
		case ' ', '\r', '\t', '\n': // whitespace
			runer.flush()
		case scanner.Comment:
			runer.flush()
		default:
			runer.accumulate(tok)
		}
	}
	runer.flush()
	return runer.tokens, nil
}

// ------------------------------------------------------------
// IDENT-RUNER

// ident_runer supplies the rules for turning runes into idents.
type ident_runer struct {
	accum  []rune
	tokens []*node_t
}

func (r *ident_runer) isIdentRune(ch rune, i int) bool {
	// This is identical to the text scanner default. I would like the
	// scanner to smartly identify "&&" "==" etc as separate tokens, even
	// when there's no whitespace separating them from idents, but I can't
	// see any way the scanner would support that behaviour.
	systemident := ch == '_' || unicode.IsLetter(ch) || unicode.IsDigit(ch) && i > 0
	return systemident
}

func (r *ident_runer) addString(s string) {
	r.addToken(newNode(string_token, s))
}

func (r *ident_runer) addToken(t *node_t) {
	r.tokens = append(r.tokens, t.reclassify())
}

func (r *ident_runer) accumulate(ch rune) {
	// Single-character tokens are directly added
	switch ch {
	case '/':
		r.flush()
		r.addString(string(ch))
	default:
		r.accum = append(r.accum, ch)
	}
}

func (r *ident_runer) flush() {
	if len(r.accum) < 1 {
		return
	}
	r.addToken(newNode(string_token, string(r.accum)))
	r.accum = nil
}
