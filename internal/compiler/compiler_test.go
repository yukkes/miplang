package compiler

import (
	"os"
	"strings"
	"testing"
)

func TestCompileLegacyTransportModelToCanonicalIR(t *testing.T) {
	src, err := os.ReadFile("../../testdata/transport.mip")
	if err != nil { t.Fatal(err) }
	model, err := Compile(string(src)); if err != nil { t.Fatal(err) }
	if model.SchemaVersion != "miplang.ir/v1alpha2" { t.Fatalf("schemaVersion=%q", model.SchemaVersion) }
	if len(model.Sets) != 1 || model.Sets[0].Name != "TRIPS" { t.Fatalf("unexpected sets: %#v", model.Sets) }
	if len(model.Variables) != 2 || model.Variables[0].Type != "integer" { t.Fatalf("unexpected variables: %#v", model.Variables) }
	if len(model.Constraints) != 1 || model.Constraints[0].Operator != "<=" { t.Fatalf("unexpected constraints: %#v", model.Constraints) }
	if len(model.Objectives) != 1 || model.Objectives[0].Sense != "minimize" { t.Fatalf("unexpected objectives: %#v", model.Objectives) }
	data1, err := MarshalCanonical(model); if err != nil { t.Fatal(err) }
	data2, err := MarshalCanonical(model); if err != nil { t.Fatal(err) }
	if string(data1) != string(data2) { t.Fatal("canonical JSON is not deterministic") }
	if !strings.HasSuffix(string(data1), "\n") { t.Fatal("canonical JSON must end with newline") }
}

func TestCompileAMPLCoreMultidimensionalModel(t *testing.T) {
	src := `
set ORIG;
set DEST;
param capacity {ORIG, DEST} >= 0;
param limit {i in ORIG} > 0;
var x {i in ORIG, j in DEST} >= 0, <= capacity[i,j];
subject to Balance {i in ORIG}: sum {j in DEST} x[i,j] = limit[i];
minimize Cost: sum {i in ORIG, j in DEST} capacity[i,j] * x[i,j];
`
	model, err := Compile(src); if err != nil { t.Fatal(err) }
	if len(model.Parameters[0].Domain) != 2 { t.Fatalf("capacity domain=%#v", model.Parameters[0].Domain) }
	if model.Parameters[1].Lower == nil || !model.Parameters[1].Lower.Strict { t.Fatalf("strict parameter bound lost: %#v", model.Parameters[1]) }
	if len(model.Variables[0].Domain) != 2 || model.Variables[0].Upper == nil { t.Fatalf("variable bounds/domain lost: %#v", model.Variables[0]) }
	if got := model.Variables[0].Upper.Expr; got.Kind != "reference" || len(got.Indices) != 2 { t.Fatalf("unexpected upper expression: %#v", got) }
	if len(model.Constraints) != 1 || model.Constraints[0].Operator != "=" { t.Fatalf("equality not normalized: %#v", model.Constraints) }
	if len(model.Constraints[0].Index) != 1 || len(model.Objectives[0].Expr.Iterators) != 2 { t.Fatalf("multidimensional indexing lost: constraint=%#v objective=%#v", model.Constraints[0], model.Objectives[0]) }
}

func TestCompileRangedConstraintIntoTwoRows(t *testing.T) {
	src := `
set I;
param lo {I};
param hi {I};
var x {I};
subject to Bounds {i in I}: lo[i] <= x[i] <= hi[i];
`
	model, err := Compile(src); if err != nil { t.Fatal(err) }
	if len(model.Constraints) != 2 { t.Fatalf("constraints=%d, want 2", len(model.Constraints)) }
	if model.Constraints[0].Name != "Bounds$lower" || model.Constraints[1].Name != "Bounds$upper" { t.Fatalf("unexpected ranged row names: %#v", model.Constraints) }
}

