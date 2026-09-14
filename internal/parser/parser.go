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
			m.Declarations = append(m.Declarations, d)
		case "param":
			d, e := p.parseParam()
			if e != nil {
				return nil, e
			}
			m.Params = append(m.Params, d)
			m.Declarations = append(m.Declarations, d)
		case "var":
			d, e := p.parseVar()
			if e != nil {
				return nil, e
			}
			m.Vars = append(m.Vars, d)
			m.Declarations = append(m.Declarations, d)
		case "constraint", "subject":
			d, e := p.parseConstraint()
			if e != nil {
				return nil, e
			}
			m.Constraints = append(m.Constraints, d)
			m.Declarations = append(m.Declarations, d)
		case "minimize", "maximize":
			d, e := p.parseObjective()
			if e != nil {
				return nil, e
			}
			m.Objectives = append(m.Objectives, d)
			m.Declarations = append(m.Declarations, d)
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
	d := ParamDecl{Name: name, Domain: domain}
	for p.peek().Kind != lexer.Semicolon {
		if p.peek().Kind == lexer.Comma {
			p.next()
			continue
		}
		if p.isKeyword("integer") {
			p.next()
			d.Integer = true
			continue
		}
		if !isBoundRelation(p.peek().Kind) {
			return ParamDecl{}, p.err(p.peek(), "expected parameter attribute")
		}
		kind, bound, e := p.parseBound()
		if e != nil {
			return ParamDecl{}, e
		}
		if kind == "lower" {
			d.Lower = bound
		} else {
			d.Upper = bound
		}
	}
	if err = p.want(lexer.Semicolon); err != nil {
		return ParamDecl{}, err
	}
	return d, nil
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
	for p.peek().Kind != lexer.Semicolon {
		if p.peek().Kind == lexer.Comma {
			p.next()
			continue
		}
		if p.isKeyword("integer") {
			p.next()
			d.Integer = true
			continue
		}
		if p.isKeyword("binary") {
			p.next()
			d.Binary, d.Integer = true, true
			continue
		}
		if !isBoundRelation(p.peek().Kind) {
			return VarDecl{}, p.err(p.peek(), "expected variable attribute")
		}
		kind, bound, e := p.parseBound()
		if e != nil {
			return VarDecl{}, e
		}
		if kind == "lower" {
			d.Lower = bound
		} else {
			d.Upper = bound
		}
	}
	if err = p.want(lexer.Semicolon); err != nil {
		return VarDecl{}, err
	}
	return d, nil
}

func (p *pstate) parseBound() (string, *Bound, error) {
	t := p.next()
	kind := ""
	strict := false
	switch t.Kind {
	case lexer.GreaterEqual:
		kind = "lower"
	case lexer.Greater:
		kind, strict = "lower", true
	case lexer.LessEqual:
		kind = "upper"
	case lexer.Less:
		kind, strict = "upper", true
	default:
		return "", nil, p.err(t, "expected bound relation")
	}
	e, err := p.expr()
	if err != nil {
		return "", nil, err
	}
	return kind, &Bound{Expr: e, Strict: strict}, nil
}

func (p *pstate) parseConstraint() (ConstraintDecl, error) {
	if p.isKeyword("subject") {
		p.next()
		if !p.isKeyword("to") {
			return ConstraintDecl{}, p.err(p.peek(), "expected 'to' after 'subject'")
		}
		p.next()
	} else {
		p.next()
	}
	name, err := p.ident()
	if err != nil {
		return ConstraintDecl{}, err
	}
	indices, err := p.optionalIndexing()
	if err != nil {
		return ConstraintDecl{}, err
	}
	if err = p.want(lexer.Colon); err != nil {
		return ConstraintDecl{}, err
	}
	first, err := p.expr()
	if err != nil {
		return ConstraintDecl{}, err
	}
	parts := []Expr{first}
	var ops []string
	for isConstraintRelation(p.peek().Kind) {
		op := p.next()
		right, e := p.expr()
		if e != nil {
			return ConstraintDecl{}, e
		}
		literal := op.Literal
		if op.Kind == lexer.EqualEqual {
			literal = "="
		}
		ops = append(ops, literal)
		parts = append(parts, right)
		if len(ops) == 2 {
			break
		}
	}
	if len(ops) == 0 {
		return ConstraintDecl{}, p.err(p.peek(), "expected relation operator")
	}
	if err = p.want(lexer.Semicolon); err != nil {
		return ConstraintDecl{}, err
	}
	return ConstraintDecl{Name: name, Indices: indices, Parts: parts, Ops: ops}, nil
}

