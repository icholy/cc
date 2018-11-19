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
	scope  *Scope
	funcs  map[string]*ast.FuncDec
	labels int
}

func New() *Compiler {
	return &Compiler{
		asm:   &strings.Builder{},
		funcs: make(map[string]*ast.FuncDec),
	}
}

type Local struct {
	Name     string
	Declared bool
	Offset   int
}

type Loop struct {
	Break, Continue string
}

type Scope struct {
	Parent *Scope
	Offset int
	Locals map[string]*Local
	Loop   *Loop
}

func (s *Scope) AddParam(index int, name string) error {
	if _, ok := s.Locals[name]; ok {
		return fmt.Errorf("duplicate parameter name: %s", name)
	}
	s.Locals[name] = &Local{
		Name:     name,
		Declared: true,
		Offset:   (index + 1) * 4,
	}
	return nil
}

func (s *Scope) FindLoop() (*Loop, error) {
	if s.Loop != nil {
		return s.Loop, nil
	}
	if s.Parent == nil {
		return nil, fmt.Errorf("not inside loop")
	}
	return s.Parent.FindLoop()
}

func (s *Scope) DeclaredLocal(name string) (*Local, error) {
	if loc, ok := s.Locals[name]; ok && loc.Declared {
		return loc, nil
	}
	if s.Parent != nil {
		return s.Parent.Local(name)
	}
	return nil, fmt.Errorf("undefined: %s", name)
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

func (s *Scope) TotalOffset() int {
	if s.Parent == nil {
		return s.Offset
	}
	return s.Parent.TotalOffset() + s.Offset
}

func (s *Scope) Declare(d *ast.VarDec) error {
	s.Offset -= 4
	if _, ok := s.Locals[d.Name]; ok {
		return fmt.Errorf("already declared: %s", d.Name)
	}
	s.Locals[d.Name] = &Local{
		Name:     d.Name,
		Offset:   s.TotalOffset(),
		Declared: false,
	}
	return nil
}

func (c *Compiler) label(name string) string {
	c.labels++
	return fmt.Sprintf("%s_L%d", name, c.labels)
}

func (c *Compiler) enterScope() {
	c.scope = &Scope{
		Parent: c.scope,
		Locals: make(map[string]*Local),
	}
}

func (c *Compiler) enterLoopScope() *Loop {
	c.enterScope()
	c.scope.Loop = &Loop{
		Break:    c.label("break"),
		Continue: c.label("continue"),
	}
	return c.scope.Loop
}

func (c *Compiler) leaveScope() {
	c.scope = c.scope.Parent
}

func (c *Compiler) Assembly() string {
	return c.asm.String()
}

func (c *Compiler) emitf(format string, args ...interface{}) {
	fmt.Fprintf(c.asm, format+"\n", args...)
}

func (c *Compiler) Compile(p *ast.Program) error {
	for _, stmt := range p.Statements {
		switch stmt := stmt.(type) {
		case *ast.FuncDec:
			if err := c.funcDec(stmt); err != nil {
				return err
			}
		default:
			return fmt.Errorf("cannot compile: %s", stmt)
		}
	}
	return nil
}

func (c *Compiler) expr(expr ast.Expr) error {
	switch expr := expr.(type) {
	case *ast.IntLit:
		c.emitf("movl $%d, %%eax", expr.Value)
	case *ast.Null:
		c.emitf("movl $1, %%eax")
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
	case *ast.Call:
		return c.call(expr)
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
	case *ast.Block:
		return c.block(stmt)
	case *ast.ExprStmt:
		return c.expr(stmt.Expr)
	case *ast.While:
		return c.whileLoop(stmt)
	case *ast.Do:
		return c.doLoop(stmt)
	case *ast.For:
		return c.forLoop(stmt)
	case *ast.Break:
		loop, err := c.scope.FindLoop()
		if err != nil {
			return err
		}
		c.emitf("jmp %s", loop.Break)
		return nil
	case *ast.Continue:
		loop, err := c.scope.FindLoop()
		if err != nil {
			return err
		}
		c.emitf("jmp %s", loop.Continue)
		return nil
	default:
		return fmt.Errorf("cannot compile: %s", stmt)
	}
}

func (c *Compiler) forLoop(f *ast.For) error {
	loop := c.enterLoopScope()
	if err := c.allocate(f.Setup); err != nil {
		return err
	}
	skipInc := c.label("for_skip_inc")
	if err := c.stmt(f.Setup); err != nil {
		return err
	}
	c.emitf("jmp %s", skipInc)
	c.emitf("%s:", loop.Continue)
	if err := c.expr(f.Increment); err != nil {
		return err
	}
	c.emitf("%s:", skipInc)
	if err := c.expr(f.Condition); err != nil {
		return err
	}
	c.emitf("cmpl $0, %%eax")
	c.emitf("je %s", loop.Break)
	if err := c.stmt(f.Body); err != nil {
		return err
	}
	c.emitf("jmp %s", loop.Continue)
	c.emitf("%s:", loop.Break)
	c.deallocate()
	c.leaveScope()
	return nil
}

func (c *Compiler) whileLoop(w *ast.While) error {
	loop := c.enterLoopScope()
	c.emitf("%s:", loop.Continue)
	if err := c.expr(w.Condition); err != nil {
		return err
	}
	c.emitf("cmpl $0, %%eax")
	c.emitf("je %s", loop.Break)
	if err := c.stmt(w.Body); err != nil {
		return err
	}
	c.emitf("jmp %s", loop.Continue)
	c.emitf("%s:", loop.Break)
	c.leaveScope()
	return nil
}

func (c *Compiler) doLoop(d *ast.Do) error {
	loop := c.enterLoopScope()
	c.emitf("%s:", loop.Continue)
	if err := c.stmt(d.Body); err != nil {
		return err
	}
	if err := c.expr(d.Condition); err != nil {
		return err
	}
	c.emitf("cmpl $0, %%eax")
	c.emitf("je %s", loop.Break)
	c.emitf("jmp %s", loop.Continue)
	c.emitf("%s:", loop.Break)
	c.leaveScope()
	return nil
}

func (c *Compiler) varDec(dec *ast.VarDec) error {
	if dec.Value != nil {
		if err := c.expr(dec.Value); err != nil {
			return err
		}
	} else {
		c.emitf("movl $0, %%eax")
	}
	loc, err := c.scope.Local(dec.Name)
	if err != nil {
		return err
	}
	loc.Declared = true
	c.emitf("movl %%eax, %d(%%ebp)", loc.Offset)
	return nil
}

func (c *Compiler) ternary(tern *ast.Ternary) error {
	afterThen, end := c.label("tern_after_then"), c.label("tern_end")
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
	afterThen, end := c.label("if_after_then"), c.label("if_end")
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
	loc, err := c.scope.DeclaredLocal(assign.Var.Name)
	if err != nil {
		return err
	}
	c.emitf("movl %%eax, %d(%%ebp)", loc.Offset)
	return nil
}

func (c *Compiler) variable(v *ast.Var) error {
	loc, err := c.scope.DeclaredLocal(v.Name)
	if err != nil {
		return err
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

func (c *Compiler) call(call *ast.Call) error {

	dec, ok := c.funcs[call.Name]
	if !ok {
		return fmt.Errorf("undefined function: %s", call.Name)
	}
	if len(dec.Params) != len(call.Arguments) {
		return fmt.Errorf(
			"bad call, wanted %d arguments, got %d: %s",
			len(dec.Params), len(call.Arguments), call.Name,
		)
	}

	for i := range call.Arguments {
		arg := call.Arguments[len(call.Arguments)-i-1]
		if err := c.expr(arg); err != nil {
			return err
		}
		c.emitf("push %%eax")
	}
	c.emitf("call _%s", call.Name)
	c.emitf("addl $%d, %%esp", len(call.Arguments)*4)
	return nil
}

func (c *Compiler) binaryOp(binary *ast.BinaryOp) error {
	if err := c.expr(binary.Left); err != nil {
		return err
	}
	c.emitf("push %%eax")
	if err := c.expr(binary.Right); err != nil {
		return err
	}
	c.emitf("pop %%ecx")
	switch binary.Op {
	case "+":
		c.emitf("add %%ecx, %%eax")
	case "-":
		c.emitf("sub %%ecx, %%eax")
	case "*":
		c.emitf("imul %%ecx, %%eax")
	case "/":
		c.emitf("xchg %%eax, %%ecx")
		c.emitf("movl $0, %%edx")
		c.emitf("idiv %%ecx, %%eax")
	case "%":
		c.emitf("xchg %%eax, %%ecx")
		c.emitf("movl $0, %%edx")
		c.emitf("idiv %%ecx, %%eax")
		c.emitf("movl %%edx, %%eax")
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
	c.emitf(".globl _%s", name)
	c.emitf("_%s:", name)
	c.emitf("push %%ebp")
	c.emitf("movl %%esp, %%ebp")
}

func (c *Compiler) prologue() {
	c.emitf("movl %%ebp, %%esp")
	c.emitf("pop %%ebp")
	c.emitf("ret")
}

func (c *Compiler) allocate(stmts ...ast.Stmt) error {
	for _, s := range stmts {
		if dec, ok := s.(*ast.VarDec); ok {
			if err := c.scope.Declare(dec); err != nil {
				return err
			}
		}
	}
	c.emitf("subl $%d, %%esp", -c.scope.Offset)
	return nil
}

func (c *Compiler) deallocate() {
	c.emitf("addl $%d, %%esp", -c.scope.Offset)
}

func (c *Compiler) block(b *ast.Block) error {
	c.enterScope()
	if err := c.allocate(b.Statements...); err != nil {
		return err
	}
	for _, stmt := range b.Statements {
		if err := c.stmt(stmt); err != nil {
			return err
		}
	}
	c.deallocate()
	c.leaveScope()
	return nil
}

func (c *Compiler) addFuncDec(f *ast.FuncDec) error {
	prev, ok := c.funcs[f.Name]
	if ok {
		if prev.Body != nil && f.Body != nil {
			return fmt.Errorf("duplicate function definition: %s", f.Name)
		}
		if prev.Body == nil && f.Body == nil {
			return fmt.Errorf("duplicate function prototype: %s", f.Name)
		}
		if len(prev.Params) != len(f.Params) {
			return fmt.Errorf("definition doesn't match prototype: %s", f.Name)
		}
		if f.Body == nil {
			return nil
		}
	}
	c.funcs[f.Name] = f
	return nil
}

func (c *Compiler) funcDec(f *ast.FuncDec) error {
	if err := c.addFuncDec(f); err != nil {
		return err
	}
	if f.Body == nil {
		return nil
	}
	c.enterScope()
	for i, p := range f.Params {
		if err := c.scope.AddParam(i, p); err != nil {
			return err
		}
	}
	c.preable(f.Name)
	if err := c.block(f.Body); err != nil {
		return err
	}
	c.emitf("movl $0, %%eax")
	c.prologue()
	c.leaveScope()
	return nil
}
