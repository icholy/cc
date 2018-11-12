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

type IntLiteral struct {
	Tok   token.Token
	Value int
}

func (i *IntLiteral) exprNode()          {}
func (i *IntLiteral) Token() token.Token { return i.Tok }
func (i *IntLiteral) String() string     { return fmt.Sprintf("IntLiteral(%d)", i.Value) }

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

type Assignment struct {
	Tok   token.Token
	Var   *Var
	Value Expr
}

func (a *Assignment) exprNode()          {}
func (a *Assignment) Token() token.Token { return a.Tok }
func (a *Assignment) String() string     { return fmt.Sprintf("Assignment(%s = %s)", a.Var, a.Value) }

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
	Body      *Block
	Else      *Block
}

func (i *If) stmtNode()          {}
func (i *If) Token() token.Token { return i.Tok }
func (i *If) String() string {
	if i.Else == nil {
		return fmt.Sprintf("IF %s THEN %s", i.Condition, i.Body)
	}
	return fmt.Sprintf("IF %s THEN %s ELSE %s", i.Condition, i.Body, i.Else)
}

type Function struct {
	Tok  token.Token
	Name string
	Body *Block
}

func (f *Function) stmtNode()          {}
func (f *Function) Token() token.Token { return f.Tok }
func (f *Function) String() string     { return fmt.Sprintf("Function(%s %s)", f.Name, f.Body) }

type Return struct {
	Tok   token.Token
	Value Expr
}

func (r *Return) stmtNode()          {}
func (r *Return) Token() token.Token { return r.Tok }
func (r *Return) String() string     { return fmt.Sprintf("Return(%s)", r.Value) }

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
