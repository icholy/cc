package ast

import (
	"fmt"
	"strings"

	"github.com/icholy/cc/token"
)

type Node interface {
	Token() token.Token
	String() string
}

type Stmt interface {
	Node
	stmtNode()
}

type Expr interface {
	Node
	exprNode()
}

type Program struct {
	Tok  token.Token
	Body Stmt
}

func (p *Program) stmtNode()          {}
func (p *Program) Token() token.Token { return p.Tok }
func (p *Program) String() string     { return p.Body.String() }

type IntLit struct {
	Tok   token.Token
	Value int
}

func (i *IntLit) exprNode()          {}
func (i *IntLit) Token() token.Token { return i.Tok }
func (i *IntLit) String() string     { return fmt.Sprintf("IntLit(%d)", i.Value) }

type BinaryOp struct {
	Tok   token.Token
	Op    string
	Left  Expr
	Right Expr
}

func (b *BinaryOp) exprNode()          {}
func (b *BinaryOp) Token() token.Token { return b.Tok }
func (b *BinaryOp) String() string     { return fmt.Sprintf("BinaryOp(%s %s %s)", b.Left, b.Op, b.Right) }

type UnaryOp struct {
	Tok   token.Token
	Op    string
	Value Expr
}

func (u *UnaryOp) exprNode()          {}
func (u *UnaryOp) Token() token.Token { return u.Tok }
func (u *UnaryOp) String() string     { return fmt.Sprintf("UnaryOp(%s %s)", u.Op, u.Value) }

type Assign struct {
	Tok   token.Token
	Var   *Var
	Value Expr
}

func (a *Assign) exprNode()          {}
func (a *Assign) Token() token.Token { return a.Tok }
func (a *Assign) String() string     { return fmt.Sprintf("Assign(%s = %s)", a.Var, a.Value) }

type VarDec struct {
	Tok   token.Token
	Name  string
	Value Expr
}

func (v *VarDec) stmtNode()          {}
func (v *VarDec) Token() token.Token { return v.Tok }
func (v *VarDec) String() string {
	if v.Value == nil {
		return fmt.Sprintf("VarDec(%s)", v.Name)
	}
	return fmt.Sprintf("VarDec(%s = %s)", v.Name, v.Value)
}

type Var struct {
	Tok  token.Token
	Name string
}

func (v *Var) exprNode()          {}
func (v *Var) Token() token.Token { return v.Tok }
func (v *Var) String() string     { return v.Name }

type If struct {
	Tok       token.Token
	Condition Expr
	Then      Stmt
	Else      Stmt
}

func (i *If) stmtNode()          {}
func (i *If) Token() token.Token { return i.Tok }
func (i *If) String() string {
	if i.Else == nil {
		return fmt.Sprintf("IF %s THEN %s", i.Condition, i.Then)
	}
	return fmt.Sprintf("IF %s THEN %s ELSE %s", i.Condition, i.Then, i.Else)
}

type FuncDec struct {
	Tok  token.Token
	Name string
	Body *Block
}

func (f *FuncDec) stmtNode()          {}
func (f *FuncDec) Token() token.Token { return f.Tok }
func (f *FuncDec) String() string     { return fmt.Sprintf("FuncDec(%s %s)", f.Name, f.Body) }

type Ret struct {
	Tok   token.Token
	Value Expr
}

func (r *Ret) stmtNode()          {}
func (r *Ret) Token() token.Token { return r.Tok }
func (r *Ret) String() string     { return fmt.Sprintf("Ret(%s)", r.Value) }

type Block struct {
	Tok        token.Token
	Statements []Stmt
}

func (b *Block) stmtNode()          {}
func (b *Block) Token() token.Token { return b.Tok }
func (b *Block) String() string {
	ss := make([]string, len(b.Statements))
	for i, stmt := range b.Statements {
		ss[i] = stmt.String()
	}
	return fmt.Sprintf("Block(%s)", strings.Join(ss, "\n"))
}

type ExprStmt struct {
	Tok  token.Token
	Expr Expr
}

func (e *ExprStmt) stmtNode()          {}
func (e *ExprStmt) Token() token.Token { return e.Tok }
func (e *ExprStmt) String() string     { return fmt.Sprintf("ExprStmt(%s)", e.Expr) }

type Ternary struct {
	Tok       token.Token
	Condition Expr
	Then      Expr
	Else      Expr
}

func (t *Ternary) exprNode()          {}
func (t *Ternary) Token() token.Token { return t.Tok }
func (t *Ternary) String() string {
	return fmt.Sprintf("Ternary(%s ? %s : %s)", t.Condition, t.Then, t.Else)
}

type For struct {
	Tok       token.Token
	Setup     Stmt
	Condition Expr
	Increment Expr
	Body      Stmt
}

func (f *For) stmtNode()          {}
func (f *For) Token() token.Token { return f.Tok }
func (f *For) String() string {
	return fmt.Sprintf("FOR(%s; %s; %s) %s", f.Setup, f.Condition, f.Increment, f.Body)
}

type While struct {
	Tok       token.Token
	Condition Expr
	Body      Stmt
}

func (w *While) stmtNode()          {}
func (w *While) Token() token.Token { return w.Tok }
func (w *While) String() string     { return fmt.Sprintf("WHILE(%s) %s", w.Condition, w.Body) }

type Do struct {
	Tok       token.Token
	Condition Expr
	Body      Stmt
}

func (d *Do) stmtNode()          {}
func (d *Do) Token() token.Token { return d.Tok }
func (d *Do) String() string     { return fmt.Sprintf("DO %s WHILE(%s)", d.Body, d.Condition) }

type Break struct {
	Tok token.Token
}

func (b *Break) stmtNode()          {}
func (b *Break) Token() token.Token { return b.Tok }
func (b *Break) String() string     { return "BREAK" }

type Continue struct {
	Tok token.Token
}

func (c *Continue) stmtNode()          {}
func (c *Continue) Token() token.Token { return c.Tok }
func (c *Continue) String() string     { return "CONTINUE" }

type Null struct {
	Tok token.Token
}

func (n *Null) stmtNode()          {}
func (n *Null) Token() token.Token { return n.Tok }
func (n *Null) String() string     { return "NULL" }
