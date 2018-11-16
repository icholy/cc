package token

import (
	"fmt"
)

type TokenType string

type Pos struct {
	Offset, Line, Col int
}

func (p Pos) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Col)
}

type Token struct {
	Pos  Pos
	Type TokenType
	Text string
}

func New(typ TokenType, text string) Token {
	return Token{Type: typ, Text: text}
}

func (t Token) Is(typ TokenType) bool {
	return t.Type == typ
}

func (t Token) OneOf(tt ...TokenType) bool {
	for _, typ := range tt {
		if t.Type == typ {
			return true
		}
	}
	return false
}

func (t Token) String() string {
	return fmt.Sprintf("%s %s(\"%s\")", t.Pos, t.Type, t.Text)
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
	RETURN    = "RETURN"
	MINUS     = "MINUS"
	PLUS      = "PLUS"
	ASTERISK  = "ASTERISK"
	SLASH     = "SLASH"
	TILDA     = "TILDA"
	BANG      = "BANG"
	AND       = "AND"
	OR        = "OR"
	EQ        = "EQ"
	NE        = "NE"
	LT        = "LT"
	LT_EQ     = "LT_EQ"
	GT        = "GT"
	GT_EQ     = "GT_EQ"
	ASSIGN    = "ASSIGN"
	IF        = "IF"
	ELSE      = "ELSE"
	COLON     = "COLON"
	QUESTION  = "QUESTION"
	DO        = "DO"
	WHILE     = "WHILE"
	FOR       = "FOR"
)

var Keywords = map[string]TokenType{
	"return": RETURN,
	"int":    INT_TYPE,
	"if":     IF,
	"else":   ELSE,
	"do":     DO,
	"while":  WHILE,
	"for":    FOR,
}
