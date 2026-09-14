package miplang

import (
	"encoding/json"
	"os"
)

type Set struct {
	Name string `json:"name"`
}
type Model struct {
	SchemaVersion string `json:"schemaVersion"`
	Name          string `json:"name"`
	Sets          []Set  `json:"sets"`
}

func Load(path string) (*Model, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var model Model
	if err := json.Unmarshal(data, &model); err != nil {
		return nil, err
	}
	return &model, nil
}
