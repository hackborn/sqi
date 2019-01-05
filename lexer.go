package sqi

import (
	"strings"
	"text/scanner"
	"unicode"
)

// scan() converts a string into tokens.
func scan(input string) ([]token_t, error) {
	var lexer scanner.Scanner
	//	src := `Links FromNode == ".pipe line"`
	//	src := `Links FromNode == ".pipe line" && ToNode == ".poof"`
	// src := `Links (FromNode == ".pipe line" && ToNode == ".poof")`
	//src := `Links (FromNode == ".pipe line" && ToNode == ".poof") [0:-1,10,11] FromPin = "ds"`
	//	src := `Links (FromNode == ".pipe line" && ToNode == ".poof") [0:-1,10,11] FromPin="sds"`
	//	src := `Links (FromNode == ".pipe line" && ToNode == ".poof") [0:-1,10,11] FromPin = "sds"`
	//src := `10-11`
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
			runer.addToken(token_t{float_token, lexer.TokenText()})
		case scanner.Int:
			runer.flush()
			runer.addToken(token_t{int_token, lexer.TokenText()})
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

// ----------------------------------------
// IDENT-RUNER

// ident_runer supplies the rules for turning runes into idents.
type ident_runer struct {
	accum  []rune
	tokens []token_t
}

func (r *ident_runer) isIdentRune(ch rune, i int) bool {
	// This is identical to the text scanner default. I would like the
	// scanner to smartly identify "&&" "==" etc as separate tokens, even
	// when there's no whitespace separating them from idents, but I can't
	// see any way the scanner woiuld support that behaviour.
	systemident := ch == '_' || unicode.IsLetter(ch) || unicode.IsDigit(ch) && i > 0
	return systemident
}

func (r *ident_runer) addString(s string) {
	r.addToken(token_t{string_token, s})
}

func (r *ident_runer) addToken(t token_t) {
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
	r.addToken(token_t{string_token, string(r.accum)})
	r.accum = nil
}

// ----------------------------------------
// TOKEN_T

type token_t struct {
	tok  Token
	text string
}

// reclassify() converts this token into one of the defined
// keywords, if appropriate. Ideally this is done directly
// in the scanning stage, but I'm not sure how to get the
// scanner to do that.
func (t token_t) reclassify() token_t {
	if t.tok != string_token {
		return t
	}
	if found, ok := keywords[t.text]; ok {
		return token_t{found, t.text}
	}
	return t
}

// ----------------------------------------
// CONST and VAR

type Token int

const (
	// Special tokens
	illegal_token Token = iota
	eof_token

	// Raw values.
	int_token    // 12345
	float_token  // 123.45
	string_token // "abc"

	// Assignment
	assign_token // =

	// Comparison
	eql_token // ==
	neq_token // !=

	// Conditional
	and_token // &&
	or_token  // ||

	// Precendence
	open_token  // (
	close_token // )
)

var (
	keywords = map[string]Token{
		`=`:  assign_token,
		`==`: eql_token,
		`!=`: neq_token,
		`&&`: and_token,
		`||`: or_token,
		`(`:  open_token,
		`)`:  close_token,
	}
)
