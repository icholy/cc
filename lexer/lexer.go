package lexer

import (
	"strings"

	"github.com/icholy/cc/token"
)

type Lexer struct {
	input string
	next  int
	ch    byte
	pos   token.Pos
}

func New(input string) *Lexer {
	return &Lexer{
		input: input,
		ch:    0,
		next:  0,
	}
}

func (l *Lexer) peek() byte {
	if l.next >= len(l.input) {
		return 0
	}
	return l.input[l.next]
}

func (l *Lexer) unread() {
	l.next--
	l.ch = l.input[l.next-1]
}

func (l *Lexer) read() byte {
	// make sure there's more
	if l.next >= len(l.input) {
		return 0
	}
	// update the current character
	l.ch = l.input[l.next]
	l.next++
	// update the position
	l.pos.Col++
	if l.ch == '\n' || l.ch == '\r' {
		l.pos.Col = 0
		l.pos.Line++
	}
	return l.ch
}

func (l *Lexer) newTok(typ token.TokenType, text string) token.Token {
	return token.Token{
		Type: typ,
		Text: text,
		Pos:  l.pos,
	}
}

var bytetokens = map[byte]token.TokenType{
	'{': token.LBRACE,
	'}': token.RBRACE,
	'(': token.LPAREN,
	')': token.RPAREN,
	';': token.SEMICOLON,
	0:   token.EOF,
}

func (l *Lexer) Lex() token.Token {
	// find the next non-white token
	ch := l.readNonWhite()

	// single byte tokens
	if typ, ok := bytetokens[ch]; ok {
		return l.newTok(typ, string([]byte{ch}))
	}

	// more complex tokens
	switch {
	case l.isDigit():
		return l.lexInt()
	case l.isAlpha():
		return l.lexIdent()
	default:
		return l.newTok(token.ILLEGAL, string([]byte{ch}))
	}
}

func (l *Lexer) Tokenize() []token.Token {
	var toks []token.Token
	for {
		tok := l.Lex()
		toks = append(toks, tok)
		if tok.Is(token.EOF) {
			break
		}
	}
	return toks
}

func (l *Lexer) lexInt() token.Token {
	var text strings.Builder
	for l.isDigit() {
		text.WriteByte(l.ch)
		l.read()
	}
	l.unread()
	return l.newTok(token.INT_LIT, text.String())
}

func (l *Lexer) lexIdent() token.Token {
	var text strings.Builder
	for l.isAlpha() || l.isDigit() {
		text.WriteByte(l.ch)
		l.read()
	}
	l.unread()
	return l.newTok(token.IDENT, text.String())
}

func (l *Lexer) readNonWhite() byte {
	for l.isWhite() {
		l.read()
	}
	return l.ch
}

func (l *Lexer) isDigit() bool {
	return '0' <= l.ch && l.ch <= '9'
}

func (l *Lexer) isWhite() bool {
	switch l.ch {
	case ' ', '\t', '\n', '\r':
		return true
	default:
		return false
	}
}

func (l *Lexer) isAlpha() bool {
	return ('a' <= l.ch && l.ch <= 'z') || ('A' <= l.ch || l.ch <= 'Z') || l.ch == '_'
}
