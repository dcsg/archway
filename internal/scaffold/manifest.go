package scaffold

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"gopkg.in/yaml.v3"
)

// detectedGoVersion returns the major.minor Go version from the system.
// It tries `go env GOVERSION` first, then falls back to runtime.Version().
func detectedGoVersion() string {
	if out, err := exec.Command("go", "env", "GOVERSION").Output(); err == nil {
		v := strings.TrimSpace(string(out))
		v = strings.TrimPrefix(v, "go")
		parts := strings.SplitN(v, ".", 3)
		if len(parts) >= 2 {
			return parts[0] + "." + parts[1]
		}
		return v
	}
	v := strings.TrimPrefix(runtime.Version(), "go")
	parts := strings.SplitN(v, ".", 3)
	if len(parts) >= 2 {
		return parts[0] + "." + parts[1]
	}
	return v
}

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
		d := variable.Default
		if d == "auto" && variable.Name == "GoVersion" {
			d = detectedGoVersion()
		}
		switch variable.Type {
		case "bool":
			out[variable.Name] = strings.EqualFold(d, "true")
		default:
			out[variable.Name] = d
		}
	}
	return out
}
