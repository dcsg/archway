package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type ArchwayConfig struct {
	Language     string               `yaml:"language" json:"language"`
	Architecture string               `yaml:"architecture" json:"architecture"`
	Capabilities []string             `yaml:"capabilities,omitempty" json:"capabilities,omitempty"`
	Components   []Component          `yaml:"components" json:"components"`
	Rules        RulesConfig          `yaml:"rules,omitempty" json:"rules,omitempty"`
	Extends      []string             `yaml:"extends,omitempty" json:"extends,omitempty"`
	Templates    TemplateSourceConfig `yaml:"templates,omitempty" json:"templates,omitempty"`
	Decisions    []Decision           `yaml:"decisions,omitempty" json:"decisions,omitempty"`
}

// Decision represents an architecture decision gate entry.
type Decision struct {
	Topic     string `yaml:"topic" json:"topic"`
	Tier      int    `yaml:"tier" json:"tier"`
	Status    string `yaml:"status" json:"status"`
	Choice    string `yaml:"choice,omitempty" json:"choice,omitempty"`
	Rationale string `yaml:"rationale,omitempty" json:"rationale,omitempty"`
	DecidedBy string `yaml:"decided-by,omitempty" json:"decided_by,omitempty"`
	DecidedAt string `yaml:"decided-at,omitempty" json:"decided_at,omitempty"`
}

// Component defines a named architectural component with dependency rules.
type Component struct {
	Name        string   `yaml:"name" json:"name"`
	In          []string `yaml:"in" json:"in"`
	MayDependOn []string `yaml:"may_depend_on" json:"may_depend_on"`
}

type RulesConfig struct {
	Naming    []NamingRule    `yaml:"naming,omitempty" json:"naming,omitempty"`
	Structure StructureConfig `yaml:"structure,omitempty" json:"structure,omitempty"`
	Functions FunctionRules   `yaml:"functions,omitempty" json:"functions,omitempty"`
}

type NamingRule struct {
	Pattern       string `yaml:"pattern" json:"pattern"`
	MustEndWith   string `yaml:"must_end_with,omitempty" json:"must_end_with,omitempty"`
	MustStartWith string `yaml:"must_start_with,omitempty" json:"must_start_with,omitempty"`
}

type StructureConfig struct {
	RequiredDirs  []string `yaml:"required_dirs,omitempty" json:"required_dirs,omitempty"`
	ForbiddenDirs []string `yaml:"forbidden_dirs,omitempty" json:"forbidden_dirs,omitempty"`
}

type FunctionRules struct {
	MaxLines        int `yaml:"max_lines,omitempty" json:"max_lines,omitempty"`
	MaxParams       int `yaml:"max_params,omitempty" json:"max_params,omitempty"`
	MaxReturnValues int `yaml:"max_return_values,omitempty" json:"max_return_values,omitempty"`
}

type TemplateSourceConfig struct {
	Source string `yaml:"source,omitempty" json:"source,omitempty"`
}

func LoadArchwayYAML(path string) (*ArchwayConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read archway.yaml: %w", err)
	}
	cfg := &ArchwayConfig{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse archway.yaml: %w", err)
	}
	return cfg, nil
}

func SaveArchwayYAML(path string, cfg *ArchwayConfig) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}
	if errs := ValidateArchwayYAML(cfg); len(errs) > 0 {
		return errs[0]
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal archway.yaml: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write archway.yaml: %w", err)
	}
	return nil
}

func ValidateArchwayYAML(cfg *ArchwayConfig) []error {
	if cfg == nil {
		return []error{fmt.Errorf("config is nil")}
	}
	var errs []error
	if strings.TrimSpace(cfg.Language) == "" {
		errs = append(errs, fmt.Errorf("language is required"))
	}
	if strings.TrimSpace(cfg.Architecture) == "" {
		errs = append(errs, fmt.Errorf("architecture is required"))
	}
	seen := map[string]bool{}
	for i, comp := range cfg.Components {
		name := strings.TrimSpace(comp.Name)
		if name == "" {
			errs = append(errs, fmt.Errorf("components[%d].name is required", i))
			continue
		}
		if seen[name] {
			errs = append(errs, fmt.Errorf("components[%d].name %q is duplicated", i, name))
		}
		seen[name] = true
		for _, dep := range comp.MayDependOn {
			if dep == name {
				errs = append(errs, fmt.Errorf("components[%d] %q must not reference itself in may_depend_on", i, name))
			}
		}
	}
	// Validate may_depend_on references exist.
	for i, comp := range cfg.Components {
		for _, dep := range comp.MayDependOn {
			if !seen[dep] {
				errs = append(errs, fmt.Errorf("components[%d] %q references unknown component %q in may_depend_on", i, comp.Name, dep))
			}
		}
	}
	return errs
}

func FindArchwayYAML(startDir string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", fmt.Errorf("resolve start dir: %w", err)
	}
	for {
		candidate := filepath.Join(dir, "archway.yaml")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}

func DefaultArchwayConfig(language, architecture string) *ArchwayConfig {
	language = strings.TrimSpace(language)
	if language == "" {
		language = "go"
	}
	architecture = strings.TrimSpace(architecture)
	if architecture == "" {
		architecture = "hexagonal"
	}

	cfg := &ArchwayConfig{
		Language:     language,
		Architecture: architecture,
		Rules: RulesConfig{
			Functions: FunctionRules{MaxLines: 80, MaxParams: 4, MaxReturnValues: 2},
		},
	}

	switch architecture {
	case "flat":
		cfg.Components = nil
		cfg.Rules.Structure = StructureConfig{
			ForbiddenDirs: []string{"utils/", "helpers/"},
		}
		cfg.Templates = TemplateSourceConfig{Source: "archway/cli"}
	default: // hexagonal
		cfg.Components = []Component{
			{Name: "domain", In: []string{"domain/**"}, MayDependOn: []string{}},
			{Name: "ports", In: []string{"port/**"}, MayDependOn: []string{"domain"}},
			{Name: "service", In: []string{"service/**"}, MayDependOn: []string{"domain", "ports"}},
			{Name: "adapters", In: []string{"adapter/**"}, MayDependOn: []string{"ports", "domain"}},
			{Name: "platform", In: []string{"platform/**"}, MayDependOn: []string{}},
		}
		cfg.Rules.Structure = StructureConfig{
			RequiredDirs:  []string{"cmd/", "domain/", "port/", "adapter/"},
			ForbiddenDirs: []string{"utils/", "helpers/"},
		}
		cfg.Templates = TemplateSourceConfig{Source: "archway/api"}
	}

	return cfg
}
