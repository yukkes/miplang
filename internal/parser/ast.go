package parser

type Model struct {
	Declarations []Declaration
	Sets         []SetDecl
	Params       []ParamDecl
	Vars         []VarDecl
	Constraints  []ConstraintDecl
	Objectives   []ObjectiveDecl
}

type Declaration interface{ declNode() }

type SetDecl struct{ Name string }

type Bound struct {
	Expr   Expr
	Strict bool
}

type ParamDecl struct {
	Name    string
	Domain  []Iterator
	Integer bool
	Lower   *Bound
	Upper   *Bound
}

type VarDecl struct {
	Name    string
	Domain  []Iterator
	Integer bool
	Binary  bool
	Lower   *Bound
	Upper   *Bound
}

type Iterator struct {
	Name string
	Set  string
}

type ConstraintDecl struct {
	Name    string
	Indices []Iterator
	Parts   []Expr
	Ops     []string
}

type ObjectiveDecl struct {
	Name    string
	Sense   string
	Indices []Iterator
	Expr    Expr
}

func (SetDecl) declNode()        {}
func (ParamDecl) declNode()      {}
func (VarDecl) declNode()        {}
func (ConstraintDecl) declNode() {}
func (ObjectiveDecl) declNode()  {}

type Expr interface{ exprNode() }
type NumberExpr struct{ Value float64 }
type RefExpr struct {
	Name    string
	Indices []string
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
	Indices []Iterator
	Body    Expr
}

func (NumberExpr) exprNode() {}
func (RefExpr) exprNode()    {}
func (UnaryExpr) exprNode()  {}
func (BinaryExpr) exprNode() {}
func (SumExpr) exprNode()    {}
