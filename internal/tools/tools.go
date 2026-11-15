package tools

import (
	"encoding/json"
	"fmt"
)

type Tool struct {
	Name        string
	Description string
	Parameters  json.RawMessage
	Run         func(args string) (string, error)
}

type Registry struct {
	byName map[string]Tool
	order  []Tool
}

// "callers can pass any number of tools, including none
func NewRegistry(tools ...Tool) *Registry {

	byname := make(map[string]Tool)
	var toolOrder []Tool

	for _, val := range tools {
		if _, exists := byname[val.Name]; exists {
			continue
		}
		byname[val.Name] = val
		toolOrder = append(toolOrder, val)
	}

	return &Registry{
		byName: byname,
		order:  toolOrder,
	}
}

func (r *Registry) List() []Tool {
	return r.order
}

func (r *Registry) Run(name, args string) (string, error) {

	t, exists := r.byName[name]

	if !exists {
		return "", fmt.Errorf("unknown tool %q", name)
	}

	return t.Run(args)
}
