package token

import (
	"fmt"
)

type TokenType string

type Pos struct {
	Line, Col int
}

func (p Pos) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Col)
}

type Token struct {
	Pos  Pos
	Type TokenType
	Text string
}

func (t Token) Is(typ TokenType) bool {
	return t.Type == typ
}

func (t Token) String() string {
	return fmt.Sprintf("%s(\"%s\")", t.Type, t.Text)
}

const (
	ILLEGAL   = "ILLEGAL"
	EOF       = "EOF"
	IDENT     = "IDENT"
	LPAREN    = "LPAREN"
	RPAREN    = "RPAREN"
	LBRACE    = "LBRACE"
	RBRACE    = "RBRACE"
	SEMICOLON = "SEMICOLON"
	INT_LIT   = "INT_LIT"
	INT_TYPE  = "INT_TYPE"
)
