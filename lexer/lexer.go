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
		pos:   token.Pos{-1, 1, 0},
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
	// update the position
	l.pos.Offset = l.next
	l.pos.Col++
	if l.ch == '\n' || l.ch == '\r' {
		l.pos.Col = 1
		l.pos.Line++
	}
	// update the index
	l.next++
	return l.ch
}

func (l *Lexer) newTok(typ token.TokenType, text string, pos token.Pos) token.Token {
	return token.Token{
		Type: typ,
		Text: text,
		Pos:  pos,
	}
}

func (l *Lexer) newByteTok(typ token.TokenType, pos token.Pos) token.Token {
	return l.newTok(typ, string([]byte{l.ch}), pos)
}

var bytetokens = map[byte]token.TokenType{
	'{': token.LBRACE,
	'}': token.RBRACE,
	'(': token.LPAREN,
	')': token.RPAREN,
	';': token.SEMICOLON,
	'-': token.MINUS,
	'+': token.PLUS,
	'*': token.ASTERISK,
	'/': token.SLASH,
	'~': token.TILDA,
	'!': token.BANG,
	'>': token.GT,
	'<': token.LT,
	'=': token.ASSIGN,
	'?': token.QUESTION,
	':': token.COLON,
}

var twobytetokens = map[string]token.TokenType{
	"&&": token.AND,
	"||": token.OR,
	"==": token.EQ,
	"!=": token.NE,
	"<=": token.LT_EQ,
	">=": token.GT_EQ,
}

func (l *Lexer) Lex() token.Token {
	// find the next non-white token
	l.read()
	l.whitespace()
	pos := l.pos

	// check for end of file
	if l.ch == 0 {
		return l.newTok(token.EOF, "", pos)
	}

	// double byte tokens
	twobytes := l.twoBytes()
	if typ, ok := twobytetokens[twobytes]; ok {
		l.read()
		return l.newTok(typ, twobytes, pos)
	}

	// single byte tokens
	if typ, ok := bytetokens[l.ch]; ok {
		return l.newByteTok(typ, pos)
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
		return l.newByteTok(token.ILLEGAL, pos)
	}
}

func (l *Lexer) twoBytes() string {
	return string([]byte{l.ch, l.peek()})
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
	pos := l.pos
	var text strings.Builder
	for l.isDigit() {
		text.WriteByte(l.ch)
		l.read()
	}
	l.unread()
	return l.newTok(token.INT_LIT, text.String(), pos)
}

func (l *Lexer) lexIdent() token.Token {
	pos := l.pos
	var text strings.Builder
	for l.isAlpha() || l.isDigit() {
		text.WriteByte(l.ch)
		l.read()
	}
	l.unread()
	return l.newTok(token.IDENT, text.String(), pos)
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

func (l *Lexer) Context(tok token.Token) string {
	col := tok.Pos.Col - 2
	if col < 0 {
		col = 0
	}
	lines := strings.Split(l.input, "\n")
	if tok.Pos.Line > len(lines) {
		return fmt.Sprintf("want %s, but only have %d lines", tok.Pos, len(lines))
	}
	return fmt.Sprintf("%s\n%s\n%s^\n%s",
		strings.Join(lines[:tok.Pos.Line-1], "\n"),
		lines[tok.Pos.Line-1],
		strings.Repeat("-", col),
		strings.Join(lines[tok.Pos.Line:], "\n"),
	)
}
