package scaffold

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"gopkg.in/yaml.v3"
)

// languageVersionDetectors maps language names to functions that detect the
// installed version. Each returns major.minor (e.g. "1.26", "22", "1.82").
var languageVersionDetectors = map[string]func() string{
	"go":         detectGoVersion,
	"typescript": detectNodeVersion,
	"javascript": detectNodeVersion,
	"rust":       detectRustVersion,
	"python":     detectPythonVersion,
}

func detectGoVersion() string {
	if out, err := exec.Command("go", "env", "GOVERSION").Output(); err == nil {
		v := strings.TrimSpace(string(out))
		v = strings.TrimPrefix(v, "go")
		return majorMinor(v)
	}
	return majorMinor(strings.TrimPrefix(runtime.Version(), "go"))
}

func detectNodeVersion() string {
	if out, err := exec.Command("node", "--version").Output(); err == nil {
		v := strings.TrimSpace(string(out))
		v = strings.TrimPrefix(v, "v")
		return majorOnly(v)
	}
	return ""
}

func detectRustVersion() string {
	if out, err := exec.Command("rustc", "--version").Output(); err == nil {
		// "rustc 1.82.0 (..." → "1.82"
		fields := strings.Fields(strings.TrimSpace(string(out)))
		if len(fields) >= 2 {
			return majorMinor(fields[1])
		}
	}
	return ""
}

func detectPythonVersion() string {
	if out, err := exec.Command("python3", "--version").Output(); err == nil {
		// "Python 3.12.1" → "3.12"
		fields := strings.Fields(strings.TrimSpace(string(out)))
		if len(fields) >= 2 {
			return majorMinor(fields[1])
		}
	}
	return ""
}

// versionVariableToLanguage maps variable names to language detector keys.
var versionVariableToLanguage = map[string]string{
	"GoVersion":     "go",
	"NodeVersion":   "typescript",
	"RustVersion":   "rust",
	"PythonVersion": "python",
}

// resolveAutoVersion detects the installed language version based on
// the variable name (e.g. "GoVersion" → detect Go, "NodeVersion" → detect Node).
// Returns empty string if the variable name is unknown or detection fails.
func resolveAutoVersion(variableName string) string {
	lang, ok := versionVariableToLanguage[variableName]
	if !ok {
		return ""
	}
	detect, ok := languageVersionDetectors[lang]
	if !ok {
		return ""
	}
	return detect()
}

// majorMinor extracts "X.Y" from "X.Y.Z".
func majorMinor(v string) string {
	parts := strings.SplitN(v, ".", 3)
	if len(parts) >= 2 {
		return parts[0] + "." + parts[1]
	}
	return v
}

// majorOnly extracts "X" from "X.Y.Z" (used for Node LTS versions).
func majorOnly(v string) string {
	parts := strings.SplitN(v, ".", 2)
	if len(parts) >= 1 {
		return parts[0]
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
		if d == "auto" {
			d = resolveAutoVersion(variable.Name)
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
