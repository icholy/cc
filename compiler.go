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

type Local struct {
	Name     string
	Declared bool
	Offset   int
}

type Scope struct {
	Parent *Scope
	Offset int
	Locals map[string]*Local
}

func (s *Scope) Local(name string) (*Local, error) {
	if loc, ok := s.Locals[name]; ok {
		return loc, nil
	}
	if s.Parent != nil {
		return s.Parent.Local(name)
	}
	return nil, fmt.Errorf("undefined: %s", name)
}

func (s *Scope) ParentOffset() int {
	if s.Parent == nil {
		return 0
	}
	return s.Parent.Offset
}

func (s *Scope) AddVarDec(d *ast.VarDec) {
	s.Offset -= 4
	s.Locals[d.Name] = &Local{
		Name:     d.Name,
		Offset:   s.ParentOffset() + s.Offset,
		Declared: false,
	}
}

type Frame struct {
	NumLabels int
	Entry     string
	Exit      string
	Scope     *Scope
}

func (f *Frame) Local(name string) (*Local, error) {
	return f.Scope.Local(name)
}

func (f *Frame) Label() string {
	f.NumLabels++
	return fmt.Sprintf("_%s_l%d", f.Entry, f.NumLabels)
}

func (c *Compiler) newFrame(f *ast.FuncDec) *Frame {
	frame := &Frame{
		Entry: fmt.Sprintf("_%s", f.Name),
		Exit:  fmt.Sprintf("_%s_exit", f.Name),
		Scope: &Scope{
			Locals: make(map[string]*Local),
		},
	}
	for _, stmt := range f.Body.Statements {
		if dec, ok := stmt.(*ast.VarDec); ok {
			frame.Scope.AddVarDec(dec)
		}
	}
	return frame
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
	case *ast.FuncDec:
		return c.funcDec(stmt)
	default:
		return fmt.Errorf("cannot compile: %s", p.Body)
	}
}

func (c *Compiler) expr(expr ast.Expr) error {
	switch expr := expr.(type) {
	case *ast.IntLit:
		c.emitf("movl $%d, %%eax", expr.Value)
	case *ast.UnaryOp:
		return c.unaryOp(expr)
	case *ast.BinaryOp:
		return c.binaryOp(expr)
	case *ast.Var:
		return c.variable(expr)
	case *ast.Assign:
		return c.assign(expr)
	case *ast.Ternary:
		return c.ternary(expr)
	default:
		return fmt.Errorf("cannot compile: %s", expr)
	}
	return nil
}

func (c *Compiler) unaryOp(unary *ast.UnaryOp) error {
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

func (c *Compiler) stmt(stmt ast.Stmt) error {
	switch stmt := stmt.(type) {
	case *ast.Ret:
		return c.ret(stmt)
	case *ast.VarDec:
		return c.varDec(stmt)
	case *ast.If:
		return c._if(stmt)
	case *ast.ExprStmt:
		return c.expr(stmt.Expr)
	default:
		return fmt.Errorf("cannot compile: %s", stmt)
	}
}

func (c *Compiler) varDec(dec *ast.VarDec) error {
	if dec.Value != nil {
		if err := c.expr(dec.Value); err != nil {
			return err
		}
	} else {
		c.emitf("movl $0, %%eax")
	}
	loc, err := c.frame().Local(dec.Name)
	if err != nil {
		return err
	}
	loc.Declared = true
	c.emitf("movl %%eax, %d(%%ebp)", loc.Offset)
	return nil
}

func (c *Compiler) ternary(tern *ast.Ternary) error {
	afterThen, end := c.frame().Label(), c.frame().Label()
	if err := c.expr(tern.Condition); err != nil {
		return err
	}
	c.emitf("cmpl $0, %%eax")
	c.emitf("je %s", afterThen)
	if err := c.expr(tern.Then); err != nil {
		return err
	}
	c.emitf("jmp %s", end)
	c.emitf("%s:", afterThen)
	if err := c.expr(tern.Else); err != nil {
		return err
	}
	c.emitf("%s:", end)
	return nil
}

func (c *Compiler) _if(ife *ast.If) error {
	afterThen, end := c.frame().Label(), c.frame().Label()
	if err := c.expr(ife.Condition); err != nil {
		return err
	}
	c.emitf("cmpl $0, %%eax")
	c.emitf("je %s", afterThen)
	if err := c.stmt(ife.Then); err != nil {
		return err
	}
	c.emitf("jmp %s", end)
	c.emitf("%s:", afterThen)
	if ife.Else != nil {
		if err := c.stmt(ife.Else); err != nil {
			return err
		}
	}
	c.emitf("%s:", end)
	return nil
}

func (c *Compiler) assign(assign *ast.Assign) error {
	if err := c.expr(assign.Value); err != nil {
		return err
	}
	loc, err := c.frame().Local(assign.Var.Name)
	if err != nil {
		return err
	}
	c.emitf("movl %%eax, %d(%%ebp)", loc.Offset)
	return nil
}

func (c *Compiler) variable(v *ast.Var) error {
	loc, err := c.frame().Local(v.Name)
	if err != nil {
		return err
	}
	if !loc.Declared {
		return fmt.Errorf("use before declaration: %s", v.Name)
	}
	c.emitf("movl %d(%%ebp), %%eax", loc.Offset)
	return nil
}

func (c *Compiler) ret(ret *ast.Ret) error {
	if err := c.expr(ret.Value); err != nil {
		return err
	}
	c.prologue()
	return nil
}

func (c *Compiler) binaryOp(binary *ast.BinaryOp) error {
	if err := c.expr(binary.Left); err != nil {
		return err
	}
	c.emitf("push %%eax\n")
	if err := c.expr(binary.Right); err != nil {
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

func (c *Compiler) preable(name string) {
	c.emitf(".globl %s", name)
	c.emitf("%s:", name)
	c.emitf("push %%ebp")
	c.emitf("movl %%esp, %%ebp")
}

func (c *Compiler) prologue() {
	c.emitf("movl %%ebp, %%esp")
	c.emitf("pop %%ebp")
	c.emitf("ret")
}

func (c *Compiler) funcDec(f *ast.FuncDec) error {

	frame := c.newFrame(f)
	c.framePush(frame)
	defer c.framePop()

	c.preable(frame.Entry)
	c.emitf("subl $%d, %%esp", -frame.Scope.Offset)

	for _, stmt := range f.Body.Statements {
		if err := c.stmt(stmt); err != nil {
			return err
		}
	}

	c.emitf("movl $0, %%eax")
	c.emitf("%s:", frame.Exit)
	c.prologue()
	return nil
}