func (p *pstate) parseObjective() (ObjectiveDecl, error) {
	sense := p.next().Literal
	name, err := p.ident()
	if err != nil {
		return ObjectiveDecl{}, err
	}
	indices, err := p.optionalAMPLIndexing()
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
	return ObjectiveDecl{Name: name, Sense: sense, Indices: indices, Expr: e}, nil
}

func (p *pstate) optionalDomain() ([]Iterator, error) {
	switch p.peek().Kind {
	case lexer.LBracket:
		return p.indexing(lexer.LBracket, lexer.RBracket)
	case lexer.LBrace:
		return p.indexing(lexer.LBrace, lexer.RBrace)
	default:
		return nil, nil
	}
}

func (p *pstate) optionalIndexing() ([]Iterator, error) {
	switch p.peek().Kind {
	case lexer.LBracket:
		return p.indexing(lexer.LBracket, lexer.RBracket)
	case lexer.LBrace:
		return p.indexing(lexer.LBrace, lexer.RBrace)
	default:
		return nil, nil
	}
}

func (p *pstate) optionalAMPLIndexing() ([]Iterator, error) {
	if p.peek().Kind != lexer.LBrace {
		return nil, nil
	}
	return p.indexing(lexer.LBrace, lexer.RBrace)
}

func (p *pstate) indexing(open, close lexer.Kind) ([]Iterator, error) {
	if err := p.want(open); err != nil {
		return nil, err
	}
	var out []Iterator
	for {
		first, err := p.ident()
		if err != nil {
			return nil, err
		}
		it := Iterator{Set: first}
		if p.isKeyword("in") {
			p.next()
			setName, e := p.ident()
			if e != nil {
				return nil, e
			}
			it.Name, it.Set = first, setName
		}
		out = append(out, it)
		if p.peek().Kind != lexer.Comma {
			break
		}
		p.next()
	}
	if err := p.want(close); err != nil {
		return nil, err
	}
	return out, nil
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
	for p.peek().Kind == lexer.Star || p.peek().Kind == lexer.Slash {
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
		if p.peek().Kind == lexer.LBracket {
			indices, err := p.indexing(lexer.LBracket, lexer.RBracket)
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
			return SumExpr{Indices: indices, Body: body}, nil
		}
		if p.peek().Kind != lexer.LBrace {
			return nil, p.err(p.peek(), "expected sum indexing")
		}
		indices, err := p.indexing(lexer.LBrace, lexer.RBrace)
		if err != nil {
			return nil, err
		}
		body, err := p.multiplicative()
		if err != nil {
			return nil, err
		}
		return SumExpr{Indices: indices, Body: body}, nil
	}
	if t.Kind == lexer.Identifier {
		name := p.next().Literal
		var indices []string
		if p.peek().Kind == lexer.LBracket {
			p.next()
			for {
				idx, err := p.ident()
				if err != nil {
					return nil, err
				}
				indices = append(indices, idx)
				if p.peek().Kind != lexer.Comma {
					break
				}
				p.next()
			}
			if err := p.want(lexer.RBracket); err != nil {
				return nil, err
			}
		}
		return RefExpr{Name: name, Indices: indices}, nil
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

func isBoundRelation(k lexer.Kind) bool {
	return k == lexer.GreaterEqual || k == lexer.Greater || k == lexer.LessEqual || k == lexer.Less
}

func isConstraintRelation(k lexer.Kind) bool {
	return k == lexer.LessEqual || k == lexer.GreaterEqual || k == lexer.Equal || k == lexer.EqualEqual
}
