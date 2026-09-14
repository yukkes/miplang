# AMPL Compatibility Conformance Specification

Version: 0.2-draft

## 1. Purpose

Every **Supported** row in `ampl-compatibility.md` must map to executable tests. A feature MUST NOT be marked Supported until its case passes in CI.

Layers are LEX (tokenization), PARSE (AST), SEM (resolution/linearity/normalization), IR (golden Canonical IR), plus CLI/RUNTIME where applicable.

## 2. Naming

Tests SHOULD embed conformance IDs, e.g. `TestAMPL_IDX_004_MultipleNamedDimensions`.

## 3. Required v0.2 cases

### Lexical

- AMPL-LEX-001: UTF-8 comments accepted; non-ASCII identifier explicitly rejected.
- AMPL-LEX-002: `x` and `X` distinct.
- AMPL-LEX-003/004: `#` and `/*...*/` comments skipped; unterminated block comment errors.
- AMPL-LEX-005/006: free whitespace; missing semicolon errors.
- AMPL-LEX-007: `1`, `.5`, `1.0`, `1e3`, `2.5E-2` numeric parsing.
- AMPL-LEX-009: supported declaration words tokenize as keywords.

### Declarations

- AMPL-SET-001: `set I;` appears in AST/IR.
- AMPL-PARAM-001..005: scalar, anonymous/named/multidimensional parameters and validation bounds; variable-dependent parameter bound rejected.
- AMPL-VAR-001..005: continuous/indexed/integer/binary variables and expression bounds.
- AMPL-OBJ-001: minimize/maximize.
- AMPL-CSTR-001: `subject to` lowers equivalently to legacy alias.

### Indexing

- AMPL-IDX-001..004: `{I}`, `{i in I}`, `{I,J}`, `{i in I,j in J}` preserve ordered structured dimensions.
- AMPL-IDX-005: multidimensional references validate arity/domain mismatch.

### Expressions

- AMPL-EXPR-001: add/subtract/unary minus.
- AMPL-EXPR-002: `p*x` accepted; `x*y` rejected.
- AMPL-EXPR-003: variable-free denominator accepted; division by variable rejected.
- AMPL-EXPR-004: parentheses grouping.
- AMPL-EXPR-005: AMPL `sum {}` single/multidimensional lowering.

### Relations

- AMPL-REL-001/002: `=` and `==` normalize to canonical `=`.
- AMPL-REL-003: `<=` and `>=` rows.
- AMPL-REL-004: strict parameter validation accepted; strict solver relation rejected.
- AMPL-REL-005: ranged constraint lowers to deterministic `$lower`/`$upper` rows.

### Files and aliases

- AMPL-FILE-001: CLI compiles `.mod`.
- LEGACY-001..003: bracket declarations, `constraint`, and legacy sum remain compatible.

## 4. Golden AMPL-style model

`testdata/transport-ampl.mod` is the primary fixture and SHOULD include two sets, indexed parameters, a two-dimensional transport variable, a summed objective, and indexed supply/demand `subject to` rows in the public AMPL transportation-model shape.

## 5. Negative cases

Compiler MUST reject unknown symbols, wrong reference arity/domain, `x*y`, `1/x`, decision-variable bounds, and strict solver relations.

## 6. Canonicalization

Equivalent `subject to C: x = 1;` / `constraint C: x == 1;`, AMPL/legacy sum syntax, and brace/bracket declaration aliases MUST normalize semantically equivalently. Complete canonical JSON comparison is preferred.

## 7. Cross-runtime gate

After emitting `miplang.ir/v1alpha2`, TypeScript, Python, Java, Go, and Rust runtimes must load the same golden IR. v0.2 does not yet require solver execution.

## 8. CI completion rule

A compatibility change is complete only when compiler tests pass, legacy conformance stays green, AMPL golden output matches byte-for-byte, all five runtime loaders accept the schema, every newly Supported row has tests, and GitHub Actions is green.
