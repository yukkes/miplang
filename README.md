# MIPLang

**Write one optimization model. Run it from any language.**

MIPLang is a lightweight algebraic modeling language for LP/MILP with TypeScript, Python, Java, Go, and Rust runtimes, designed for HiGHS.

> **Status: early alpha / v0.2 draft.** The compiler now implements an AMPL-compatible LP/MILP core syntax and emits Canonical IR v1alpha2. Native HiGHS solver adapters and runtime data binding are the next milestones.

## Why MIPLang?

Application teams often rewrite the same mathematical model in every implementation language. MIPLang keeps the mathematical model in one source file and moves language-specific integration behind a stable Canonical IR.

```text
model.mod / model.mip
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

The compiler owns parsing, declaration order, name/domain resolution, indexed expressions, and LP/MILP linearity checks. Runtimes do not implement another modeling DSL.

## AMPL-style example

```ampl
set ORIG;
set DEST;

param supply {ORIG} >= 0;
param demand {DEST} >= 0;
param cost {ORIG, DEST} >= 0;

var Trans {ORIG, DEST} >= 0;

minimize Total_Cost:
    sum {i in ORIG, j in DEST} cost[i,j] * Trans[i,j];

subject to Supply {i in ORIG}:
    sum {j in DEST} Trans[i,j] = supply[i];

subject to Demand {j in DEST}:
    sum {i in ORIG} Trans[i,j] = demand[j];
```

Compile it to versioned Canonical IR:

```bash
go run ./cmd/miplang compile examples/transport.mod -o transport.ir.json
```

## AMPL LP/MILP Core compatibility

v0.2 accepts common AMPL modeling forms used by LP/MILP models, including:

- `#` and `/* ... */` comments
- `set`, `param`, `var`
- `{I}`, `{i in I}`, and multi-dimensional `{i in I, j in J}` indexing
- `x[i]`, `x[i,j]`, and higher-dimensional references
- parameter validation bounds
- continuous, integer, and binary variables
- parameter-expression variable bounds such as `>= lo[i], <= hi[i]`
- `minimize` / `maximize`
- `subject to`
- `sum {i in I} ...`
- `+`, `-`, `*`, `/`, parentheses
- `=`, `==`, `<=`, `>=`
- ranged constraints such as `lo[i] <= x[i] <= hi[i]`
- `.mod` source files

MIPLang keeps the v0.1 forms such as `constraint`, `param p[I]`, and `sum[i in I] { ... }` as legacy aliases. Equivalent forms normalize to the same Canonical IR.

This is deliberately **not full AMPL compatibility**. `.dat`, general set expressions, ranges, filtered indexing, nonlinear functions, complementarity, AMPL scripting, suffixes, `.nl`, and AMPL presolve behavior are not part of the v0.2 profile.

See:

- [`docs/spec/language-spec.md`](docs/spec/language-spec.md) — normative MIPLang language specification
- [`docs/spec/ampl-compatibility.md`](docs/spec/ampl-compatibility.md) — AMPL feature-by-feature compatibility matrix
- [`docs/spec/ampl-conformance.md`](docs/spec/ampl-conformance.md) — compatibility IDs and executable test requirements

## Canonical IR

The current deterministic symbolic IR is:

```json
{
  "schemaVersion": "miplang.ir/v1alpha2"
}
```

v1alpha2 represents indexing structurally, including multi-dimensional domains and references, and preserves symbolic parameter/variable bounds. Its JSON Schema is [`schema/miplang-ir.schema.json`](schema/miplang-ir.schema.json).

The next lowering stage will bind runtime data and produce Numeric Sparse IR for HiGHS.

## Safety guarantees for LP/MILP

The compiler rejects, before solver invocation:

- use-before-declaration and unknown symbols
- unknown set domains
- reference arity/domain mismatches
- variable-dependent parameter or variable bounds
- variable-by-variable products
- division by decision-variable expressions
- strict `<` / `>` solver constraints
- impossible constant binary bounds

## Repository layout

```text
cmd/miplang/          Go CLI compiler
internal/lexer/       Lexer with source positions
internal/parser/      Recursive-descent parser and AST
internal/compiler/    Semantic validation and IR lowering
internal/ir/          Canonical IR types
docs/spec/            Language, AMPL compatibility, conformance specs
schema/               Versioned IR schema
runtimes/             TypeScript, Python, Java, Go, Rust loaders
examples/             Example MIPLang / AMPL-style models
testdata/             Golden and cross-language conformance fixtures
```

## Development

Compiler and Go tests:

```bash
go test ./...
```

Golden IR checks:

```bash
go run ./cmd/miplang compile testdata/transport.mip -o /tmp/transport.ir.json
diff -u testdata/transport.ir.json /tmp/transport.ir.json

go run ./cmd/miplang compile testdata/transport-ampl.mod -o /tmp/transport-ampl.ir.json
diff -u testdata/transport-ampl.ir.json /tmp/transport-ampl.ir.json
```

GitHub Actions verifies the compiler plus TypeScript, Python, Java, Go, and Rust runtime contracts.

## Roadmap

1. Complete AMPL LP/MILP Core v0.2 conformance and diagnostics.
2. Add runtime data binding and Numeric Sparse IR.
3. Add HiGHS adapters for TypeScript, Python, Java, Go, and Rust.
4. Add selected AMPL data/indexing features such as `.dat`, ranges, and filtered indexing.
5. Add portable lexicographic multi-objective priorities/tolerances.
6. Benchmark model instantiation and HiGHS solve parity across all runtimes.

MIPLang is an independent implementation. AMPL compatibility refers only to the explicitly documented source-level subset and does not imply AMPL implementation, binary, `.nl`, or translator compatibility.

## License

MIT License. See [`LICENSE`](LICENSE).
