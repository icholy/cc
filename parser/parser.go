package parser

import (
	"fmt"
	"strconv"

	"github.com/icholy/cc/ast"
	"github.com/icholy/cc/lexer"
	"github.com/icholy/cc/token"
)

type Parser struct {
	peek token.Token
	cur  token.Token
	lex  *lexer.Lexer
}

func Parse(input string) (*ast.Program, error) {
	l := lexer.New(input)
	p := New(l)
	prog, err := p.Parse()
	if err != nil {
		return nil, fmt.Errorf("%s:\n%s", err, l.Context(p.cur))
	}
	return prog, nil
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{lex: l}
	p.next()
	p.next()
	return p
}

func (p *Parser) next() {
	p.cur = p.peek
	p.peek = p.lex.Lex()
}

func (p *Parser) expect(typ token.TokenType) error {
	if !p.cur.Is(typ) {
		return fmt.Errorf("invalid token: %s, expecting %s", p.cur, typ)
	}
	p.next()
	return nil
}

func (p *Parser) Parse() (*ast.Program, error) {
	p.trace("Parse")
	prog := &ast.Program{Tok: p.cur}
	fn, err := p.funcDec()
	if err != nil {
		return nil, err
	}
	prog.Body = fn
	if err := p.expect(token.EOF); err != nil {
		return nil, err
	}
	return prog, nil
}

func (p *Parser) blockStmt() (ast.Stmt, error) {
	p.trace("BlockStmt")
	switch {
	case p.cur.Is(token.INT_TYPE):
		return p.varDec()
	default:
		return p.stmt()
	}
}

func (p *Parser) block() (*ast.Block, error) {
	p.trace("Block")
	if err := p.expect(token.LBRACE); err != nil {
		return nil, err
	}
	block := &ast.Block{Tok: p.cur}
	for !p.cur.OneOf(token.RBRACE, token.EOF) {
		stmt, err := p.blockStmt()
		if err != nil {
			return nil, err
		}
		block.Statements = append(block.Statements, stmt)
	}
	if err := p.expect(token.RBRACE); err != nil {
		return nil, err
	}
	return block, nil
}

func (p *Parser) funcDec() (*ast.FuncDec, error) {
	p.trace("FuncDec")
	fd := &ast.FuncDec{Tok: p.cur}
	if err := p.expect(token.INT_TYPE); err != nil {
		return nil, err
	}
	fd.Name = p.cur.Text
	if err := p.expect(token.IDENT); err != nil {
		return nil, err
	}
	if err := p.expect(token.LPAREN); err != nil {
		return nil, err
	}
	if err := p.expect(token.RPAREN); err != nil {
		return nil, err
	}
	block, err := p.block()
	if err != nil {
		return nil, err
	}
	fd.Body = block
	return fd, nil
}

func (p *Parser) trace(s string) {
	// fmt.Println(s, p.cur)
}

func (p *Parser) stmt() (ast.Stmt, error) {
	p.trace("Stmt")
	switch {
	case p.cur.Is(token.LBRACE):
		return p.block()
	case p.cur.Is(token.IF):
		return p._if()
	case p.cur.Is(token.RETURN):
		return p.ret()
	case p.cur.Is(token.WHILE):
		return p.whileLoop()
	case p.cur.Is(token.DO):
		return p.doLoop()
	default:
		return p.exprStmt()
	}
}

func (p *Parser) varDec() (*ast.VarDec, error) {
	p.trace("VarDec")
	decl := &ast.VarDec{Tok: p.cur}
	if err := p.expect(token.INT_TYPE); err != nil {
		return nil, err
	}
	decl.Name = p.cur.Text
	if err := p.expect(token.IDENT); err != nil {
		return nil, err
	}
	if p.cur.Is(token.ASSIGN) {
		p.next()
		value, err := p.expr()
		if err != nil {
			return nil, err
		}
		decl.Value = value
	}
	if err := p.expect(token.SEMICOLON); err != nil {
		return nil, err
	}
	return decl, nil
}

func (p *Parser) exprStmt() (*ast.ExprStmt, error) {
	p.trace("ExprStmt")
	stmt := &ast.ExprStmt{Tok: p.cur}
	expr, err := p.expr()
	if err != nil {
		return nil, err
	}
	stmt.Expr = expr
	if err := p.expect(token.SEMICOLON); err != nil {
		return nil, err
	}
	return stmt, nil
}

func (p *Parser) ret() (*ast.Ret, error) {
	p.trace("Ret")
	ret := &ast.Ret{Tok: p.cur}
	if err := p.expect(token.RETURN); err != nil {
		return nil, err
	}
	expr, err := p.expr()
	if err != nil {
		return nil, err
	}
	ret.Value = expr
	if err := p.expect(token.SEMICOLON); err != nil {
		return nil, err
	}
	return ret, nil
}

func (p *Parser) whileLoop() (*ast.While, error) {
	w := &ast.While{Tok: p.cur}
	if err := p.expect(token.WHILE); err != nil {
		return nil, err
	}
	if err := p.expect(token.LPAREN); err != nil {
		return nil, err
	}
	var err error
	w.Condition, err = p.expr()
	if err != nil {
		return nil, err
	}
	if err := p.expect(token.RPAREN); err != nil {
		return nil, err
	}
	w.Body, err = p.stmt()
	if err != nil {
		return nil, err
	}
	return w, nil
}

