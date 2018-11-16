package lexer

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
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
		withRetval(4, "and_false.c",
			token.New(token.INT_LIT, "1"),
			token.New(token.AND, "&&"),
			token.New(token.INT_LIT, "0"),
		),
		withRetval(4, "and_true.c",
			token.New(token.INT_LIT, "1"),
			token.New(token.AND, "&&"),
			token.New(token.MINUS, "-"),
			token.New(token.INT_LIT, "1"),
		),
	}

	for _, tt := range tests {
		t.Run(tt.SrcPath, tt.Run)
	}
}

func TestLexerPos(t *testing.T) {
	tests := []struct {
		file     string
		index    int
		expected token.Pos
	}{
		{
			file:     "../testdata/stage_3/valid/add.c",
			index:    0,
			expected: token.Pos{Offset: 0, Line: 1, Col: 1},
		},
		{
			file:     "../testdata/stage_3/valid/add.c",
			index:    5,
			expected: token.Pos{Offset: 18, Line: 2, Col: 6},
		},
	}
	for _, tt := range tests {
		t.Run(filepath.Base(tt.file), func(t *testing.T) {
			data, err := ioutil.ReadFile(tt.file)
			assert.NilError(t, err)
			toks := New(string(data)).Tokenize()
			assert.Assert(t, tt.index < len(toks))
			actual := toks[tt.index].Pos
			assert.Equal(t, tt.expected, actual)
		})
	}
}
