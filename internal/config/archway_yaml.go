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
	Rules        RulesConfig          `yaml:"rules,omitempty" json:"rules,omitempty"`
	Extends      []string             `yaml:"extends,omitempty" json:"extends,omitempty"`
	Templates    TemplateSourceConfig `yaml:"templates,omitempty" json:"templates,omitempty"`
}

type RulesConfig struct {
	Dependencies []DependencyRule `yaml:"dependencies,omitempty" json:"dependencies,omitempty"`
	Naming       []NamingRule     `yaml:"naming,omitempty" json:"naming,omitempty"`
	Structure    StructureConfig  `yaml:"structure,omitempty" json:"structure,omitempty"`
	Functions    FunctionRules    `yaml:"functions,omitempty" json:"functions,omitempty"`
}

type DependencyRule struct {
	Layer       string   `yaml:"layer" json:"layer"`
	Packages    []string `yaml:"packages" json:"packages"`
	MayDependOn []string `yaml:"may_depend_on" json:"may_depend_on"`
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
	for i, dep := range cfg.Rules.Dependencies {
		if strings.TrimSpace(dep.Layer) == "" {
			errs = append(errs, fmt.Errorf("rules.dependencies[%d].layer is required", i))
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
			Dependencies: []DependencyRule{
				{Layer: "domain", Packages: []string{"domain/**"}, MayDependOn: []string{}},
				{Layer: "ports", Packages: []string{"port/**"}, MayDependOn: []string{"domain"}},
				{Layer: "adapters", Packages: []string{"adapter/**"}, MayDependOn: []string{"ports", "domain"}},
			},
			Structure: StructureConfig{
				RequiredDirs:  []string{"cmd/", "internal/domain/", "internal/port/", "internal/adapter/"},
				ForbiddenDirs: []string{"utils/", "helpers/"},
			},
			Functions: FunctionRules{MaxLines: 80, MaxParams: 4, MaxReturnValues: 2},
		},
		Templates: TemplateSourceConfig{Source: "archway/go-hexagonal"},
	}
	return cfg
}
