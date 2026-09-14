package parser

import "testing"

func TestParseLegacyIndexedMILPModel(t *testing.T) {
	src := `
set TRIPS;
param capacity[TRIPS];
param cost[TRIPS];
var qFixed[TRIPS] integer >= 0;
var qSpot[TRIPS] integer >= 0;
constraint Capacity[t in TRIPS]: qFixed[t] + qSpot[t] <= capacity[t];
minimize Cost: sum[t in TRIPS] { cost[t] * (qFixed[t] + qSpot[t]) };
`
	m, err := Parse(src)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Sets) != 1 || m.Sets[0].Name != "TRIPS" {
		t.Fatalf("unexpected sets: %#v", m.Sets)
	}
	if len(m.Params) != 2 || len(m.Params[0].Domain) != 1 || m.Params[0].Domain[0].Set != "TRIPS" {
		t.Fatalf("unexpected params: %#v", m.Params)
	}
	if len(m.Vars) != 2 || !m.Vars[0].Integer {
		t.Fatalf("unexpected vars: %#v", m.Vars)
	}
	if len(m.Constraints) != 1 || len(m.Constraints[0].Indices) != 1 {
		t.Fatalf("unexpected constraints: %#v", m.Constraints)
	}
	if len(m.Objectives) != 1 || m.Objectives[0].Sense != "minimize" {
		t.Fatalf("unexpected objectives: %#v", m.Objectives)
	}
}

func TestParseAMPLCoreModel(t *testing.T) {
	src := `
# AMPL-style core syntax
set ORIG;
set DEST;
param capacity {ORIG, DEST} >= 0;
param limit {i in ORIG} > 0;
var x {i in ORIG, j in DEST} >= 0, <= capacity[i,j];
subject to Balance {i in ORIG}:
    sum {j in DEST} x[i,j] = limit[i];
minimize Cost:
    sum {i in ORIG, j in DEST} capacity[i,j] * x[i,j];
`
	m, err := Parse(src)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Params) != 2 || len(m.Params[0].Domain) != 2 {
		t.Fatalf("unexpected parameter domain: %#v", m.Params)
	}
	if m.Params[1].Lower == nil || !m.Params[1].Lower.Strict {
		t.Fatalf("strict parameter lower bound not parsed: %#v", m.Params[1])
	}
	if len(m.Vars) != 1 || len(m.Vars[0].Domain) != 2 || m.Vars[0].Upper == nil {
		t.Fatalf("unexpected variable: %#v", m.Vars)
	}
	if len(m.Constraints) != 1 || len(m.Constraints[0].Indices) != 1 {
		t.Fatalf("unexpected constraint indexing: %#v", m.Constraints)
	}
	if len(m.Constraints[0].Ops) != 1 || m.Constraints[0].Ops[0] != "=" {
		t.Fatalf("unexpected relation: %#v", m.Constraints[0].Ops)
	}
	sum, ok := m.Objectives[0].Expr.(SumExpr)
	if !ok || len(sum.Indices) != 2 {
		t.Fatalf("unexpected objective sum: %#v", m.Objectives[0].Expr)
	}
}

func TestParseAMPLRangedConstraintAndDivision(t *testing.T) {
	src := `
set I;
param lo {I};
param hi {I};
param scale {I} > 0;
var x {I} >= 0;
subject to Bounds {i in I}: lo[i] <= x[i] / scale[i] <= hi[i];
`
	m, err := Parse(src)
	if err != nil {
		t.Fatal(err)
	}
	c := m.Constraints[0]
	if len(c.Parts) != 3 || len(c.Ops) != 2 || c.Ops[0] != "<=" || c.Ops[1] != "<=" {
		t.Fatalf("unexpected ranged constraint: %#v", c)
	}
	middle, ok := c.Parts[1].(BinaryExpr)
	if !ok || middle.Op != "/" {
		t.Fatalf("division not parsed: %#v", c.Parts[1])
	}
}
