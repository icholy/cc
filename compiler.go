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
	asm    *strings.Builder
	frames []*Frame
}

func New() *Compiler {
	return &Compiler{
		asm: &strings.Builder{},
	}
}

func (c *Compiler) frame() *Frame {
	l := len(c.frames)
	return c.frames[l-1]
}

func (c *Compiler) framePush(frame *Frame) {
	c.frames = append(c.frames, frame)
}

func (c *Compiler) framePop() *Frame {
	l := len(c.frames)
	f := c.frames[l-1]
	c.frames = c.frames[:l-1]
	return f
}

func (c *Compiler) Assembly() string {
	return c.asm.String()
}

func (c *Compiler) emitf(format string, args ...interface{}) {
	fmt.Fprintf(c.asm, format+"\n", args...)
}

func (c *Compiler) Compile(p *ast.Program) error {
	switch stmt := p.Body.(type) {
	case *ast.Function:
		return c.compileFunction(stmt)
	default:
		return fmt.Errorf("cannot compile: %s", p.Body)
	}
}

func (c *Compiler) compileExpr(expr ast.Expr) error {
	switch expr := expr.(type) {
	case *ast.IntLiteral:
		c.emitf("movl $%d, %%eax", expr.Value)
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
		c.emitf("neg %%eax")
	case "~":
		c.emitf("not %%eax")
	case "!":
		c.emitf("cmpl $0, %%eax")
		c.emitf("movl $0, %%eax")
		c.emitf("sete %%al")
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
		c.emitf("add %%ecx, %%eax")
	case "-":
		c.emitf("sub %%ecx, %%eax")
	case "*":
		c.emitf("imul %%ecx, %%eax")
	case "/":
		c.emitf("idiv %%ecx, %%eax")
	case "==":
		c.emitf("cmpl %%eax, %%ecx")
		c.emitf("movl $0, %%eax")
		c.emitf("sete %%al")
	case "!=":
		c.emitf("cmpl %%eax, %%ecx")
		c.emitf("movl $0, %%eax")
		c.emitf("setne %%al")
	case ">":
		c.emitf("cmpl %%eax, %%ecx")
		c.emitf("movl $0, %%eax")
		c.emitf("setg %%al")
	case ">=":
		c.emitf("cmpl %%eax, %%ecx")
		c.emitf("movl $0, %%eax")
		c.emitf("setge %%al")
	case "<":
		c.emitf("cmpl %%eax, %%ecx")
		c.emitf("movl $0, %%eax")
		c.emitf("setl %%al")
	case "<=":
		c.emitf("cmpl %%eax, %%ecx")
		c.emitf("movl $0, %%eax")
		c.emitf("setle %%al")
	case "||":
		c.emitf("orl %%eax, %%ecx")
		c.emitf("movl $0, %%eax")
		c.emitf("setne %%al")
	case "&&":
		c.emitf("cmpl $0, %%ecx")
		c.emitf("setne %%cl")
		c.emitf("cmpl $0, %%eax")
		c.emitf("movl $0, %%eax")
		c.emitf("setne %%al")
		c.emitf("andb %%cl, %%al")
	default:
		return fmt.Errorf("invalid binary op: %s", binary)
	}
	return nil
}

type Frame struct {
	NumLocals int
	Offsets   map[string]int
}

func (c *Compiler) newFrame(block *ast.Block) *Frame {
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

	frame := c.newFrame(f.Body)
	c.framePush(frame)
	defer c.framePop()

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
