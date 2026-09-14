# AMPL Compatibility Profile

Version: 0.2-draft  
Profile name: **AMPL LP/MILP Core**

## 1. Purpose

MIPLang is an independent language inspired by algebraic modeling systems. This profile defines which AMPL source forms MIPLang accepts and how those forms map to MIPLang semantics. It does **not** claim complete AMPL compatibility. The target is source-level LP/MILP constructs that lower deterministically to Canonical IR and ultimately HiGHS.

The profile is based only on publicly available AMPL documentation/examples; it does not rely on implementation code, binary reverse engineering, `.nl` internals, or undocumented behavior.

## 2. Status legend

| Status | Meaning |
|---|---|
| **Supported** | Required now and covered by conformance tests. |
| **Partial** | Useful subset supported; limits documented. |
| **Planned** | Intended later. |
| **Out of scope** | Not part of MIPLang's LP/MILP goal. |

## 3. Compatibility matrix

### 3.1 Lexical rules

| Feature | Status | Conformance ID |
|---|---|---|
| UTF-8 source / ASCII identifiers | Partial | AMPL-LEX-001 |
| Case-sensitive names | Supported | AMPL-LEX-002 |
| `#` comments | Supported | AMPL-LEX-003 |
| `/* ... */` comments | Supported | AMPL-LEX-004 |
| Free-form whitespace | Supported | AMPL-LEX-005 |
| Semicolon terminator | Supported | AMPL-LEX-006 |
| Decimal/scientific constants (`e/E`) | Partial | AMPL-LEX-007 |
| String literals | Planned | AMPL-LEX-008 |
| Supported reserved words | Partial | AMPL-LEX-009 |

Legacy `//` comments are also accepted.

### 3.2 Entity declarations

| Feature | Status | Conformance ID |
|---|---|---|
| `set NAME;` | Supported | AMPL-SET-001 |
| `within` / `cross` set domains | Planned | AMPL-SET-002/003 |
| scalar/indexed/multidimensional `param` | Supported | AMPL-PARAM-001..004 |
| parameter `>= <= > <` validation | Supported | AMPL-PARAM-005 |
| parameter integer/binary/default/`:=` | Planned | AMPL-PARAM-006..009 |
| continuous/indexed/integer/binary `var` | Supported | AMPL-VAR-001..004 |
| expression variable bounds | Supported | AMPL-VAR-005 |
| initial value / set-valued var domain | Planned | AMPL-VAR-006/007 |
| `minimize` / `maximize` | Supported | AMPL-OBJ-001 |
| `subject to` | Supported | AMPL-CSTR-001 |

### 3.3 Indexing

| Feature | Status | Conformance ID |
|---|---|---|
| `{I}` | Supported | AMPL-IDX-001 |
| `{i in I}` | Supported | AMPL-IDX-002 |
| `{I,J}` | Supported | AMPL-IDX-003 |
| `{i in I, j in J}` | Supported | AMPL-IDX-004 |
| multidimensional `x[i,j,...]` | Supported | AMPL-IDX-005 |
| tuple patterns / filtering / ranges / dependent domains | Planned | AMPL-IDX-006..009 |

### 3.4 Arithmetic

| Feature | Status | Conformance ID |
|---|---|---|
| `+`, `-`, unary `-` | Supported | AMPL-EXPR-001 |
| `*` with LP linearity | Supported | AMPL-EXPR-002 |
| `/` with variable-free denominator | Supported | AMPL-EXPR-003 |
| parentheses | Supported | AMPL-EXPR-004 |
| `sum {i in I} expr` | Supported | AMPL-EXPR-005 |
| `prod`, `if then else`, selected functions | Planned | AMPL-EXPR-006/008/009 |
| powers on decision variables | Out of scope | AMPL-EXPR-007 |

### 3.5 Relations

| Feature | Status | Conformance ID |
|---|---|---|
| `=` | Supported | AMPL-REL-001 |
| `==` synonym | Supported | AMPL-REL-002 |
| `<=`, `>=` | Supported | AMPL-REL-003 |
| strict parameter `<`, `>` validation | Supported | AMPL-REL-004 |
| ranged `lo <= body <= hi` | Supported | AMPL-REL-005 |
| `!=`, boolean logic | Planned | AMPL-REL-006/007 |
| implication/complementarity | Out of scope | AMPL-REL-008/009 |

### 3.6 Files/runtime

| Feature | Status | Conformance ID |
|---|---|---|
| `.mod` model file | Supported | AMPL-FILE-001 |
| `.dat` / `data;` | Planned for v0.3 | AMPL-FILE-002/003 |
| command scripts/includes | Out of core scope | AMPL-FILE-004 |
| solver-neutral symbolic model | Supported conceptually | AMPL-RUN-001 |
| HiGHS execution | Planned next milestone | AMPL-RUN-002 |
| `.nl`, presolve equivalence, suffixes, command options | Out of scope | AMPL-RUN-003..006 |

## 4. Legacy MIPLang aliases

v0.2 MAY accept v0.1 forms such as `param p[I]`, `var x[I]`, `constraint C[i in I]`, and `sum[i in I] { ... }`. New docs SHOULD use AMPL-style braces, `subject to`, and `sum {}`. Aliases MUST produce equivalent Canonical IR.

## 5. Compatibility policy

1. No silent semantic drift: accepted AMPL syntax MUST implement documented LP/MILP meaning or be rejected.
2. Unsupported is better than approximate support.
3. Compatibility is source-level, not translator/presolve/`.nl` compatibility.
4. Canonical IR is the stability boundary.
5. Every Supported row requires a conformance test.

## 6. Reference sources

Public AMPL material used to define this profile includes the AMPL Quick Introduction, AMPL Book/reference pages, public `prod.mod`, `transp.mod`, `multmip2.mod` examples, and AMPL style guide. Public examples are behavioral references only; no AMPL implementation code is copied.
