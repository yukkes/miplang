package parser

type Model struct {
	Sets        []SetDecl
	Params      []ParamDecl
	Vars        []VarDecl
	Constraints []ConstraintDecl
	Objectives  []ObjectiveDecl
}

type SetDecl struct{ Name string }
type ParamDecl struct {
	Name   string
	Domain string
}
type VarDecl struct {
	Name    string
	Domain  string
	Integer bool
	Binary  bool
	Lower   *float64
	Upper   *float64
}
type Iterator struct {
	Name string
	Set  string
}
type ConstraintDecl struct {
	Name  string
	Index *Iterator
	Left  Expr
	Op    string
	Right Expr
}
type ObjectiveDecl struct {
	Name  string
	Sense string
	Expr  Expr
}

type Expr interface{ exprNode() }
type NumberExpr struct{ Value float64 }
type RefExpr struct {
	Name  string
	Index string
}
type UnaryExpr struct {
	Op    string
	Value Expr
}
type BinaryExpr struct {
	Op          string
	Left, Right Expr
}
type SumExpr struct {
	Index Iterator
	Body  Expr
}

func (NumberExpr) exprNode() {}
func (RefExpr) exprNode()    {}
func (UnaryExpr) exprNode()  {}
func (BinaryExpr) exprNode() {}
func (SumExpr) exprNode()    {}
