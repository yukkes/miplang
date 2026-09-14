package main

import (
	"fmt"
	"io"
	"os"

	"github.com/yukkes/miplang/internal/compiler"
)

func main() { os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr)) }

func run(args []string, _ io.Reader, stdout, stderr io.Writer) int {
	if len(args) < 2 || args[0] != "compile" {
		fmt.Fprintln(stderr, "usage: miplang compile <model.mip> [-o <path>]")
		return 2
	}
	input := args[1]
	output := ""
	if len(args) > 2 {
		if len(args) != 4 || args[2] != "-o" {
			fmt.Fprintln(stderr, "usage: miplang compile <model.mip> [-o <path>]")
			return 2
		}
		output = args[3]
	}
	src, err := os.ReadFile(input)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	model, err := compiler.Compile(string(src))
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	data, err := compiler.MarshalCanonical(model)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if output == "" {
		if _, err := stdout.Write(data); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return 0
	}
	if err := os.WriteFile(output, data, 0o644); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}
