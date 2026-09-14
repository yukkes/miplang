package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunCompileWritesCanonicalIR(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "transport.ir.json")
	var stdout, stderr bytes.Buffer
	code := run([]string{"compile", "../../testdata/transport.mip", "-o", outPath}, bytes.NewReader(nil), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"schemaVersion": "miplang.ir/v1alpha1"`) {
		t.Fatalf("unexpected output: %s", data)
	}
}

func TestRunCompileInvalidModelReturnsNonZero(t *testing.T) {
	dir := t.TempDir()
	model := filepath.Join(dir, "bad.mip")
	if err := os.WriteFile(model, []byte("minimize Cost: missing;"), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := run([]string{"compile", model}, bytes.NewReader(nil), &stdout, &stderr)
	if code == 0 {
		t.Fatal("expected non-zero exit code")
	}
	if !strings.Contains(stderr.String(), "unknown symbol missing") {
		t.Fatalf("stderr=%q", stderr.String())
	}
}
