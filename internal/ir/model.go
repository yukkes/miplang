package ir

const SchemaVersion = "miplang.ir/v1alpha2"

type Model struct {
	SchemaVersion string       `json:"schemaVersion"`
	Name          string       `json:"name"`
	Sets          []Set        `json:"sets"`
	Parameters    []Parameter  `json:"parameters"`
	Variables     []Variable   `json:"variables"`
	Constraints   []Constraint `json:"constraints"`
	Objectives    []Objective  `json:"objectives"`
}

type Set struct {
	Name string `json:"name"`
}

type Iterator struct {
	Name string `json:"name,omitempty"`
	Set  string `json:"set"`
}

type Bound struct {
	Expr   Expr `json:"expr"`
	Strict bool `json:"strict,omitempty"`
}

type Parameter struct {
	Name    string     `json:"name"`
	Domain  []Iterator `json:"domain,omitempty"`
	Integer bool       `json:"integer,omitempty"`
	Lower   *Bound     `json:"lower,omitempty"`
	Upper   *Bound     `json:"upper,omitempty"`
}

type Variable struct {
	Name   string     `json:"name"`
	Domain []Iterator `json:"domain,omitempty"`
	Type   string     `json:"type"`
	Lower  *Bound     `json:"lower,omitempty"`
	Upper  *Bound     `json:"upper,omitempty"`
}

type Constraint struct {
	Name     string     `json:"name"`
	Index    []Iterator `json:"index,omitempty"`
	Left     Expr       `json:"left"`
	Operator string     `json:"operator"`
	Right    Expr       `json:"right"`
}

type Objective struct {
	Name  string     `json:"name"`
	Sense string     `json:"sense"`
	Index []Iterator `json:"index,omitempty"`
	Expr  Expr       `json:"expr"`
}

type Expr struct {
	Kind      string     `json:"kind"`
	Value     *float64   `json:"value,omitempty"`
	Name      string     `json:"name,omitempty"`
	Indices   []string   `json:"indices,omitempty"`
	Operator  string     `json:"operator,omitempty"`
	Left      *Expr      `json:"left,omitempty"`
	Right     *Expr      `json:"right,omitempty"`
	Operand   *Expr      `json:"operand,omitempty"`
	Iterators []Iterator `json:"iterators,omitempty"`
	Body      *Expr      `json:"body,omitempty"`
}