func TestCompileRejectsUnknownSymbol(t *testing.T) {
	_, err := Compile(`set I; var x[I] >= 0; minimize Cost: sum[i in I] { missing[i] };`)
	if err == nil || !strings.Contains(err.Error(), "unknown symbol missing") { t.Fatalf("error=%v, want unknown symbol", err) }
}
func TestCompileRejectsWrongReferenceArity(t *testing.T) {
	_, err := Compile(`set I; set J; param a {I,J}; var x {i in I}; minimize Cost: sum {i in I} a[i] * x[i];`)
	if err == nil || !strings.Contains(err.Error(), "expects 2 indices, got 1") { t.Fatalf("error=%v, want arity error", err) }
}
func TestCompileRejectsIndexFromWrongSet(t *testing.T) {
	_, err := Compile(`set I; set J; param a {I}; var x {j in J}; minimize Cost: sum {j in J} a[j] * x[j];`)
	if err == nil || !strings.Contains(err.Error(), "ranges over J, expected I") { t.Fatalf("error=%v, want domain mismatch", err) }
}
func TestCompileRejectsVariableDependentBound(t *testing.T) {
	_, err := Compile(`set I; var x {I}; var y {i in I} <= x[i]; minimize Cost: 0;`)
	if err == nil || !strings.Contains(err.Error(), "bound cannot depend on a decision variable") { t.Fatalf("error=%v, want variable-bound error", err) }
}
func TestCompileRejectsNonlinearVariableProduct(t *testing.T) {
	_, err := Compile(`set I; var x[I] >= 0; var y[I] >= 0; minimize Cost: sum[i in I] { x[i] * y[i] };`)
	if err == nil || !strings.Contains(err.Error(), "nonlinear product") { t.Fatalf("error=%v, want nonlinear product", err) }
}
func TestCompileRejectsDivisionByVariable(t *testing.T) {
	_, err := Compile(`set I; var x {I}; var y {I}; minimize Cost: sum {i in I} x[i] / y[i];`)
	if err == nil || !strings.Contains(err.Error(), "division by a decision-variable expression") { t.Fatalf("error=%v, want nonlinear division", err) }
}
func TestAMPL_DECL_001_RejectsForwardReference(t *testing.T) {
	_, err := Compile(`var x <= later; param later; minimize C: x;`)
	if err == nil || !strings.Contains(err.Error(), "unknown symbol later") { t.Fatalf("error=%v, want forward-reference rejection", err) }
}
func TestAMPL_DECL_002_RejectsSetUsedBeforeDeclaration(t *testing.T) {
	_, err := Compile(`param p {I}; set I; minimize C: 0;`)
	if err == nil || !strings.Contains(err.Error(), "unknown set I") { t.Fatalf("error=%v, want set forward-reference rejection", err) }
}
func TestAMPL_REL_001_002_AndLegacyAliasesCanonicalizeIdentically(t *testing.T) {
	ampl := `set I; param rhs {I}; var x {I} >= 0; subject to Balance {i in I}: x[i] = rhs[i]; minimize C: sum {i in I} x[i];`
	legacy := `set I; param rhs[I]; var x[I] >= 0; constraint Balance[i in I]: x[i] == rhs[i]; minimize C: sum[i in I] { x[i] };`
	am, err := Compile(ampl); if err != nil { t.Fatal(err) }
	lm, err := Compile(legacy); if err != nil { t.Fatal(err) }
	a, _ := MarshalCanonical(am); b, _ := MarshalCanonical(lm)
	if string(a) != string(b) { t.Fatalf("canonical forms differ:\nAMPL:\n%s\nlegacy:\n%s", a, b) }
}
func TestAMPL_VAR_004_RejectsImpossibleConstantBinaryBounds(t *testing.T) {
	for _, src := range []string{`var x binary >= 2; minimize C: x;`, `var x binary <= -1; minimize C: x;`} {
		_, err := Compile(src); if err == nil || !strings.Contains(err.Error(), "binary variable") { t.Fatalf("source=%q error=%v, want binary-bound rejection", src, err) }
	}
}
func TestAMPL_REL_004_RejectsStrictSolverConstraint(t *testing.T) {
	_, err := Compile(`var x; subject to Strict: x < 1; minimize C: x;`)
	if err == nil { t.Fatal("expected strict solver relation rejection") }
}
