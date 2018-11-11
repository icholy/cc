package lexer

import (
	"io/ioutil"
	"testing"

	"gotest.tools/assert"

	"github.com/icholy/cc/token"
)

func TestLexer(t *testing.T) {
	tests := []struct {
		SrcPath  string
		Expected []token.Token
	}{
		{
			SrcPath: "../testdata/stage_1/valid/return_2.c",
			Expected: []token.Token{
				token.New(token.IDENT, "int"),
				token.New(token.IDENT, "main"),
				token.New(token.LPAREN, "("),
				token.New(token.RPAREN, ")"),
				token.New(token.LBRACE, "{"),
				token.New(token.IDENT, "return"),
				token.New(token.INT_LIT, "2"),
				token.New(token.SEMICOLON, ";"),
				token.New(token.RBRACE, "}"),
				token.New(token.EOF, ""),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.SrcPath, func(t *testing.T) {
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
		})
	}
}
