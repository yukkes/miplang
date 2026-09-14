package miplang

import "testing"

func TestLoadSharedIR(t *testing.T) {
	m, err := Load("../../testdata/transport.ir.json")
	if err != nil {
		t.Fatal(err)
	}
	if m.SchemaVersion != "miplang.ir/v1alpha1" || m.Name != "anonymous" {
		t.Fatalf("unexpected model: %#v", m)
	}
	if len(m.Sets) != 1 || m.Sets[0].Name != "TRIPS" {
		t.Fatalf("unexpected sets: %#v", m.Sets)
	}
}
