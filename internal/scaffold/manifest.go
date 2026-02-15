package scaffold

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type Manifest struct {
	Name        string               `yaml:"name" json:"name"`
	Description string               `yaml:"description" json:"description"`
	Language    string               `yaml:"language" json:"language"`
	Version     string               `yaml:"version" json:"version"`
	Variables   []VariableDefinition `yaml:"variables" json:"variables"`
	Hooks       []string             `yaml:"hooks,omitempty" json:"hooks,omitempty"`
}

type VariableDefinition struct {
	Name        string   `yaml:"name" json:"name"`
	Type        string   `yaml:"type" json:"type"`
	Description string   `yaml:"description" json:"description"`
	Default     string   `yaml:"default,omitempty" json:"default,omitempty"`
	Required    bool     `yaml:"required" json:"required"`
	Choices     []string `yaml:"choices,omitempty" json:"choices,omitempty"`
}

func ParseManifest(data []byte) (*Manifest, error) {
	m := &Manifest{}
	if err := yaml.Unmarshal(data, m); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}
	if strings.TrimSpace(m.Name) == "" {
		return nil, fmt.Errorf("manifest missing name")
	}
	if strings.TrimSpace(m.Language) == "" {
		return nil, fmt.Errorf("manifest missing language")
	}
	for i := range m.Variables {
		v := &m.Variables[i]
		if strings.TrimSpace(v.Type) == "" {
			v.Type = "string"
		}
	}
	return m, nil
}

func (m *Manifest) Defaults() map[string]interface{} {
	out := make(map[string]interface{}, len(m.Variables))
	for _, variable := range m.Variables {
		if variable.Default == "" {
			continue
		}
		switch variable.Type {
		case "bool":
			out[variable.Name] = strings.EqualFold(variable.Default, "true")
		default:
			out[variable.Name] = variable.Default
		}
	}
	return out
}
