package scaffold

import "testing"

func TestParseWizard(t *testing.T) {
	cfg, err := ParseWizard([]byte(`steps:
  - id: basics
    questions:
      - variable: ServiceName
        prompt: Service name?
`))
	if err != nil {
		t.Fatalf("ParseWizard() error = %v", err)
	}
	if len(cfg.Steps) != 1 || len(cfg.Steps[0].Questions) != 1 {
		t.Fatalf("unexpected wizard config: %+v", cfg)
	}
	if cfg.Steps[0].Questions[0].Type != "input" {
		t.Fatalf("default question type = %q, want input", cfg.Steps[0].Questions[0].Type)
	}
}
