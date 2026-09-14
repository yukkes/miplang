package compiler

import (
	"encoding/json"
	"fmt"

	"github.com/yukkes/miplang/internal/ir"
	"github.com/yukkes/miplang/internal/parser"
)

type symbol struct {
	kind   string
	domain []string
}

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

	for _, declaration := range ast.Declarations {
		switch d := declaration.(type) {
		case parser.SetDecl:
			if err := ctx.declare(d.Name, symbol{kind: "set"}); err != nil {
				return nil, err
			}
			ctx.sets[d.Name] = struct{}{}
			out.Sets = append(out.Sets, ir.Set{Name: d.Name})

		case parser.ParamDecl:
			domain, local, err := ctx.prepareDomain(d.Domain)
			if err != nil {
				return nil, fmt.Errorf("param %s: %w", d.Name, err)
			}
			lower, err := lowerBound(d.Lower, local, false)
			if err != nil {
				return nil, fmt.Errorf("param %s lower bound: %w", d.Name, err)
			}
			upper, err := lowerBound(d.Upper, local, false)
			if err != nil {
				return nil, fmt.Errorf("param %s upper bound: %w", d.Name, err)
			}
			if err := ctx.declare(d.Name, symbol{kind: "param", domain: domainSets(domain)}); err != nil {
				return nil, err
			}
			out.Parameters = append(out.Parameters, ir.Parameter{
				Name: d.Name, Domain: domain, Integer: d.Integer, Lower: lower, Upper: upper,
			})

		case parser.VarDecl:
			domain, local, err := ctx.prepareDomain(d.Domain)
			if err != nil {
				return nil, fmt.Errorf("var %s: %w", d.Name, err)
			}
			lower, err := lowerBound(d.Lower, local, true)
			if err != nil {
				return nil, fmt.Errorf("var %s lower bound: %w", d.Name, err)
			}
			upper, err := lowerBound(d.Upper, local, true)
			if err != nil {
				return nil, fmt.Errorf("var %s upper bound: %w", d.Name, err)
			}
			if d.Binary {
				if err := validateBinaryBounds(lower, upper); err != nil {
					return nil, fmt.Errorf("var %s: %w", d.Name, err)
				}
			}
			if err := ctx.declare(d.Name, symbol{kind: "var", domain: domainSets(domain)}); err != nil {
				return nil, err
			}
			typ := "continuous"
			if d.Integer {
				typ = "integer"
			}
			if d.Binary {
				typ = "binary"
			}
			out.Variables = append(out.Variables, ir.Variable{
				Name: d.Name, Domain: domain, Type: typ, Lower: lower, Upper: upper,
			})

		case parser.ConstraintDecl:
			index, local, err := ctx.prepareDomain(d.Indices)
			if err != nil {
				return nil, fmt.Errorf("constraint %s: %w", d.Name, err)
			}
			parts := make([]ir.Expr, len(d.Parts))
			degrees := make([]int, len(d.Parts))
			for i, part := range d.Parts {
				parts[i], degrees[i], err = lowerExpr(part, local)
				if err != nil {
					return nil, fmt.Errorf("constraint %s: %w", d.Name, err)
				}
			}
			switch len(d.Ops) {
			case 1:
				out.Constraints = append(out.Constraints, ir.Constraint{
					Name: d.Name, Index: index, Left: parts[0], Operator: normalizeRelation(d.Ops[0]), Right: parts[1],
				})
			case 2:
				if degrees[0] > 0 || degrees[2] > 0 {
					return nil, fmt.Errorf("constraint %s: ranged constraint outer expressions cannot contain decision variables", d.Name)
				}
				out.Constraints = append(out.Constraints,
					ir.Constraint{Name: d.Name + "$lower", Index: index, Left: parts[0], Operator: normalizeRelation(d.Ops[0]), Right: parts[1]},
					ir.Constraint{Name: d.Name + "$upper", Index: index, Left: parts[1], Operator: normalizeRelation(d.Ops[1]), Right: parts[2]},
				)
			default:
				return nil, fmt.Errorf("constraint %s: unsupported relation chain", d.Name)
			}

		case parser.ObjectiveDecl:
			index, local, err := ctx.prepareDomain(d.Indices)
			if err != nil {
				return nil, fmt.Errorf("objective %s: %w", d.Name, err)
			}
			e, _, err := lowerExpr(d.Expr, local)
			if err != nil {
				return nil, fmt.Errorf("objective %s: %w", d.Name, err)
			}
			out.Objectives = append(out.Objectives, ir.Objective{Name: d.Name, Sense: d.Sense, Index: index, Expr: e})

		default:
			return nil, fmt.Errorf("unsupported declaration %T", declaration)
		}
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

func (c *context) prepareDomain(items []parser.Iterator) ([]ir.Iterator, *context, error) {
	out := make([]ir.Iterator, 0, len(items))
	for _, item := range items {
		if _, ok := c.sets[item.Set]; !ok {
			return nil, nil, fmt.Errorf("unknown set %s", item.Set)
		}
		out = append(out, ir.Iterator{Name: item.Name, Set: item.Set})
	}
	local, err := c.withIterators(items)
	if err != nil {
		return nil, nil, err
	}
	return out, local, nil
}

func (c *context) withIterators(items []parser.Iterator) (*context, error) {
	n := &context{symbols: c.symbols, sets: c.sets, iterators: map[string]string{}}
	for k, v := range c.iterators {
		n.iterators[k] = v
	}
	for _, it := range items {
		if _, ok := c.sets[it.Set]; !ok {
			return nil, fmt.Errorf("unknown set %s", it.Set)
		}
		if it.Name == "" {
			continue
		}
		if _, exists := n.iterators[it.Name]; exists {
			return nil, fmt.Errorf("duplicate index %s", it.Name)
		}
		n.iterators[it.Name] = it.Set
	}
	return n, nil
}

func lowerBound(b *parser.Bound, c *context, variableBound bool) (*ir.Bound, error) {
	if b == nil {
		return nil, nil
	}
	if variableBound && b.Strict {
		return nil, fmt.Errorf("strict variable bounds are not supported")
	}
	e, degree, err := lowerExpr(b.Expr, c)
	if err != nil {
		return nil, err
	}
	if degree > 0 {
		return nil, fmt.Errorf("bound cannot depend on a decision variable")
	}
	return &ir.Bound{Expr: e, Strict: b.Strict}, nil
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
		if len(v.Indices) != len(s.domain) {
			return ir.Expr{}, 0, fmt.Errorf("symbol %s expects %d indices, got %d", v.Name, len(s.domain), len(v.Indices))
		}
		for i, idx := range v.Indices {
			set, exists := c.iterators[idx]
			if !exists {
				return ir.Expr{}, 0, fmt.Errorf("unknown index %s", idx)
			}
			if set != s.domain[i] {
				return ir.Expr{}, 0, fmt.Errorf("index %s ranges over %s, expected %s for %s dimension %d", idx, set, s.domain[i], v.Name, i+1)
			}
		}
		degree := 0
		if s.kind == "var" {
			degree = 1
		}
		return ir.Expr{Kind: "reference", Name: v.Name, Indices: append([]string(nil), v.Indices...)}, degree, nil
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
		switch v.Op {
		case "*":
			if dl > 0 && dr > 0 {
				return ir.Expr{}, 0, fmt.Errorf("nonlinear product is not allowed in LP/MILP")
			}
			degree = dl + dr
		case "/":
			if dr > 0 {
				return ir.Expr{}, 0, fmt.Errorf("division by a decision-variable expression is not allowed in LP/MILP")
			}
			degree = dl
		}
		return ir.Expr{Kind: "binary", Operator: v.Op, Left: &left, Right: &right}, degree, nil
	case parser.SumExpr:
		local, err := c.withIterators(v.Indices)
		if err != nil {
			return ir.Expr{}, 0, err
		}
		body, d, err := lowerExpr(v.Body, local)
		if err != nil {
			return ir.Expr{}, 0, err
		}
		iterators := make([]ir.Iterator, 0, len(v.Indices))
		for _, it := range v.Indices {
			iterators = append(iterators, ir.Iterator{Name: it.Name, Set: it.Set})
		}
		return ir.Expr{Kind: "sum", Iterators: iterators, Body: &body}, d, nil
	default:
		return ir.Expr{}, 0, fmt.Errorf("unsupported expression %T", e)
	}
}

