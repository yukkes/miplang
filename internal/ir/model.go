package ir

const SchemaVersion = "miplang.ir/v1alpha1"

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
type Parameter struct {
	Name   string `json:"name"`
	Domain string `json:"domain,omitempty"`
}
type Variable struct {
	Name   string   `json:"name"`
	Domain string   `json:"domain,omitempty"`
	Type   string   `json:"type"`
	Lower  *float64 `json:"lower,omitempty"`
	Upper  *float64 `json:"upper,omitempty"`
}
type Iterator struct {
	Name string `json:"name"`
	Set  string `json:"set"`
}
type Constraint struct {
	Name     string    `json:"name"`
	Index    *Iterator `json:"index,omitempty"`
	Left     Expr      `json:"left"`
	Operator string    `json:"operator"`
	Right    Expr      `json:"right"`
}
type Objective struct {
	Name  string `json:"name"`
	Sense string `json:"sense"`
	Expr  Expr   `json:"expr"`
}
type Expr struct {
	Kind     string    `json:"kind"`
	Value    *float64  `json:"value,omitempty"`
	Name     string    `json:"name,omitempty"`
	Index    string    `json:"index,omitempty"`
	Operator string    `json:"operator,omitempty"`
	Left     *Expr     `json:"left,omitempty"`
	Right    *Expr     `json:"right,omitempty"`
	Operand  *Expr     `json:"operand,omitempty"`
	Iterator *Iterator `json:"iterator,omitempty"`
	Body     *Expr     `json:"body,omitempty"`
}
