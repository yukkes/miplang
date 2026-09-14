package compiler

import (
	"encoding/json"
	"fmt"

	"github.com/yukkes/miplang/internal/ir"
	"github.com/yukkes/miplang/internal/parser"
)

type symbol struct{ kind, domain string }
type context struct {
	symbols   map[string]symbol
	sets      map[string]struct{}
	iterators map[string]string
}

func Compile(source string) (*ir.Model, error) {
	ast, err := parser.Parse(source)
	if err != nil {
		return nil, err
	}
	ctx := &context{symbols: map[string]symbol{}, sets: map[string]struct{}{}, iterators: map[string]string{}}
	out := &ir.Model{SchemaVersion: ir.SchemaVersion, Name: "anonymous"}
	for _, s := range ast.Sets {
		if err := ctx.declare(s.Name, symbol{kind: "set"}); err != nil {
			return nil, err
		}
		ctx.sets[s.Name] = struct{}{}
		out.Sets = append(out.Sets, ir.Set{Name: s.Name})
	}
	for _, p := range ast.Params {
		if err := ctx.requireDomain(p.Domain); err != nil {
			return nil, fmt.Errorf("param %s: %w", p.Name, err)
		}
		if err := ctx.declare(p.Name, symbol{kind: "param", domain: p.Domain}); err != nil {
			return nil, err
		}
		out.Parameters = append(out.Parameters, ir.Parameter{Name: p.Name, Domain: p.Domain})
	}
	for _, v := range ast.Vars {
		if err := ctx.requireDomain(v.Domain); err != nil {
			return nil, fmt.Errorf("var %s: %w", v.Name, err)
		}
		if err := ctx.declare(v.Name, symbol{kind: "var", domain: v.Domain}); err != nil {
			return nil, err
		}
		typ := "continuous"
		if v.Integer {
			typ = "integer"
		}
		if v.Binary {
			typ = "binary"
		}
		out.Variables = append(out.Variables, ir.Variable{Name: v.Name, Domain: v.Domain, Type: typ, Lower: v.Lower, Upper: v.Upper})
	}
	for _, c := range ast.Constraints {
		local := ctx.withIterator(c.Index)
		if local == nil {
			return nil, fmt.Errorf("constraint %s: unknown set %s", c.Name, c.Index.Set)
		}
		left, dl, err := lowerExpr(c.Left, local)
		if err != nil {
			return nil, fmt.Errorf("constraint %s: %w", c.Name, err)
		}
		right, dr, err := lowerExpr(c.Right, local)
		if err != nil {
			return nil, fmt.Errorf("constraint %s: %w", c.Name, err)
		}
		_ = dl
		_ = dr
		var idx *ir.Iterator
		if c.Index != nil {
			idx = &ir.Iterator{Name: c.Index.Name, Set: c.Index.Set}
		}
		out.Constraints = append(out.Constraints, ir.Constraint{Name: c.Name, Index: idx, Left: left, Operator: c.Op, Right: right})
	}
	for _, o := range ast.Objectives {
		e, _, err := lowerExpr(o.Expr, ctx)
		if err != nil {
			return nil, fmt.Errorf("objective %s: %w", o.Name, err)
		}
		out.Objectives = append(out.Objectives, ir.Objective{Name: o.Name, Sense: o.Sense, Expr: e})
	}
	return out, nil
}

func MarshalCanonical(model *ir.Model) ([]byte, error) {
	b, err := json.MarshalIndent(model, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(b, '\n'), nil
}

func (c *context) declare(name string, s symbol) error {
	if _, ok := c.symbols[name]; ok {
		return fmt.Errorf("duplicate symbol %s", name)
	}
	c.symbols[name] = s
	return nil
}
func (c *context) requireDomain(domain string) error {
	if domain == "" {
		return nil
	}
	if _, ok := c.sets[domain]; !ok {
		return fmt.Errorf("unknown set %s", domain)
	}
	return nil
}
func (c *context) withIterator(it *parser.Iterator) *context {
	if it == nil {
		return c
	}
	if _, ok := c.sets[it.Set]; !ok {
		return nil
	}
	n := &context{symbols: c.symbols, sets: c.sets, iterators: map[string]string{}}
	for k, v := range c.iterators {
		n.iterators[k] = v
	}
	n.iterators[it.Name] = it.Set
	return n
}

func lowerExpr(e parser.Expr, c *context) (ir.Expr, int, error) {
	switch v := e.(type) {
	case parser.NumberExpr:
		n := v.Value
		return ir.Expr{Kind: "number", Value: &n}, 0, nil
	case parser.RefExpr:
		s, ok := c.symbols[v.Name]
		if !ok {
			return ir.Expr{}, 0, fmt.Errorf("unknown symbol %s", v.Name)
		}
		if s.kind == "set" {
			return ir.Expr{}, 0, fmt.Errorf("set %s cannot be used as a numeric expression", v.Name)
		}
		if s.domain != "" && v.Index == "" {
			return ir.Expr{}, 0, fmt.Errorf("indexed symbol %s requires an index", v.Name)
		}
		if s.domain == "" && v.Index != "" {
			return ir.Expr{}, 0, fmt.Errorf("scalar symbol %s cannot be indexed", v.Name)
		}
		if v.Index != "" {
			set, exists := c.iterators[v.Index]
			if !exists {
				return ir.Expr{}, 0, fmt.Errorf("unknown index %s", v.Index)
			}
			if set != s.domain {
				return ir.Expr{}, 0, fmt.Errorf("index %s ranges over %s, expected %s for %s", v.Index, set, s.domain, v.Name)
			}
		}
		degree := 0
		if s.kind == "var" {
			degree = 1
		}
		return ir.Expr{Kind: "reference", Name: v.Name, Index: v.Index}, degree, nil
	case parser.UnaryExpr:
		op, d, err := lowerExpr(v.Value, c)
		if err != nil {
			return ir.Expr{}, 0, err
		}
		return ir.Expr{Kind: "unary", Operator: v.Op, Operand: &op}, d, nil
	case parser.BinaryExpr:
		left, dl, err := lowerExpr(v.Left, c)
		if err != nil {
			return ir.Expr{}, 0, err
		}
		right, dr, err := lowerExpr(v.Right, c)
		if err != nil {
			return ir.Expr{}, 0, err
		}
		degree := max(dl, dr)
		if v.Op == "*" {
			if dl > 0 && dr > 0 {
				return ir.Expr{}, 0, fmt.Errorf("nonlinear product is not allowed in LP/MILP")
			}
			degree = dl + dr
		}
		return ir.Expr{Kind: "binary", Operator: v.Op, Left: &left, Right: &right}, degree, nil
	case parser.SumExpr:
		local := c.withIterator(&v.Index)
		if local == nil {
			return ir.Expr{}, 0, fmt.Errorf("unknown set %s", v.Index.Set)
		}
		body, d, err := lowerExpr(v.Body, local)
		if err != nil {
			return ir.Expr{}, 0, err
		}
		it := ir.Iterator{Name: v.Index.Name, Set: v.Index.Set}
		return ir.Expr{Kind: "sum", Iterator: &it, Body: &body}, d, nil
	default:
		return ir.Expr{}, 0, fmt.Errorf("unsupported expression %T", e)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
