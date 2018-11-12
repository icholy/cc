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

func (p *Parser) peekIsOneOf(tt ...token.TokenType) bool {
	for _, t := range tt {
		if p.peek.Is(t) {
			return true
		}
	}
	return false
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

func (p *Parser) parseBlock() (*ast.Block, error) {
	if err := p.expectPeek(token.LBRACE); err != nil {
		return nil, err
	}
	block := &ast.Block{Tok: p.cur}
	for !p.peekIsOneOf(token.RBRACE, token.EOF) {
		stmt, err := p.parseStmt()
		if err != nil {
			return nil, err
		}
		block.Statements = append(block.Statements, stmt)
	}
	if err := p.expectPeek(token.RBRACE); err != nil {
		return nil, err
	}
	return block, nil
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
	block, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	fn.Body = block
	return fn, nil
}

func (p *Parser) parseStmt() (ast.Stmt, error) {
	switch {
	case p.peek.Is(token.INT_TYPE):
		return p.parseVarDec()
	case p.peek.Is(token.RETURN):
		return p.parseReturn()
	default:
		return p.parseExprStmt()
	}
}

func (p *Parser) parseVarDec() (ast.Stmt, error) {
	decl := &ast.VarDec{Tok: p.peek}
	if err := p.expectPeek(token.INT_TYPE); err != nil {
		return nil, err
	}
	if err := p.expectPeek(token.IDENT); err != nil {
		return nil, err
	}
	decl.Name = p.cur.Text
	if p.peek.Is(token.ASSIGN) {
		p.next()
		value, err := p.parseIntLit()
		if err != nil {
			return nil, err
		}
		decl.Value = value
	}
	if err := p.expectPeek(token.SEMICOLON); err != nil {
		return nil, err
	}
	return decl, nil
}

func (p *Parser) parseExprStmt() (ast.Stmt, error) {
	stmt := &ast.ExprStmt{Tok: p.peek}
	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	stmt.Expr = expr
	if err := p.expectPeek(token.SEMICOLON); err != nil {
		return nil, err
	}
	return stmt, nil
}

func (p *Parser) parseReturn() (ast.Stmt, error) {
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
	case p.peek.Is(token.IDENT):
		return p.parseAssign()
	default:
		return p.parseOr()
	}
}

func (p *Parser) parseVar() (*ast.Var, error) {
	if err := p.expectPeek(token.IDENT); err != nil {
		return nil, err
	}
	return &ast.Var{Tok: p.cur, Name: p.cur.Text}, nil
}

func (p *Parser) parseAssign() (ast.Expr, error) {
	v, err := p.parseVar()
	if err != nil {
		return nil, err
	}
	expr := &ast.Assignment{Tok: p.cur, Var: v}
	if err := p.expectPeek(token.ASSIGN); err != nil {
		return nil, err
	}
	value, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	expr.Value = value
	return expr, nil
}

func (p *Parser) parseOr() (ast.Expr, error) {
	return p.parseBinary(p.parseAnd, token.OR)
}

func (p *Parser) parseAnd() (ast.Expr, error) {
	return p.parseBinary(p.parseEquality, token.AND)
}

func (p *Parser) parseEquality() (ast.Expr, error) {
	return p.parseBinary(p.parseRelational, token.EQ, token.NE)
}

func (p *Parser) parseRelational() (ast.Expr, error) {
	return p.parseBinary(p.parseAdditive, token.GT, token.LT, token.GT_EQ, token.LT_EQ)
}

func (p *Parser) parseAdditive() (ast.Expr, error) {
	return p.parseBinary(p.parseTerm, token.PLUS, token.MINUS)
}

func (p *Parser) parseTerm() (ast.Expr, error) {
	return p.parseBinary(p.parseFactor, token.ASTERISK, token.SLASH)
}

func (p *Parser) parseFactor() (ast.Expr, error) {
	switch {
	case p.peek.Is(token.IDENT):
		return p.parseVar()
	case p.peek.Is(token.INT_LIT):
		return p.parseIntLit()
	case p.peek.Is(token.LPAREN):
		return p.parseGrouped()
	case p.isUnaryOp(p.peek):
		return p.parseUnaryOp()
	default:
		return nil, fmt.Errorf("invalid factor: %s", p.cur)
	}
}

func (p *Parser) parseBinary(parse func() (ast.Expr, error), types ...token.TokenType) (ast.Expr, error) {
	expr, err := parse()
	if err != nil {
		return nil, err
	}
	for p.peekIsOneOf(types...) {
		p.next()
		bin := &ast.BinaryOp{Tok: p.cur, Op: p.cur.Text, Left: expr}
		right, err := parse()
		if err != nil {
			return nil, err
		}
		bin.Right = right
		expr = bin
	}
	return expr, nil
}

func (p *Parser) isUnaryOp(tok token.Token) bool {
	switch tok.Type {
	case token.BANG, token.MINUS, token.TILDA:
		return true
	default:
		return false
	}
}

func (p *Parser) parseUnaryOp() (ast.Expr, error) {
	p.next()
	if !p.isUnaryOp(p.cur) {
		return nil, fmt.Errorf("invalid unary op: %s", p.cur)
	}
	unary := &ast.UnaryOp{Tok: p.cur, Op: p.cur.Text}
	expr, err := p.parseFactor()
	if err != nil {
		return nil, err
	}
	unary.Value = expr
	return unary, nil
}

func (p *Parser) parseGrouped() (ast.Expr, error) {
	if err := p.expectPeek(token.LPAREN); err != nil {
		return nil, err
	}
	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if err := p.expectPeek(token.RPAREN); err != nil {
		return nil, err
	}
	return expr, nil
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
