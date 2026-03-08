package scaffold

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// FlexOption supports both plain strings and labeled {label, value} in YAML.
// Plain string "foo" unmarshals as FlexOption{Label: "foo", Value: "foo"}.
// Map {label: "Foo Bar", value: "foo"} unmarshals as FlexOption{Label: "Foo Bar", Value: "foo"}.
type FlexOption struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

func (o *FlexOption) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode {
		o.Label = value.Value
		o.Value = value.Value
		return nil
	}
	type raw struct {
		Label string `yaml:"label"`
		Value string `yaml:"value"`
	}
	var r raw
	if err := value.Decode(&r); err != nil {
		return err
	}
	o.Label = r.Label
	o.Value = r.Value
	return nil
}

type WizardConfig struct {
	Steps []WizardStep `yaml:"steps" json:"steps"`
}

type WizardStep struct {
	ID        string           `yaml:"id" json:"id"`
	Questions []WizardQuestion `yaml:"questions" json:"questions"`
}

type WizardQuestion struct {
	Variable string       `yaml:"variable" json:"variable"`
	Prompt   string       `yaml:"prompt" json:"prompt"`
	Type     string       `yaml:"type" json:"type"`
	Validate string       `yaml:"validate,omitempty" json:"validate,omitempty"`
	Options  []FlexOption `yaml:"options,omitempty" json:"options,omitempty"`
	When     string       `yaml:"when,omitempty" json:"when,omitempty"`
}

func ParseWizard(data []byte) (*WizardConfig, error) {
	cfg := &WizardConfig{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse wizard: %w", err)
	}
	for i := range cfg.Steps {
		for j := range cfg.Steps[i].Questions {
			if cfg.Steps[i].Questions[j].Type == "" {
				cfg.Steps[i].Questions[j].Type = "input"
			}
		}
	}
	return cfg, nil
}

// ProviderWizardConfig is the top-level wizard that runs before template selection.
// It lives at the provider level (e.g., providers/golang/templates/wizard.yaml).
type ProviderWizardConfig struct {
	Steps     []WizardStep `yaml:"steps" json:"steps"`
	Routing   []RouteRule  `yaml:"routing" json:"routing"`
	FastPaths []FastPath   `yaml:"fast_paths,omitempty" json:"fast_paths,omitempty"`
}

type RouteRule struct {
	Template string `yaml:"template" json:"template"`
	When     string `yaml:"when" json:"when"`
}

type FastPath struct {
	When      string                 `yaml:"when" json:"when"`
	SkipSteps []string               `yaml:"skip_steps" json:"skip_steps"`
	Defaults  map[string]interface{} `yaml:"defaults,omitempty" json:"defaults,omitempty"`
}

func ParseProviderWizard(data []byte) (*ProviderWizardConfig, error) {
	cfg := &ProviderWizardConfig{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse provider wizard: %w", err)
	}
	for i := range cfg.Steps {
		for j := range cfg.Steps[i].Questions {
			if cfg.Steps[i].Questions[j].Type == "" {
				cfg.Steps[i].Questions[j].Type = "input"
			}
		}
	}
	return cfg, nil
}

// ResolveTemplate finds the first routing rule whose condition matches.
func (c *ProviderWizardConfig) ResolveTemplate(state map[string]interface{}) (string, error) {
	for _, rule := range c.Routing {
		if evaluateWhen(rule.When, state) {
			return rule.Template, nil
		}
	}
	return "", fmt.Errorf("no template matched intent answers")
}

// ResolveFastPath returns the first matching fast path, or nil if none match.
func (c *ProviderWizardConfig) ResolveFastPath(state map[string]interface{}) *FastPath {
	for i, fp := range c.FastPaths {
		if evaluateWhen(fp.When, state) {
			return &c.FastPaths[i]
		}
	}
	return nil
}
