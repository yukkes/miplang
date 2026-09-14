package compiler

import (
	"os"
	"strings"
	"testing"
)

func TestCompileTransportModelToCanonicalIR(t *testing.T) {
	src, err := os.ReadFile("../../testdata/transport.mip")
	if err != nil {
		t.Fatal(err)
	}
	model, err := Compile(string(src))
	if err != nil {
		t.Fatal(err)
	}
	if model.SchemaVersion != "miplang.ir/v1alpha1" {
		t.Fatalf("schemaVersion=%q", model.SchemaVersion)
	}
	if len(model.Sets) != 1 || model.Sets[0].Name != "TRIPS" {
		t.Fatalf("unexpected sets: %#v", model.Sets)
	}
	if len(model.Variables) != 2 || model.Variables[0].Type != "integer" {
		t.Fatalf("unexpected variables: %#v", model.Variables)
	}
	if len(model.Constraints) != 1 || model.Constraints[0].Operator != "<=" {
		t.Fatalf("unexpected constraints: %#v", model.Constraints)
	}
	if len(model.Objectives) != 1 || model.Objectives[0].Sense != "minimize" {
		t.Fatalf("unexpected objectives: %#v", model.Objectives)
	}
	data1, err := MarshalCanonical(model)
	if err != nil {
		t.Fatal(err)
	}
	data2, err := MarshalCanonical(model)
	if err != nil {
		t.Fatal(err)
	}
	if string(data1) != string(data2) {
		t.Fatal("canonical JSON is not deterministic")
	}
	if !strings.HasSuffix(string(data1), "\n") {
		t.Fatal("canonical JSON must end with newline")
	}
}

func TestCompileRejectsUnknownSymbol(t *testing.T) {
	_, err := Compile(`set I; var x[I] >= 0; minimize Cost: sum[i in I] { missing[i] };`)
	if err == nil || !strings.Contains(err.Error(), "unknown symbol missing") {
		t.Fatalf("error=%v, want unknown symbol", err)
	}
}

func TestCompileRejectsNonlinearVariableProduct(t *testing.T) {
	_, err := Compile(`set I; var x[I] >= 0; var y[I] >= 0; minimize Cost: sum[i in I] { x[i] * y[i] };`)
	if err == nil || !strings.Contains(err.Error(), "nonlinear product") {
		t.Fatalf("error=%v, want nonlinear product", err)
	}
}
