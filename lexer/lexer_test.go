package lexer

import (
	"fmt"
	"io/ioutil"
	"testing"

	"gotest.tools/assert"

	"github.com/icholy/cc/token"
)

type lexerTest struct {
	SrcPath  string
	Expected []token.Token
}

func (tt *lexerTest) Run(t *testing.T) {
	src, err := ioutil.ReadFile(tt.SrcPath)
	assert.NilError(t, err)
	l := New(string(src))
	var actual []token.Token
	for i, e := range tt.Expected {
		tok := l.Lex()
		tok.Pos = token.Pos{}
		actual = append(actual, tok)
		if tok != e {
			t.Logf("got:  %v", actual)
			t.Logf("want: %v", tt.Expected)
			t.Fatalf("test[%d] - wrong token. want=%s, got=%s", i, e, tok)
		}
	}
}

func withRetval(stage int, name string, retval ...token.Token) lexerTest {
	tt := lexerTest{
		SrcPath: fmt.Sprintf("../testdata/stage_%d/valid/%s", stage, name),
	}
	tt.Expected = append(tt.Expected,
		token.New(token.INT_TYPE, "int"),
		token.New(token.IDENT, "main"),
		token.New(token.LPAREN, "("),
		token.New(token.RPAREN, ")"),
		token.New(token.LBRACE, "{"),
		token.New(token.RETURN, "return"),
	)
	tt.Expected = append(tt.Expected, retval...)
	tt.Expected = append(tt.Expected,
		token.New(token.SEMICOLON, ";"),
		token.New(token.RBRACE, "}"),
		token.New(token.EOF, ""),
	)
	return tt
}

func TestLexer(t *testing.T) {
	tests := []lexerTest{
		withRetval(1, "multi_digit.c", token.New(token.INT_LIT, "100")),
		withRetval(1, "newlines.c", token.New(token.INT_LIT, "0")),
		withRetval(1, "return_2.c", token.New(token.INT_LIT, "2")),
		withRetval(1, "no_newlines.c", token.New(token.INT_LIT, "0")),
		withRetval(1, "return_0.c", token.New(token.INT_LIT, "0")),
		withRetval(1, "spaces.c", token.New(token.INT_LIT, "0")),
		withRetval(2, "bitwise.c",
			token.New(token.BANG, "!"),
			token.New(token.INT_LIT, "12"),
		),
		withRetval(2, "bitwise_zero.c",
			token.New(token.TILDA, "~"),
			token.New(token.INT_LIT, "0"),
		),
		withRetval(2, "bitwise_zero.c",
			token.New(token.TILDA, "~"),
			token.New(token.INT_LIT, "0"),
		),
		withRetval(2, "neg.c",
			token.New(token.MINUS, "-"),
			token.New(token.INT_LIT, "5"),
		),
		withRetval(2, "nested_ops.c",
			token.New(token.BANG, "!"),
			token.New(token.MINUS, "-"),
			token.New(token.INT_LIT, "3"),
		),
		withRetval(3, "add.c",
			token.New(token.INT_LIT, "1"),
			token.New(token.PLUS, "+"),
			token.New(token.INT_LIT, "2"),
		),
		withRetval(3, "precedence.c",
			token.New(token.INT_LIT, "2"),
			token.New(token.PLUS, "+"),
			token.New(token.INT_LIT, "3"),
			token.New(token.ASTERISK, "*"),
			token.New(token.INT_LIT, "4"),
		),
	}

	for _, tt := range tests {
		t.Run(tt.SrcPath, tt.Run)
	}
}
