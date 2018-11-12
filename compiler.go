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
	c := New()
	if err := c.Compile(prog); err != nil {
		return "", err
	}
	return c.Assembly(), nil
}

type Compiler struct {
	asm *strings.Builder
}

func New() *Compiler {
	return &Compiler{
		asm: &strings.Builder{},
	}
}

func (c *Compiler) Assembly() string {
	return c.asm.String()
}

func (c *Compiler) emitf(format string, args ...interface{}) {
	fmt.Fprintf(c.asm, format, args...)
}

func (c *Compiler) Compile(p *ast.Program) error {
	switch stmt := p.Body.(type) {
	case *ast.Function:
		return c.compileFunction(stmt)
	default:
		return fmt.Errorf("cannot compile: %s", p.Body)
	}
	return nil
}

func (c *Compiler) compileExpr(expr ast.Expr) error {
	switch expr := expr.(type) {
	case *ast.IntLiteral:
		c.emitf("movl $%d, %%eax\n", expr.Value)
	case *ast.UnaryOp:
		return c.compileUnaryOp(expr)
	case *ast.BinaryOp:
		return c.compileBinaryOp(expr)
	default:
		return fmt.Errorf("cannot compile: %s", expr)
	}
	return nil
}

func (c *Compiler) compileUnaryOp(unary *ast.UnaryOp) error {
	switch unary.Op {
	case "-":
		c.emitf("neg %%eax\n")
	case "~":
		c.emitf("not %%eax\n")
	case "!":
		c.emitf("cmpl $0, %%eax\n")
		c.emitf("movl $0, %%eax\n")
		c.emitf("sete %%al\n")
	default:
		return fmt.Errorf("invalid unary op: %s", unary)
	}
	return nil
}

func (c *Compiler) compileStmt(stmt ast.Stmt) error {
	switch stmt := stmt.(type) {
	case *ast.Return:
		if err := c.compileExpr(stmt.Value); err != nil {
			return err
		}
		c.emitf("ret\n")
	default:
		return fmt.Errorf("cannot compile: %s", stmt)
	}
	return nil
}

func (c *Compiler) compileBinaryOp(binary *ast.BinaryOp) error {
	if err := c.compileExpr(binary.Left); err != nil {
		return err
	}
	c.emitf("push %%eax\n")
	if err := c.compileExpr(binary.Right); err != nil {
		return err
	}
	c.emitf("pop %%ecx\n")
	switch binary.Op {
	case "+":
		c.emitf("add %%ecx, %%eax\n")
	case "-":
		c.emitf("sub %%ecx, %%eax\n")
	case "*":
		c.emitf("imul %%ecx, %%eax\n")
	case "/":
		c.emitf("idiv %%ecx, %%eax\n")
	case "==":
		c.emitf("cmpl %%eax, %%ecx\n")
		c.emitf("movl $0, %%eax\n")
		c.emitf("sete %%al\n")
	case "!=":
		c.emitf("cmpl %%eax, %%ecx\n")
		c.emitf("movl $0, %%eax\n")
		c.emitf("setne %%al\n")
	case ">":
		c.emitf("cmpl %%eax, %%ecx\n")
		c.emitf("movl $0, %%eax\n")
		c.emitf("setg %%al\n")
	case ">=":
		c.emitf("cmpl %%eax, %%ecx\n")
		c.emitf("movl $0, %%eax\n")
		c.emitf("setge %%al\n")
	case "<":
		c.emitf("cmpl %%eax, %%ecx\n")
		c.emitf("movl $0, %%eax\n")
		c.emitf("setl %%al\n")
	case "<=":
		c.emitf("cmpl %%eax, %%ecx\n")
		c.emitf("movl $0, %%eax\n")
		c.emitf("setle %%al\n")
	case "||":
		c.emitf("orl %%eax, %%ecx\n")
		c.emitf("movl $0, %%eax\n")
		c.emitf("setne %%al\n")
	case "&&":
		c.emitf("cmpl $0, %%ecx\n")
		c.emitf("setne %%cl\n")
		c.emitf("cmpl $0, %%eax\n")
		c.emitf("movl $0, %%eax\n")
		c.emitf("setne %%al\n")
		c.emitf("andb %%cl, %%al\n")
	default:
		return fmt.Errorf("invalid binary op: %s", binary)
	}
	return nil
}

type Frame struct {
	NumLocals int
	Offsets   map[string]int
}

func newFrame(block *ast.Block) *Frame {
	frame := &Frame{Offsets: make(map[string]int)}
	for _, stmt := range block.Statements {
		if dec, ok := stmt.(*ast.VarDec); ok {
			frame.NumLocals++
			frame.Offsets[dec.Name] = frame.NumLocals * 4
		}
	}
	return frame
}

func (c *Compiler) compileFunction(f *ast.Function) error {

	c.emitf(".globl _%s\n", f.Name)
	c.emitf("_%s:\n", f.Name)
	//c.emitf("push %%ebp\n")
	//c.emitf("movl %%esp, %%ebp\n")

	for _, stmt := range f.Body.Statements {
		if err := c.compileStmt(stmt); err != nil {
			return err
		}
	}

	//c.emitf("movl %%ebp, %%esp\n")
	//c.emitf("pop %%ebp\n")
	c.emitf("ret")
	return nil
}
