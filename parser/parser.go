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
	return p.Parse()
}

func New(l *lexer.Lexer) *Parser {
	return &Parser{lex: l, peek: l.Lex()}
}

func (p *Parser) next() {
	p.cur = p.peek
	p.peek = p.lex.Lex()
}

func (p *Parser) expectPeek(typ token.TokenType) error {
	if !p.peek.Is(typ) {
		return fmt.Errorf("invalid token: %s, expecting %s", p.peek, typ)
	}
	p.next()
	return nil
}

func (p *Parser) Parse() (*ast.Program, error) {
	prog := &ast.Program{Tok: p.cur}
	fn, err := p.parseFunction()
	if err != nil {
		return nil, err
	}
	prog.Body = fn
	return prog, nil
}

func (p *Parser) parseFunction() (*ast.Function, error) {
	if err := p.expectPeek(token.INT_TYPE); err != nil {
		return nil, err
	}
	fn := &ast.Function{Tok: p.cur}
	if err := p.expectPeek(token.IDENT); err != nil {
		return nil, err
	}
	fn.Name = p.cur.Text
	if err := p.expectPeek(token.LPAREN); err != nil {
		return nil, err
	}
	if err := p.expectPeek(token.RPAREN); err != nil {
		return nil, err
	}
	if err := p.expectPeek(token.LBRACE); err != nil {
		return nil, err
	}
	body, err := p.parseStmt()
	if err != nil {
		return nil, err
	}
	fn.Body = body
	if err := p.expectPeek(token.RBRACE); err != nil {
		return nil, err
	}
	return fn, nil
}

func (p *Parser) parseStmt() (ast.Stmt, error) {
	if err := p.expectPeek(token.RETURN); err != nil {
		return nil, err
	}
	ret := &ast.Return{Tok: p.cur}
	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	ret.Value = expr
	if err := p.expectPeek(token.SEMICOLON); err != nil {
		return nil, err
	}
	return ret, nil
}

func (p *Parser) parseExpr() (ast.Expr, error) {
	switch {
	case p.peek.Is(token.INT_LIT):
		return p.parseIntLit()
	default:
		return nil, fmt.Errorf("unexpected: %s", p.peek)
	}
}

func (p *Parser) parseUnaryOp() (ast.Expr, error) {
	p.next()
	switch p.cur.Type {
	case token.BANG, token.MINUS, token.TILDA:
	default:
		return nil, fmt.Errorf("invalid unary op: %s", p.cur)
	}
	unary := &ast.UnaryOp{Tok: p.cur, Op: p.cur.Text}
	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	unary.Value = expr
	return unary, nil
}

func (p *Parser) parseIntLit() (ast.Expr, error) {
	if err := p.expectPeek(token.INT_LIT); err != nil {
		return nil, err
	}
	lit := &ast.IntLiteral{Tok: p.cur}
	value, err := strconv.Atoi(p.cur.Text)
	if err != nil {
		return nil, err
	}
	lit.Value = value
	return lit, nil
}
