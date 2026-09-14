package parser

import (
	"fmt"
	"strconv"

	"github.com/yukkes/miplang/internal/lexer"
)

type pstate struct {
	tokens []lexer.Token
	pos    int
}

func Parse(src string) (*Model, error) {
	toks, err := lexer.Lex(src)
	if err != nil {
		return nil, err
	}
	p := &pstate{tokens: toks}
	m := &Model{}
	for p.peek().Kind != lexer.EOF {
		t := p.peek()
		if t.Kind != lexer.Keyword {
			return nil, p.err(t, "expected declaration")
		}
		switch t.Literal {
		case "set":
			d, e := p.parseSet()
			if e != nil {
				return nil, e
			}
			m.Sets = append(m.Sets, d)
		case "param":
			d, e := p.parseParam()
			if e != nil {
				return nil, e
			}
			m.Params = append(m.Params, d)
		case "var":
			d, e := p.parseVar()
			if e != nil {
				return nil, e
			}
			m.Vars = append(m.Vars, d)
		case "constraint":
			d, e := p.parseConstraint()
			if e != nil {
				return nil, e
			}
			m.Constraints = append(m.Constraints, d)
		case "minimize", "maximize":
			d, e := p.parseObjective()
			if e != nil {
				return nil, e
			}
			m.Objectives = append(m.Objectives, d)
		default:
			return nil, p.err(t, "unsupported declaration %q", t.Literal)
		}
	}
	return m, nil
}

