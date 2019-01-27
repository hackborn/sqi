package sqi

import (
	"strings"
	"text/scanner"
	"unicode"
)

// scan converts a string into a flat list of tokens.
func scan(input string) ([]*nodeT, error) {
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
			runer.addToken(newNode(floatToken, lexer.TokenText()))
		case scanner.Int:
			runer.flush()
			runer.addToken(newNode(intToken, lexer.TokenText()))
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
	tokens []*nodeT
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
	r.addToken(newNode(stringToken, s))
}

func (r *ident_runer) addToken(t *nodeT) {
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
	// Recognize a token collection (i.e. tokens with no whitespace)
	// and place the remained in a string.
	accum := string(r.accum)
	for accum != "" {
		tok, s := r.extractToken(accum)
		if tok == nil {
			r.addToken(newNode(stringToken, s))
			accum = ""
		} else {
			r.addToken(newNode(tok.Symbol, tok.Text))
			accum = s
		}
	}
	r.accum = nil
}

func (r *ident_runer) extractToken(s string) (*tokenT, string) {
	var tok *tokenT
	for k, v := range keyword_map {
		if strings.HasPrefix(s, k) && (tok == nil || len(k) > len(tok.Text)) {
			tok = v
		}
	}
	if tok == nil {
		return nil, s
	}
	return tok, s[len(tok.Text):]
}
