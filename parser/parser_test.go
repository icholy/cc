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
	AssertParsingStage(t, 3)
	AssertParsingStage(t, 4)
}

func withRetval(retval ast.Expr) *ast.Program {
	return &ast.Program{
		Body: &ast.Function{
			Name: "main",
			Body: &ast.Return{
				Value: retval,
			},
		},
	}
}

func TestAST(t *testing.T) {
	AssertEqualAST(t, "../testdata/stage_1/valid/return_2.c", withRetval(&ast.IntLiteral{Value: 2}))
	AssertEqualAST(t, "../testdata/stage_2/valid/neg.c", withRetval(&ast.UnaryOp{
		Op: "-",
		Value: &ast.IntLiteral{
			Value: 5,
		},
	}))
	AssertEqualAST(t, "../testdata/stage_3/valid/add.c", withRetval(
		&ast.BinaryOp{
			Op:    "+",
			Left:  &ast.IntLiteral{Value: 1},
			Right: &ast.IntLiteral{Value: 2},
		},
	))
	AssertEqualAST(t, "../testdata/stage_3/valid/associativity.c", withRetval(
		&ast.BinaryOp{
			Op: "-",
			Left: &ast.BinaryOp{
				Op:    "-",
				Left:  &ast.IntLiteral{Value: 1},
				Right: &ast.IntLiteral{Value: 2},
			},
			Right: &ast.IntLiteral{Value: 3},
		},
	))
	AssertEqualAST(t, "../testdata/stage_3/valid/precedence.c", withRetval(
		&ast.BinaryOp{
			Op:   "+",
			Left: &ast.IntLiteral{Value: 2},
			Right: &ast.BinaryOp{
				Op:    "*",
				Left:  &ast.IntLiteral{Value: 3},
				Right: &ast.IntLiteral{Value: 4},
			},
		},
	))
	AssertEqualAST(t, "../testdata/stage_4/valid/eq_true.c", withRetval(
		&ast.BinaryOp{
			Op:    "==",
			Left:  &ast.IntLiteral{Value: 1},
			Right: &ast.IntLiteral{Value: 1},
		},
	))
}

type validityTest struct {
	SrcPath string
	Valid   bool
}

func AssertEqualAST(t *testing.T, srcpath string, expected *ast.Program) {
	t.Run(srcpath, func(t *testing.T) {
		src, err := ioutil.ReadFile(srcpath)
		assert.NilError(t, err)
		actual, err := Parse(string(src))
		assert.NilError(t, err)
		assert.DeepEqual(t, expected, actual, cmp.Transformer("Token", func(tok token.Token) token.Token {
			return token.Token{}
		}))
	})
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
