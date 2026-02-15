package scaffold

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type WizardConfig struct {
	Steps []WizardStep `yaml:"steps" json:"steps"`
}

type WizardStep struct {
	ID        string           `yaml:"id" json:"id"`
	Questions []WizardQuestion `yaml:"questions" json:"questions"`
}

type WizardQuestion struct {
	Variable string   `yaml:"variable" json:"variable"`
	Prompt   string   `yaml:"prompt" json:"prompt"`
	Type     string   `yaml:"type" json:"type"`
	Validate string   `yaml:"validate,omitempty" json:"validate,omitempty"`
	Options  []string `yaml:"options,omitempty" json:"options,omitempty"`
	When     string   `yaml:"when,omitempty" json:"when,omitempty"`
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