func (p *Parser) doLoop() (*ast.Do, error) {
	d := &ast.Do{Tok: p.cur}
	if err := p.expect(token.DO); err != nil {
		return nil, err
	}
	var err error
	d.Body, err = p.stmt()
	if err != nil {
		return nil, err
	}
	if err := p.expect(token.WHILE); err != nil {
		return nil, err
	}
	if err := p.expect(token.LPAREN); err != nil {
		return nil, err
	}
	d.Condition, err = p.expr()
	if err != nil {
		return nil, err
	}
	if err := p.expect(token.RPAREN); err != nil {
		return nil, err
	}
	if err := p.expect(token.SEMICOLON); err != nil {
		return nil, err
	}
	return d, nil
}

func (p *Parser) _if() (*ast.If, error) {
	p.trace("If")
	stmt := &ast.If{Tok: p.cur}
	if err := p.expect(token.IF); err != nil {
		return nil, err
	}
	if err := p.expect(token.LPAREN); err != nil {
		return nil, err
	}
	var err error
	stmt.Condition, err = p.expr()
	if err != nil {
		return nil, err
	}
	if err := p.expect(token.RPAREN); err != nil {
		return nil, err
	}
	stmt.Then, err = p.stmt()
	if err != nil {
		return nil, err
	}
	if p.cur.Is(token.ELSE) {
		p.next()
		stmt.Else, err = p.stmt()
		if err != nil {
			return nil, err
		}
	}
	return stmt, nil
}

func (p *Parser) expr() (ast.Expr, error) {
	p.trace("Expr")
	return p.assign()
}

func (p *Parser) variable() (*ast.Var, error) {
	p.trace("Var")
	v := &ast.Var{Tok: p.cur, Name: p.cur.Text}
	if err := p.expect(token.IDENT); err != nil {
		return nil, err
	}
	return v, nil
}

func (p *Parser) assign() (ast.Expr, error) {
	p.trace("Assignment")
	expr, err := p.ternary()
	if err != nil {
		return nil, err
	}
	if !p.cur.Is(token.ASSIGN) {
		return expr, nil
	}

	assign := &ast.Assign{Tok: p.cur}
	p.next()
	v, ok := expr.(*ast.Var)
	if !ok {
		return nil, fmt.Errorf("cannot assign to: %s", expr)
	}
	assign.Var = v
	assign.Value, err = p.expr()
	return assign, nil
}

func (p *Parser) ternary() (ast.Expr, error) {
	expr, err := p.or()
	if err != nil {
		return nil, err
	}
	if !p.cur.Is(token.QUESTION) {
		return expr, nil
	}
	tern := &ast.Ternary{Tok: p.cur, Condition: expr}
	p.next()
	tern.Then, err = p.expr()
	if err != nil {
		return nil, err
	}
	if err := p.expect(token.COLON); err != nil {
		return nil, err
	}
	tern.Else, err = p.ternary()
	if err != nil {
		return nil, err
	}
	return tern, nil
}

func (p *Parser) binary(parse func() (ast.Expr, error), types ...token.TokenType) (ast.Expr, error) {
	expr, err := parse()
	if err != nil {
		return nil, err
	}
	for p.cur.OneOf(types...) {
		bin := &ast.BinaryOp{Tok: p.cur, Op: p.cur.Text, Left: expr}
		p.next()
		right, err := parse()
		if err != nil {
			return nil, err
		}
		bin.Right = right
		expr = bin
	}
	return expr, nil
}

func (p *Parser) or() (ast.Expr, error) {
	return p.binary(p.and, token.OR)
}

func (p *Parser) and() (ast.Expr, error) {
	return p.binary(p.equality, token.AND)
}

func (p *Parser) equality() (ast.Expr, error) {
	return p.binary(p.relational, token.EQ, token.NE)
}

func (p *Parser) relational() (ast.Expr, error) {
	return p.binary(p.additive, token.GT, token.LT, token.GT_EQ, token.LT_EQ)
}

func (p *Parser) additive() (ast.Expr, error) {
	return p.binary(p.term, token.PLUS, token.MINUS)
}

func (p *Parser) term() (ast.Expr, error) {
	return p.binary(p.factor, token.ASTERISK, token.SLASH, token.PERCENT)
}

func (p *Parser) factor() (ast.Expr, error) {
	p.trace("factor")
	switch {
	case p.cur.Is(token.IDENT):
		return p.variable()
	case p.cur.Is(token.INT_LIT):
		return p.intLit()
	case p.cur.Is(token.LPAREN):
		return p.grouped()
	case p.isUnaryOp(p.cur):
		return p.unaryOp()
	default:
		return nil, fmt.Errorf("invalid factor: %s", p.cur)
	}
}

func (p *Parser) isUnaryOp(tok token.Token) bool {
	switch tok.Type {
	case token.BANG, token.MINUS, token.TILDA:
		return true
	default:
		return false
	}
}

func (p *Parser) unaryOp() (ast.Expr, error) {
	p.trace("UnaryOp")
	if !p.isUnaryOp(p.cur) {
		return nil, fmt.Errorf("invalid unary op: %s", p.cur)
	}
	unary := &ast.UnaryOp{Tok: p.cur, Op: p.cur.Text}
	p.next()
	expr, err := p.factor()
	if err != nil {
		return nil, err
	}
	unary.Value = expr
	return unary, nil
}

func (p *Parser) grouped() (ast.Expr, error) {
	p.trace("Grouped")
	if err := p.expect(token.LPAREN); err != nil {
		return nil, err
	}
	expr, err := p.expr()
	if err != nil {
		return nil, err
	}
	if err := p.expect(token.RPAREN); err != nil {
		return nil, err
	}
	return expr, nil
}

func (p *Parser) intLit() (*ast.IntLit, error) {
	p.trace("IntLit")
	lit := &ast.IntLit{Tok: p.cur}
	value, err := strconv.Atoi(p.cur.Text)
	if err != nil {
		return nil, err
	}
	lit.Value = value
	p.next()
	return lit, nil
}
