# MIPLang Language Specification

Version: 0.2-draft  
Status: Normative draft

## 1. Scope

MIPLang is a lightweight algebraic modeling language for linear programming (LP) and mixed-integer linear programming (MILP). A MIPLang source model is compiled to a versioned canonical intermediate representation (IR) that is consumed by language-specific runtimes.

This specification defines MIPLang itself. AMPL source compatibility is defined separately in `ampl-compatibility.md`.

MIPLang intentionally does not define nonlinear optimization semantics. Constructs that require a product, quotient, or function of decision-variable expressions beyond linear form MUST be rejected by the compiler.

The normative terms **MUST**, **MUST NOT**, **SHOULD**, **SHOULD NOT**, and **MAY** are used in their usual requirements sense.

## 2. Compilation model

```text
MIPLang source
    -> lexer
    -> parser / AST
    -> name and type resolution
    -> LP/MILP semantic validation
    -> Canonical IR
    -> runtime data binding
    -> Numeric Sparse IR
    -> solver runtime
```

The source language MUST NOT depend on a specific host programming language. TypeScript, Python, Java, Go, and Rust runtimes consume the same Canonical IR contract.

## 3. Lexical rules

### 3.1 Character set

Source files are UTF-8 text. Identifiers in v0.2 are restricted to ASCII letters, digits, and underscore for implementation simplicity. An identifier MUST start with an ASCII letter or underscore and MAY contain ASCII letters, digits, and underscores thereafter. Identifiers are case-sensitive.

### 3.2 Whitespace and comments

Whitespace separates tokens and otherwise has no semantic meaning. Each declaration MUST end with a semicolon. The compiler MUST accept `#`, legacy `//`, and non-nesting `/* ... */` comments.

### 3.3 Numeric literals

v0.2 MUST accept decimal integer and floating-point literals, including optional `e`/`E` exponents. The leading sign is parsed as unary syntax.

### 3.4 Operators and punctuation

```text
+  -  *  /
=  ==  <=  >=  <  >
:  ;  ,
{  }  [  ]  (  )
```

`=` and `==` are source aliases for equality. Canonical IR MUST normalize both to `=`.

## 4. Model declarations

A source file is a sequence of declarations. Names MUST be declared before use. v0.2 declarations are `set`, `param`, `var`, `minimize`, `maximize`, `subject to`, and the legacy `constraint` alias.

### 4.1 Sets

```text
set TRIPS;
```

Core sets are externally populated index domains. Computed set expressions are outside v0.2.

### 4.2 Parameters

```text
param capacity;
param capacity {TRIPS};
param cost {i in ORIG, j in DEST};
param cost {ORIG, DEST} >= 0;
```

Legacy `param capacity[TRIPS];` MAY be accepted. Parameter validation bounds MUST NOT depend on decision variables.

### 4.3 Variables

```text
var x {I} >= 0;
var units {I} integer >= 0;
var use {I} binary;
var Make {p in PROD} >= commit[p], <= market[p];
```

Variable bounds MAY be parameter expressions and MUST NOT depend on decision variables. Binary variables have implicit domain 0/1; contradictory explicit bounds MUST be rejected.

### 4.4 Objectives

```text
minimize TotalCost: sum {i in I} cost[i] * x[i];
maximize Profit: sum {p in PROD} profit[p] * Make[p];
```

Objective expressions MUST be linear in decision variables.

### 4.5 Constraints

Canonical syntax is `subject to`; `constraint` is a legacy alias. Both MUST lower identically. Supported forms are `<=`, `>=`, `=`, `==`, and ranged `lower <= body <= upper`. Ranged constraints are normalized deterministically to `<name>$lower` and `<name>$upper`. Strict `<`/`>` are permitted for parameter validation but MUST NOT become solver rows.

## 5. Indexing

Supported forms include `{I}`, `{i in I}`, `{i in I, j in J}`, and `{I, J}`. Bare sets create anonymous dimensions. Multiple entries denote Cartesian-product dimensions in left-to-right order.

References may be scalar or indexed: `capacity`, `capacity[i]`, `cost[i,j]`, `flow[i,j,p]`. The compiler MUST validate reference arity and named-iterator domain compatibility.

An iterator introduced by a constraint, sum, or declaration bound is visible only within that construct.

## 6. Expressions

Supported arithmetic operators are `+ - * /`, unary `-`, parentheses, references, literals, and iterated `sum`. Canonical summation is `sum {i in I} expr`; legacy `sum[i in I] { expr }` MAY remain accepted. Both MUST lower identically.

Every objective and constraint expression MUST be affine in decision variables. Multiplication is valid only when at most one operand contains decision variables. Division is valid only when the denominator contains no decision variable. Division by zero discovered during binding MUST be rejected before solver invocation.

## 7. Static semantics

The compiler MUST diagnose duplicate declarations, unknown symbols/sets, wrong reference arity, scalar/index misuse, iterator-domain mismatch, decision variables in bounds, nonlinear multiplication/division, and strict solver inequalities. Diagnostics SHOULD include source line/column where known.

## 8. Canonical IR semantics

The v0.2 target is `miplang.ir/v1alpha2`. Dimensions and indices MUST be structured arrays rather than comma-concatenated strings. Equivalent aliases (`subject to`/`constraint`, `=`/`==`, AMPL/legacy sum and indexing aliases) MUST normalize equivalently. Canonical JSON serialization MUST be deterministic.

## 9. Runtime boundary

Canonical IR remains symbolic. Runtime binding supplies concrete set members and parameter values before Numeric Sparse IR is produced for HiGHS. Language runtimes MUST NOT implement independent algebraic grammars.

## 10. Explicitly out of scope for v0.2

Nonlinear/quadratic optimization, complementarity, logical constraints, AMPL command language, solver option commands, database/table handlers, `.nl`, suffix semantics, complete AMPL data language, and exact AMPL presolve/diagnostics are out of scope.

## 11. Reference implementation

The reference compiler is dependency-light Go under `internal/lexer`, `internal/parser`, and `internal/compiler`. This specification, not parser implementation details, is the source of truth.
