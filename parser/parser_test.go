package parser

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/icholy/cc/ast"
	"github.com/icholy/cc/token"

	"gotest.tools/assert"
)

func TestValidParsing(t *testing.T) {
	// AssertParsingStage(t, 1)
	// AssertParsingStage(t, 2)
	// AssertParsingStage(t, 3)
	// AssertParsingStage(t, 4)
	// AssertParsingStage(t, 5)
	// AssertParsingStage(t, 6)
	// AssertParsingStage(t, 7)
	AssertParsingStage(t, 8)
}

func withRetval(retval ast.Expr) *ast.Program {
	return &ast.Program{
		Body: &ast.FuncDec{
			Name: "main",
			Body: &ast.Block{
				Statements: []ast.Stmt{
					&ast.Ret{
						Value: retval,
					},
				},
			},
		},
	}
}

func TestAST(t *testing.T) {
	AssertEqualAST(t, "../testdata/stage_1/valid/return_2.c", withRetval(&ast.IntLit{Value: 2}))
	AssertEqualAST(t, "../testdata/stage_2/valid/neg.c", withRetval(&ast.UnaryOp{
		Op: "-",
		Value: &ast.IntLit{
			Value: 5,
		},
	}))
	AssertEqualAST(t, "../testdata/stage_3/valid/add.c", withRetval(
		&ast.BinaryOp{
			Op:    "+",
			Left:  &ast.IntLit{Value: 1},
			Right: &ast.IntLit{Value: 2},
		},
	))
	AssertEqualAST(t, "../testdata/stage_3/valid/associativity.c", withRetval(
		&ast.BinaryOp{
			Op: "-",
			Left: &ast.BinaryOp{
				Op:    "-",
				Left:  &ast.IntLit{Value: 1},
				Right: &ast.IntLit{Value: 2},
			},
			Right: &ast.IntLit{Value: 3},
		},
	))
	AssertEqualAST(t, "../testdata/stage_3/valid/precedence.c", withRetval(
		&ast.BinaryOp{
			Op:   "+",
			Left: &ast.IntLit{Value: 2},
			Right: &ast.BinaryOp{
				Op:    "*",
				Left:  &ast.IntLit{Value: 3},
				Right: &ast.IntLit{Value: 4},
			},
		},
	))
	AssertEqualAST(t, "../testdata/stage_4/valid/eq_true.c", withRetval(
		&ast.BinaryOp{
			Op:    "==",
			Left:  &ast.IntLit{Value: 1},
			Right: &ast.IntLit{Value: 1},
		},
	))
	AssertEqualAST(t, "../testdata/stage_6/valid/return_ternary.c", withRetval(
		&ast.Ternary{
			Condition: &ast.IntLit{Value: 1},
			Then:      &ast.IntLit{Value: 2},
			Else: &ast.Ternary{
				Condition: &ast.IntLit{Value: 3},
				Then:      &ast.IntLit{Value: 4},
				Else:      &ast.IntLit{Value: 5},
			},
		},
	))
	AssertEqualAST(t, "../testdata/stage_6/valid/else.c", &ast.Program{
		Body: &ast.FuncDec{
			Name: "main",
			Body: &ast.Block{
				Statements: []ast.Stmt{
					&ast.VarDec{
						Name:  "a",
						Value: &ast.IntLit{Value: 0},
					},
					&ast.If{
						Condition: &ast.Var{Name: "a"},
						Then: &ast.Ret{
							Value: &ast.IntLit{Value: 1},
						},
						Else: &ast.Ret{
							Value: &ast.IntLit{Value: 2},
						},
					},
				},
			},
		},
	})
	AssertEqualAST(t, "../testdata/stage_8/valid/for.c", &ast.Program{
		Body: &ast.FuncDec{
			Name: "main",
			Body: &ast.Block{
				Statements: []ast.Stmt{
					&ast.VarDec{
						Name:  "a",
						Value: &ast.IntLit{Value: 0},
					},
					&ast.For{
						Setup: &ast.ExprStmt{
							Expr: &ast.Assign{
								Var:   &ast.Var{Name: "a"},
								Value: &ast.IntLit{Value: 0},
							},
						},
						Condition: &ast.BinaryOp{
							Op:    "<",
							Left:  &ast.Var{Name: "a"},
							Right: &ast.IntLit{Value: 3},
						},
						Increment: &ast.Assign{
							Var: &ast.Var{Name: "a"},
							Value: &ast.BinaryOp{
								Op:    "+",
								Left:  &ast.Var{Name: "a"},
								Right: &ast.IntLit{Value: 1},
							},
						},
						Body: &ast.ExprStmt{
							Expr: &ast.Assign{
								Var: &ast.Var{Name: "a"},
								Value: &ast.BinaryOp{
									Op:    "*",
									Left:  &ast.Var{Name: "a"},
									Right: &ast.IntLit{Value: 2},
								},
							},
						},
					},
					&ast.Ret{
						Value: &ast.Var{Name: "a"},
					},
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
		name := fmt.Sprintf("stage_%d/%s", stage, filepath.Base(tt.SrcPath))
		t.Run(name, func(t *testing.T) {
			if strings.Contains(name, "__no_parse") {
				t.Skip()
			}
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
