package compiler

import (
	"strings"
	"testing"
)

func TestAMPL_LEX_002_IdentifiersAreCaseSensitive(t *testing.T) {
	model, err := Compile(`var x; var X; minimize C: x + X;`); if err != nil { t.Fatal(err) }
	if len(model.Variables) != 2 || model.Variables[0].Name != "x" || model.Variables[1].Name != "X" { t.Fatalf("variables=%#v", model.Variables) }
}
func TestAMPL_LEX_006_MissingSemicolonIsRejected(t *testing.T) {
	_, err := Compile("set I\nparam p;"); if err == nil { t.Fatal("expected missing-semicolon diagnostic") }
}
func TestAMPL_LEX_007_ScientificLiteralReachesIR(t *testing.T) {
	model, err := Compile(`param p >= 2.5E-2; minimize C: p;`); if err != nil { t.Fatal(err) }
	if model.Parameters[0].Lower == nil || model.Parameters[0].Lower.Expr.Value == nil { t.Fatalf("lower=%#v", model.Parameters[0].Lower) }
	if got := *model.Parameters[0].Lower.Expr.Value; got != 0.025 { t.Fatalf("value=%g", got) }
}
func TestAMPL_PARAM_001_ScalarParameter(t *testing.T) {
	model, err := Compile(`param p; minimize C: p;`); if err != nil { t.Fatal(err) }
	if len(model.Parameters) != 1 || len(model.Parameters[0].Domain) != 0 { t.Fatalf("parameter=%#v", model.Parameters) }
}
func TestAMPL_PARAM_005_ParameterBoundCannotDependOnVariable(t *testing.T) {
	_, err := Compile(`var x; param p >= x; minimize C: x;`)
	if err == nil || !strings.Contains(err.Error(), "bound cannot depend on a decision variable") { t.Fatalf("error=%v", err) }
}
func TestAMPL_VAR_001_003_004_VariableTypes(t *testing.T) {
	model, err := Compile(`var c; var i integer; var b binary; minimize C: c + i + b;`); if err != nil { t.Fatal(err) }
	want := []string{"continuous", "integer", "binary"}; if len(model.Variables) != len(want) { t.Fatalf("variables=%#v", model.Variables) }
	for i, typ := range want { if model.Variables[i].Type != typ { t.Fatalf("variable %d type=%q want=%q", i, model.Variables[i].Type, typ) } }
}
func TestAMPL_OBJ_001_Maximize(t *testing.T) {
	model, err := Compile(`var x; maximize Profit: x;`); if err != nil { t.Fatal(err) }
	if len(model.Objectives) != 1 || model.Objectives[0].Sense != "maximize" { t.Fatalf("objectives=%#v", model.Objectives) }
}
func TestAMPL_EXPR_001_UnaryMinus(t *testing.T) {
	model, err := Compile(`var x; minimize C: -x + 2;`); if err != nil { t.Fatal(err) }
	if model.Objectives[0].Expr.Kind != "binary" || model.Objectives[0].Expr.Left == nil || model.Objectives[0].Expr.Left.Kind != "unary" { t.Fatalf("expr=%#v", model.Objectives[0].Expr) }
}
func TestAMPL_EXPR_003_DivisionByParameterIsLinear(t *testing.T) {
	model, err := Compile(`set I; param rate {I} > 0; var x {I}; minimize C: sum {i in I} (1 / rate[i]) * x[i];`); if err != nil { t.Fatal(err) }
	if model.Objectives[0].Expr.Kind != "sum" { t.Fatalf("expr=%#v", model.Objectives[0].Expr) }
}
func TestAMPL_REL_003_GreaterEqualConstraint(t *testing.T) {
	model, err := Compile(`var x; subject to Minimum: x >= 1; minimize C: x;`); if err != nil { t.Fatal(err) }
	if model.Constraints[0].Operator != ">=" { t.Fatalf("constraint=%#v", model.Constraints[0]) }
}
