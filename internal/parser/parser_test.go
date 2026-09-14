package parser

import "testing"

func TestParseIndexedMILPModel(t *testing.T) {
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
	if len(m.Params) != 2 {
		t.Fatalf("params=%d, want 2", len(m.Params))
	}
	if len(m.Vars) != 2 || !m.Vars[0].Integer {
		t.Fatalf("unexpected vars: %#v", m.Vars)
	}
	if len(m.Constraints) != 1 || m.Constraints[0].Index == nil {
		t.Fatalf("unexpected constraints: %#v", m.Constraints)
	}
	if len(m.Objectives) != 1 || m.Objectives[0].Sense != "minimize" {
		t.Fatalf("unexpected objectives: %#v", m.Objectives)
	}
}