func (p *pstate) parseSet() (SetDecl, error) {
	p.next()
	name, err := p.ident()
	if err != nil {
		return SetDecl{}, err
	}
	if err = p.want(lexer.Semicolon); err != nil {
		return SetDecl{}, err
	}
	return SetDecl{Name: name}, nil
}
func (p *pstate) parseParam() (ParamDecl, error) {
	p.next()
	name, err := p.ident()
	if err != nil {
		return ParamDecl{}, err
	}
	domain, err := p.optionalDomain()
	if err != nil {
		return ParamDecl{}, err
	}
	if err = p.want(lexer.Semicolon); err != nil {
		return ParamDecl{}, err
	}
	return ParamDecl{Name: name, Domain: domain}, nil
}
func (p *pstate) parseVar() (VarDecl, error) {
	p.next()
	name, err := p.ident()
	if err != nil {
		return VarDecl{}, err
	}
	domain, err := p.optionalDomain()
	if err != nil {
		return VarDecl{}, err
	}
	d := VarDecl{Name: name, Domain: domain}
	if p.isKeyword("integer") {
		p.next()
		d.Integer = true
	}
	if p.isKeyword("binary") {
		p.next()
		d.Binary, d.Integer = true, true
	}
	for p.peek().Kind == lexer.GreaterEqual || p.peek().Kind == lexer.LessEqual {
		op := p.next().Kind
		n, e := p.number()
		if e != nil {
			return VarDecl{}, e
		}
		if op == lexer.GreaterEqual {
			d.Lower = &n
		} else {
			d.Upper = &n
		}
	}
	if err = p.want(lexer.Semicolon); err != nil {
		return VarDecl{}, err
	}
	return d, nil
}
func (p *pstate) parseConstraint() (ConstraintDecl, error) {
	p.next()
	name, err := p.ident()
	if err != nil {
		return ConstraintDecl{}, err
	}
	var idx *Iterator
	if p.peek().Kind == lexer.LBracket {
		it, e := p.iterator()
		if e != nil {
			return ConstraintDecl{}, e
		}
		idx = &it
	}
	if err = p.want(lexer.Colon); err != nil {
		return ConstraintDecl{}, err
	}
	left, err := p.expr()
	if err != nil {
		return ConstraintDecl{}, err
	}
	opTok := p.next()
	if opTok.Kind != lexer.LessEqual && opTok.Kind != lexer.GreaterEqual && opTok.Kind != lexer.EqualEqual {
		return ConstraintDecl{}, p.err(opTok, "expected relation operator")
	}
	right, err := p.expr()
	if err != nil {
		return ConstraintDecl{}, err
	}
	if err = p.want(lexer.Semicolon); err != nil {
		return ConstraintDecl{}, err
	}
	return ConstraintDecl{Name: name, Index: idx, Left: left, Op: opTok.Literal, Right: right}, nil
}
func (p *pstate) parseObjective() (ObjectiveDecl, error) {
	sense := p.next().Literal
	name, err := p.ident()
	if err != nil {
		return ObjectiveDecl{}, err
	}
	if err = p.want(lexer.Colon); err != nil {
		return ObjectiveDecl{}, err
	}
	e, err := p.expr()
	if err != nil {
		return ObjectiveDecl{}, err
	}
	if err = p.want(lexer.Semicolon); err != nil {
		return ObjectiveDecl{}, err
	}
	return ObjectiveDecl{Name: name, Sense: sense, Expr: e}, nil
}
func (p *pstate) optionalDomain() (string, error) {
	if p.peek().Kind != lexer.LBracket {
		return "", nil
	}
	p.next()
	name, err := p.ident()
	if err != nil {
		return "", err
	}
	if err = p.want(lexer.RBracket); err != nil {
		return "", err
	}
	return name, nil
}
func (p *pstate) iterator() (Iterator, error) {
	if err := p.want(lexer.LBracket); err != nil {
		return Iterator{}, err
	}
	name, err := p.ident()
	if err != nil {
		return Iterator{}, err
	}
	if !p.isKeyword("in") {
		return Iterator{}, p.err(p.peek(), "expected 'in'")
	}
	p.next()
	set, err := p.ident()
	if err != nil {
		return Iterator{}, err
	}
	if err = p.want(lexer.RBracket); err != nil {
		return Iterator{}, err
	}
	return Iterator{Name: name, Set: set}, nil
}
func (p *pstate) expr() (Expr, error) { return p.additive() }
func (p *pstate) additive() (Expr, error) {
	left, err := p.multiplicative()
	if err != nil {
		return nil, err
	}
	for p.peek().Kind == lexer.Plus || p.peek().Kind == lexer.Minus {
		op := p.next().Literal
		right, e := p.multiplicative()
		if e != nil {
			return nil, e
		}
		left = BinaryExpr{Op: op, Left: left, Right: right}
	}
	return left, nil
}
func (p *pstate) multiplicative() (Expr, error) {
	left, err := p.unary()
	if err != nil {
		return nil, err
	}
	for p.peek().Kind == lexer.Star {
		op := p.next().Literal
		right, e := p.unary()
		if e != nil {
			return nil, e
		}
		left = BinaryExpr{Op: op, Left: left, Right: right}
	}
	return left, nil
}
func (p *pstate) unary() (Expr, error) {
	if p.peek().Kind == lexer.Minus {
		p.next()
		v, err := p.unary()
		if err != nil {
			return nil, err
		}
		return UnaryExpr{Op: "-", Value: v}, nil
	}
	return p.primary()
}
func (p *pstate) primary() (Expr, error) {
	t := p.peek()
	if t.Kind == lexer.Number {
		n, err := p.number()
		if err != nil {
			return nil, err
		}
		return NumberExpr{Value: n}, nil
	}
	if p.isKeyword("sum") {
		p.next()
		it, err := p.iterator()
		if err != nil {
			return nil, err
		}
		if err = p.want(lexer.LBrace); err != nil {
			return nil, err
		}
		body, err := p.expr()
		if err != nil {
			return nil, err
		}
		if err = p.want(lexer.RBrace); err != nil {
			return nil, err
		}
		return SumExpr{Index: it, Body: body}, nil
	}
	if t.Kind == lexer.Identifier {
		name := p.next().Literal
		idx := ""
		if p.peek().Kind == lexer.LBracket {
			p.next()
			var err error
			idx, err = p.ident()
			if err != nil {
				return nil, err
			}
			if err = p.want(lexer.RBracket); err != nil {
				return nil, err
			}
		}
		return RefExpr{Name: name, Index: idx}, nil
	}
	if t.Kind == lexer.LParen {
		p.next()
		e, err := p.expr()
		if err != nil {
			return nil, err
		}
		if err = p.want(lexer.RParen); err != nil {
			return nil, err
		}
		return e, nil
	}
	return nil, p.err(t, "expected expression")
}
func (p *pstate) number() (float64, error) {
	t := p.next()
	if t.Kind != lexer.Number {
		return 0, p.err(t, "expected number")
	}
	v, err := strconv.ParseFloat(t.Literal, 64)
	if err != nil {
		return 0, p.err(t, "invalid number")
	}
	return v, nil
}
func (p *pstate) ident() (string, error) {
	t := p.next()
	if t.Kind != lexer.Identifier {
		return "", p.err(t, "expected identifier")
	}
	return t.Literal, nil
}
func (p *pstate) want(k lexer.Kind) error {
	t := p.next()
	if t.Kind != k {
		return p.err(t, "expected %s", k)
	}
	return nil
}
func (p *pstate) isKeyword(s string) bool {
	t := p.peek()
	return t.Kind == lexer.Keyword && t.Literal == s
}
func (p *pstate) peek() lexer.Token { return p.tokens[p.pos] }
func (p *pstate) next() lexer.Token {
	t := p.tokens[p.pos]
	if p.pos < len(p.tokens)-1 {
		p.pos++
	}
	return t
}
func (p *pstate) err(t lexer.Token, format string, args ...any) error {
	return fmt.Errorf("%d:%d: %s", t.Line, t.Column, fmt.Sprintf(format, args...))
}
