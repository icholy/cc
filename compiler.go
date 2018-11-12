package cc

import (
	"fmt"
	"strings"

	"github.com/icholy/cc/ast"
	"github.com/icholy/cc/parser"
)

func Compile(src string) (string, error) {
	prog, err := parser.Parse(src)
	if err != nil {
		return "", err
	}
	var asm strings.Builder
	if err := compileProgram(prog, &asm); err != nil {
		return "", err
	}
	return asm.String(), nil
}

func compileProgram(p *ast.Program, asm *strings.Builder) error {
	switch stmt := p.Body.(type) {
	case *ast.Function:
		return compileFunction(stmt, asm)
	default:
		return fmt.Errorf("cannot compile: %s", p.Body)
	}
	return nil
}

func compileExpr(expr ast.Expr, asm *strings.Builder) error {
	switch expr := expr.(type) {
	case *ast.IntLiteral:
		fmt.Fprintf(asm, "movl $%d, %%eax\n", expr.Value)
	case *ast.UnaryOp:
		return compileUnaryOp(expr, asm)
	case *ast.BinaryOp:
		return compileBinaryOp(expr, asm)
	default:
		return fmt.Errorf("cannot compile: %s", expr)
	}
	return nil
}

func compileUnaryOp(unary *ast.UnaryOp, asm *strings.Builder) error {
	switch unary.Op {
	case "-":
		fmt.Fprintf(asm, "neg %%eax\n")
	case "~":
		fmt.Fprintf(asm, "not %%eax\n")
	case "!":
		fmt.Fprintf(asm, "cmpl $0, %%eax\n")
		fmt.Fprintf(asm, "movl $0, %%eax\n")
		fmt.Fprintf(asm, "sete %%al\n")
	default:
		return fmt.Errorf("invalid unary op: %s", unary)
	}
	return nil
}

func compileStmt(stmt ast.Stmt, asm *strings.Builder) error {
	switch stmt := stmt.(type) {
	case *ast.Return:
		if err := compileExpr(stmt.Value, asm); err != nil {
			return err
		}
		fmt.Fprintf(asm, "ret\n")
	default:
		return fmt.Errorf("cannot compile: %s", stmt)
	}
	return nil
}

func compileBinaryOp(binary *ast.BinaryOp, asm *strings.Builder) error {
	if err := compileExpr(binary.Left, asm); err != nil {
		return err
	}
	fmt.Fprintf(asm, "push %%eax\n")
	if err := compileExpr(binary.Right, asm); err != nil {
		return err
	}
	fmt.Fprintf(asm, "pop %%ecx\n")
	switch binary.Op {
	case "+":
		fmt.Fprintf(asm, "add %%ecx, %%eax\n")
	case "-":
		fmt.Fprintf(asm, "sub %%ecx, %%eax\n")
	case "*":
		fmt.Fprintf(asm, "imul %%ecx, %%eax\n")
	case "/":
		fmt.Fprintf(asm, "idiv %%ecx, %%eax\n")
	default:
		return fmt.Errorf("invalid binary op: %s", binary)
	}
	return nil
}

func compileFunction(f *ast.Function, asm *strings.Builder) error {
	fmt.Fprintf(asm, ".globl _%s\n", f.Name)
	fmt.Fprintf(asm, "_%s:\n", f.Name)
	return compileStmt(f.Body, asm)
}