func normalizeRelation(op string) string {
	if op == "==" {
		return "="
	}
	return op
}

func domainSets(domain []ir.Iterator) []string {
	sets := make([]string, len(domain))
	for i, d := range domain {
		sets[i] = d.Set
	}
	return sets
}

func validateBinaryBounds(lower, upper *ir.Bound) error {
	if lower != nil {
		if v, ok := constantValue(lower.Expr); ok && v > 1 {
			return fmt.Errorf("binary variable lower bound %g is greater than 1", v)
		}
	}
	if upper != nil {
		if v, ok := constantValue(upper.Expr); ok && v < 0 {
			return fmt.Errorf("binary variable upper bound %g is less than 0", v)
		}
	}
	return nil
}

func constantValue(e ir.Expr) (float64, bool) {
	switch e.Kind {
	case "number":
		if e.Value == nil {
			return 0, false
		}
		return *e.Value, true
	case "unary":
		if e.Operand == nil || e.Operator != "-" {
			return 0, false
		}
		v, ok := constantValue(*e.Operand)
		return -v, ok
	case "binary":
		if e.Left == nil || e.Right == nil {
			return 0, false
		}
		left, lok := constantValue(*e.Left)
		right, rok := constantValue(*e.Right)
		if !lok || !rok {
			return 0, false
		}
		switch e.Operator {
		case "+":
			return left + right, true
		case "-":
			return left - right, true
		case "*":
			return left * right, true
		case "/":
			if right == 0 {
				return 0, false
			}
			return left / right, true
		}
	}
	return 0, false
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
