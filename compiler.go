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
	case *ast.Var:
		return c.compileVar(expr)
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
		return c.compileReturn(stmt)
	case *ast.VarDec:
		return c.compileVarDec(stmt)
	default:
		return fmt.Errorf("cannot compile: %s", stmt)
	}
	return nil
}

func (c *Compiler) compileVarDec(dec *ast.VarDec) error {
	if dec.Value != nil {
		if err := c.compileExpr(dec.Value); err != nil {
			return err
		}
	} else {
		c.emitf("movl $0, %%eax")
	}
	offset, err := c.frame().Offset(dec.Name)
	if err != nil {
		return err
	}
	c.emitf("movl %%eax, %d(%%ebp)", offset)
	return nil
}

func (c *Compiler) compileAssign(assign *ast.Assignment) error {
	if err := c.compileExpr(assign.Value); err != nil {
		return err
	}
	offset, err := c.frame().Offset(assign.Var.Name)
	if err != nil {
		return err
	}
	c.emitf("movl %%eax, %d(%%ebp)", offset)
	return nil
}

func (c *Compiler) compileVar(v *ast.Var) error {
	offset, err := c.frame().Offset(v.Name)
	if err != nil {
		return err
	}
	c.emitf("movl %d(%%ebp), %%eax", offset)
	return nil
}

func (c *Compiler) compileReturn(ret *ast.Return) error {
	if err := c.compileExpr(ret.Value); err != nil {
		return err
	}
	c.emitf("jmp %s", c.frame().Exit)
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
	Entry     string
	Exit      string
	NumLocals int
	Offsets   map[string]int
}

func (f *Frame) Offset(name string) (int, error) {
	off, ok := f.Offsets[name]
	if !ok {
		return 0, fmt.Errorf("undefined: %s", name)
	}
	return off, nil
}

func (f *Frame) Size() int {
	return f.NumLocals * 4
}

func (c *Compiler) newFrame(f *ast.Function) *Frame {
	frame := &Frame{
		Entry:   fmt.Sprintf("_%s", f.Name),
		Exit:    fmt.Sprintf("_%s_exit", f.Name),
		Offsets: make(map[string]int),
	}
	for _, stmt := range f.Body.Statements {
		if dec, ok := stmt.(*ast.VarDec); ok {
			frame.NumLocals++
			frame.Offsets[dec.Name] = frame.NumLocals * -4
		}
	}
	return frame
}

func (c *Compiler) compileFunction(f *ast.Function) error {

	frame := c.newFrame(f)
	c.framePush(frame)
	defer c.framePop()

	c.emitf(".globl %s", frame.Entry)
	c.emitf("%s:", frame.Entry)
	c.emitf("push %%ebp")
	c.emitf("movl %%esp, %%ebp")
	c.emitf("subl $%d, %%esp", frame.Size())

	for _, stmt := range f.Body.Statements {
		if err := c.compileStmt(stmt); err != nil {
			return err
		}
	}

	c.emitf("%s:", frame.Exit)
	c.emitf("movl %%ebp, %%esp")
	c.emitf("pop %%ebp")
	c.emitf("ret")
	return nil
}
