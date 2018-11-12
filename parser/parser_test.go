package parser

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/icholy/cc/ast"
	"github.com/icholy/cc/token"

	"gotest.tools/assert"
)

func TestValidParsing(t *testing.T) {
	AssertParsingStage(t, 1)
	AssertParsingStage(t, 2)
}

func TestAST(t *testing.T) {
	AssertEqualAST(t, "../testdata/stage_1/valid/return_2.c", &ast.Program{
		Body: &ast.Function{
			Name: "main",
			Body: &ast.Return{
				Value: &ast.IntLiteral{
					Value: 2,
				},
			},
		},
	})
}

type validityTest struct {
	SrcPath string
	Valid   bool
}

func AssertEqualAST(t *testing.T, srcpath string, expected *ast.Program) {
	src, err := ioutil.ReadFile(srcpath)
	assert.NilError(t, err)
	actual, err := Parse(string(src))
	assert.NilError(t, err)
	assert.DeepEqual(t, expected, actual, cmp.Transformer("Token", func(tok token.Token) token.Token {
		return token.Token{}
	}))
}

func AssertParsingStage(t *testing.T, stage int) {
	var tests []validityTest
	valid, err := filepath.Glob(fmt.Sprintf("../testdata/stage_%d/valid/*.c", stage))
	assert.NilError(t, err)
	for _, path := range valid {
		tests = append(tests, validityTest{path, true})
	}
	invalid, err := filepath.Glob(fmt.Sprintf("../testdata/stage_%d/invalid/*.c", stage))
	assert.NilError(t, err)
	for _, path := range invalid {
		tests = append(tests, validityTest{path, false})
	}
	for _, tt := range tests {
		t.Run(tt.SrcPath, func(t *testing.T) {
			src, err := ioutil.ReadFile(tt.SrcPath)
			assert.NilError(t, err)
			_, err = Parse(string(src))
			if tt.Valid {
				assert.NilError(t, err)
			} else {
				assert.Assert(t, err != nil)
			}
		})
	}
}
