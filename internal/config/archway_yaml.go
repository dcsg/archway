package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// SeverityOverride configures a path-scoped severity for a violation rule.
// Severity must be one of: "must", "should", "ignore".
// Reason is required and is shown in guide output and check results.
// Paths are glob patterns (same syntax as component in: fields).
// An empty Paths slice is a catch-all and matches every file.
type SeverityOverride struct {
	Severity string   `yaml:"severity"`
	Reason   string   `yaml:"reason"`
	Paths    []string `yaml:"paths,omitempty"`
}

// SeverityOverrides maps an override key to an ordered list of overrides.
// Keys match violation Rule or Name fields directly — not capability names.
//
// Override key namespace:
//   - Dependency violations:   violation.Rule  → "dependency", "arch/domain", "arch/handler", …
//   - Architecture violations: violation.Rule  → "orphan_package", "missing_component"
//   - Structure violations:    violation.Rule  → "required_dir", "forbidden_dir"
//   - Function violations:     violation.Rule  → "max_lines", "max_params", "max_return_values"
//   - Proxy rules:             violation.RuleID → user-defined IDs from .archway/rules/ YAML files
//   - Anti-patterns:           NOT overridable (hardcoded safety rules)
//
// First match wins when evaluating a file path.
type SeverityOverrides map[string][]SeverityOverride

type ArchwayConfig struct {
	Language          string               `yaml:"language" json:"language"`
	Architecture      string               `yaml:"architecture" json:"architecture"`
	Capabilities      []string             `yaml:"capabilities,omitempty" json:"capabilities,omitempty"`
	Components        []Component          `yaml:"components" json:"components"`
	Rules             RulesConfig          `yaml:"rules,omitempty" json:"rules,omitempty"`
	Check             CheckConfig          `yaml:"check,omitempty" json:"check,omitempty"`
	Extends           []string             `yaml:"extends,omitempty" json:"extends,omitempty"`
	Templates         TemplateSourceConfig `yaml:"templates,omitempty" json:"templates,omitempty"`
	Decisions         []Decision           `yaml:"decisions,omitempty" json:"decisions,omitempty"`
	Guide             GuideConfig          `yaml:"guide" json:"guide"`
	SeverityOverrides SeverityOverrides    `yaml:"severity_overrides,omitempty" json:"severity_overrides,omitempty"`
}

// CheckConfig controls behaviour of `archway check`.
type CheckConfig struct {
	// Exclude lists path globs that the checker should ignore entirely.
	// Use this for generated code, vendored dependencies, or tooling dirs
	// that legitimately don't belong to any declared component.
	// Example: ["generated/**", "vendor/**", "tools/**"]
	Exclude []string `yaml:"exclude,omitempty" json:"exclude,omitempty"`
}

// GuideConfig controls how `archway guide` instructs AI agents.
type GuideConfig struct {
	// Mode controls the agent instruction style.
	// "passive"  — background context, architecture notes appended at end (default).
	// "audit"    — agent reads codebase on every session and leads with gap analysis.
	// "prompted" — passive mode + suggested prompts appended at the end of the guide.
	Mode             string   `yaml:"mode,omitempty" json:"mode,omitempty"`
	SuggestedPrompts bool     `yaml:"suggested-prompts,omitempty" json:"suggested_prompts,omitempty"`
	Targets          []string `yaml:"targets,omitempty" json:"targets,omitempty"`
}

// GuideMode returns the resolved guide mode, defaulting to "passive".
func (g GuideConfig) GuideMode() string {
	switch g.Mode {
	case "audit", "prompted":
		return g.Mode
	default:
		return "passive"
	}
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
	// Validate severity_overrides entries.
	validSeverities := map[string]bool{"must": true, "should": true, "ignore": true}
	for key, overrides := range cfg.SeverityOverrides {
		for i, o := range overrides {
			if !validSeverities[o.Severity] {
				errs = append(errs, fmt.Errorf("severity_overrides[%q][%d].severity %q is invalid (must be one of: must, should, ignore)", key, i, o.Severity))
			}
			if strings.TrimSpace(o.Reason) == "" {
				errs = append(errs, fmt.Errorf("severity_overrides[%q][%d].reason is required", key, i))
			}
			// Warn if a catch-all (no paths) is not the last entry — subsequent entries are unreachable.
			if len(o.Paths) == 0 && i < len(overrides)-1 {
				errs = append(errs, fmt.Errorf("severity_overrides[%q][%d] is a catch-all (no paths) but is not the last entry — subsequent entries are unreachable", key, i))
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

	if language == "typescript" {
		return defaultTypeScriptConfig(cfg, architecture)
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
		}
		cfg.Rules.Structure = StructureConfig{
			RequiredDirs:  []string{"cmd/", "domain/", "port/", "adapter/"},
			ForbiddenDirs: []string{"utils/", "helpers/"},
		}
		// cmd/ is the application entry point and wiring layer — outside the
		// hexagonal architecture boundary. Exclude it from orphan detection.
		cfg.Check = CheckConfig{Exclude: []string{"cmd/**"}}
		cfg.Templates = TemplateSourceConfig{Source: "archway/api"}
	}

	return cfg
}

func defaultTypeScriptConfig(cfg *ArchwayConfig, architecture string) *ArchwayConfig {
	switch architecture {
	case "flat":
		cfg.Components = nil
		cfg.Rules.Structure = StructureConfig{
			ForbiddenDirs: []string{"utils/", "helpers/"},
		}
		cfg.Templates = TemplateSourceConfig{Source: "archway/typescript/flat"}
	default: // hexagonal
		cfg.Components = []Component{
			{Name: "domain", In: []string{"src/domain/**"}, MayDependOn: []string{}},
			{Name: "application", In: []string{"src/application/**"}, MayDependOn: []string{"domain"}},
			{Name: "infrastructure", In: []string{"src/infrastructure/**"}, MayDependOn: []string{"domain", "application"}},
			{Name: "transport", In: []string{"src/transport/**"}, MayDependOn: []string{"domain", "application", "infrastructure"}},
		}
		cfg.Rules.Structure = StructureConfig{
			RequiredDirs:  []string{"src/domain/", "src/application/", "src/infrastructure/", "src/transport/"},
			ForbiddenDirs: []string{"utils/", "helpers/"},
		}
		cfg.Templates = TemplateSourceConfig{Source: "archway/typescript/hexagonal"}
	}
	return cfg
}
