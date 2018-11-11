package lexer

import (
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

func makeValidStage1(name string, retval string) lexerTest {
	return lexerTest{
		SrcPath: filepath.Join("../testdata/stage_1/valid", name),
		Expected: []token.Token{
			token.New(token.INT_TYPE, "int"),
			token.New(token.IDENT, "main"),
			token.New(token.LPAREN, "("),
			token.New(token.RPAREN, ")"),
			token.New(token.LBRACE, "{"),
			token.New(token.RETURN, "return"),
			token.New(token.INT_LIT, retval),
			token.New(token.SEMICOLON, ";"),
			token.New(token.RBRACE, "}"),
			token.New(token.EOF, ""),
		},
	}
}

func TestLexerStage1(t *testing.T) {
	tests := []lexerTest{
		makeValidStage1("multi_digit.c", "100"),
		makeValidStage1("newlines.c", "0"),
		makeValidStage1("return_2.c", "2"),
		makeValidStage1("no_newlines.c", "0"),
		makeValidStage1("return_0.c", "0"),
		makeValidStage1("spaces.c", "0"),
	}

	for _, tt := range tests {
		t.Run(tt.SrcPath, tt.Run)
	}
}

func makeValidStage2(name string, retval ...token.Token) lexerTest {
	tt := lexerTest{
		SrcPath: filepath.Join("../testdata/stage_2/valid", name),
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

func TestLexerStage2(t *testing.T) {
	tests := []lexerTest{
		makeValidStage2("bitwise.c",
			token.New(token.BANG, "!"),
			token.New(token.INT_LIT, "12"),
		),
		makeValidStage2("bitwise_zero.c",
			token.New(token.TILDA, "~"),
			token.New(token.INT_LIT, "0"),
		),
		makeValidStage2("bitwise_zero.c",
			token.New(token.TILDA, "~"),
			token.New(token.INT_LIT, "0"),
		),
		makeValidStage2("neg.c",
			token.New(token.MINUS, "-"),
			token.New(token.INT_LIT, "5"),
		),
		makeValidStage2("nested_ops.c",
			token.New(token.BANG, "!"),
			token.New(token.MINUS, "-"),
			token.New(token.INT_LIT, "3"),
		),
	}
	for _, tt := range tests {
		t.Run(tt.SrcPath, tt.Run)
	}
}
