# MIPLang

**Write one optimization model. Run it from any language.**

MIPLang is a lightweight algebraic modeling language for LP/MILP with TypeScript, Python, Java, Go, and Rust runtimes, powered by HiGHS.

> **Status: early alpha.** The current v0.1 vertical slice implements the language parser, semantic LP/MILP validation, canonical IR, CLI compiler, and cross-language IR loaders. Native HiGHS solver adapters are the next milestone.

## Why MIPLang?

Application teams often end up rewriting the same mathematical model for every implementation language. MIPLang keeps the mathematical model in one `.mip` file and moves language-specific code behind a stable canonical IR.

```text
model.mip
   │
   ▼
MIPLang compiler (Go)
   │
   ▼
Canonical IR
   │
   ├── TypeScript runtime
   ├── Python runtime
   ├── Java runtime
   ├── Go runtime
   └── Rust runtime
            │
            ▼
          HiGHS
```

The compiler owns parsing, name resolution, indexed expressions, and LP/MILP linearity checks. Runtimes do not implement another modeling DSL.

## Example

```mip
set TRIPS;

param capacity[TRIPS];
param cost[TRIPS];

var qFixed[TRIPS] integer >= 0;
var qSpot[TRIPS] integer >= 0;

constraint Capacity[t in TRIPS]:
    qFixed[t] + qSpot[t] <= capacity[t];

minimize Cost:
    sum[t in TRIPS] {
        cost[t] * (qFixed[t] + qSpot[t])
    };
```

Compile it to the versioned canonical IR:

```bash
go run ./cmd/miplang compile examples/transport.mip -o transport.ir.json
```

The compiler rejects unknown symbols and nonlinear variable-by-variable products before a solver sees the model.

## v0.1 language surface

- `set`
- scalar and one-dimensional `param`
- continuous, integer, and binary `var`
- lower/upper variable bounds
- scalar and indexed `constraint`
- `minimize` / `maximize`
- `sum[i in SET] { ... }`
- `+`, `-`, `*`, parentheses
- compile-time symbol/domain validation
- compile-time rejection of nonlinear variable products

## Canonical IR

The canonical IR is deterministic JSON identified by:

```json
{
  "schemaVersion": "miplang.ir/v1alpha1"
}
```

Its schema is in [`schema/miplang-ir.schema.json`](schema/miplang-ir.schema.json). All runtimes consume this same representation.

## Repository layout

```text
cmd/miplang/          Go CLI compiler
internal/lexer/       Lexer with source positions
internal/parser/      Recursive-descent parser and AST
internal/compiler/    Semantic validation and IR lowering
internal/ir/          Canonical IR types
schema/               Versioned IR schema
runtimes/             TypeScript, Python, Java, Go, Rust loaders
examples/             Example MIPLang models
testdata/             Cross-language conformance fixtures
```

## Development

Compiler tests:

```bash
go test ./...
```

Runtime conformance tests are intentionally independent and all read `testdata/transport.ir.json`. GitHub Actions verifies Go, Python, TypeScript, Java, and Rust.

## Roadmap

1. Stabilize symbolic IR and language diagnostics.
2. Add external/runtime data binding and numeric sparse IR.
3. Add HiGHS adapters for TypeScript, Python, Java, Go, and Rust.
4. Add lexicographic multi-objective priorities and tolerances.
5. Benchmark model instantiation and HiGHS solve parity across runtimes.

MIPLang is inspired by algebraic modeling languages such as AMPL, JuMP, Pyomo, and MiniZinc, but it is an independent language and is not an AMPL-compatible implementation.

## License

Apache License 2.0. See [`LICENSE`](LICENSE).
