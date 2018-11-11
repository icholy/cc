package lexer

import (
	"fmt"
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
		l.ch = 0
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

func (l *Lexer) newByteTok(typ token.TokenType) token.Token {
	return l.newTok(typ, string([]byte{l.ch}))
}

var bytetokens = map[byte]token.TokenType{
	'{': token.LBRACE,
	'}': token.RBRACE,
	'(': token.LPAREN,
	')': token.RPAREN,
	';': token.SEMICOLON,
}

func (l *Lexer) Lex() token.Token {
	// find the next non-white token
	l.read()
	l.whitespace()

	// check for end of file
	if l.ch == 0 {
		return l.newTok(token.EOF, "")
	}

	// single byte tokens
	if typ, ok := bytetokens[l.ch]; ok {
		return l.newByteTok(typ)
	}

	// more complex tokens
	switch {
	case l.isDigit():
		return l.lexInt()
	case l.isAlpha():
		tok := l.lexIdent()
		if typ, ok := token.Keywords[tok.Text]; ok {
			tok.Type = typ
		}
		return tok
	default:
		return l.newByteTok(token.ILLEGAL)
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

func (l *Lexer) whitespace() {
	for l.isWhite() {
		l.read()
	}
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
	return ('a' <= l.ch && l.ch <= 'z') || ('A' <= l.ch && l.ch <= 'Z') || l.ch == '_'
}

func (l *Lexer) Context(n int) string {
	start := l.next - n
	if start < 0 || n == 0 {
		start = 0
	}
	end := l.next + n
	if end >= len(l.input) || n == 0 {
		end = len(l.input) - 1
	}
	offset := l.next - start - 1
	src := l.input[start:end]
	src = strings.Replace(src, "\n", " ", -1)
	src = strings.Replace(src, "\r", " ", -1)
	return fmt.Sprintf("%s\n%s^", src, strings.Repeat(" ", offset))
}
