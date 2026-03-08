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

func TestFlexOptionUnmarshalScalar(t *testing.T) {
	cfg, err := ParseWizard([]byte(`steps:
  - id: test
    questions:
      - variable: Choice
        prompt: Pick one
        type: select
        options:
          - foo
          - bar
`))
	if err != nil {
		t.Fatalf("ParseWizard() error = %v", err)
	}
	opts := cfg.Steps[0].Questions[0].Options
	if len(opts) != 2 {
		t.Fatalf("options len = %d, want 2", len(opts))
	}
	if opts[0].Label != "foo" || opts[0].Value != "foo" {
		t.Fatalf("option[0] = %+v, want Label=Value=foo", opts[0])
	}
	if opts[1].Label != "bar" || opts[1].Value != "bar" {
		t.Fatalf("option[1] = %+v, want Label=Value=bar", opts[1])
	}
}

func TestFlexOptionUnmarshalLabeled(t *testing.T) {
	cfg, err := ParseWizard([]byte(`steps:
  - id: test
    questions:
      - variable: Choice
        prompt: Pick one
        type: select
        options:
          - label: "API endpoints (REST, gRPC)"
            value: api
          - label: "CLI commands"
            value: cli
`))
	if err != nil {
		t.Fatalf("ParseWizard() error = %v", err)
	}
	opts := cfg.Steps[0].Questions[0].Options
	if len(opts) != 2 {
		t.Fatalf("options len = %d, want 2", len(opts))
	}
	if opts[0].Label != "API endpoints (REST, gRPC)" || opts[0].Value != "api" {
		t.Fatalf("option[0] = %+v", opts[0])
	}
	if opts[1].Label != "CLI commands" || opts[1].Value != "cli" {
		t.Fatalf("option[1] = %+v", opts[1])
	}
}

func TestParseProviderWizard(t *testing.T) {
	data := []byte(`steps:
  - id: intent
    questions:
      - variable: ProjectCapabilities
        prompt: "What capabilities?"
        type: multiselect
        options:
          - label: "API"
            value: api
          - label: "CLI"
            value: cli
routing:
  - template: cli
    when: "CLIOnly"
  - template: api
    when: 'ProjectStage == "production"'
fast_paths:
  - when: "CLIOnly"
    skip_steps: [transports, datastores]
  - when: 'ProjectStage == "prototype"'
    skip_steps: [transports]
    defaults:
      HasHTTP: true
`)
	cfg, err := ParseProviderWizard(data)
	if err != nil {
		t.Fatalf("ParseProviderWizard() error = %v", err)
	}
	if len(cfg.Steps) != 1 {
		t.Fatalf("steps len = %d, want 1", len(cfg.Steps))
	}
	if len(cfg.Routing) != 2 {
		t.Fatalf("routing len = %d, want 2", len(cfg.Routing))
	}
	if len(cfg.FastPaths) != 2 {
		t.Fatalf("fast_paths len = %d, want 2", len(cfg.FastPaths))
	}
	if cfg.FastPaths[1].Defaults["HasHTTP"] != true {
		t.Fatalf("fast_path[1].Defaults[HasHTTP] = %v, want true", cfg.FastPaths[1].Defaults["HasHTTP"])
	}
}

func TestResolveTemplate(t *testing.T) {
	cfg := &ProviderWizardConfig{
		Routing: []RouteRule{
			{Template: "cli", When: "CLIOnly"},
			{Template: "api", When: `ProjectStage == "production"`},
		},
	}

	tests := []struct {
		name    string
		state   map[string]interface{}
		want    string
		wantErr bool
	}{
		{
			name:  "CLI only",
			state: map[string]interface{}{"CLIOnly": true},
			want:  "cli",
		},
		{
			name:  "production stage",
			state: map[string]interface{}{"CLIOnly": false, "ProjectStage": "production"},
			want:  "api",
		},
		{
			name:    "no match",
			state:   map[string]interface{}{"CLIOnly": false, "ProjectStage": "unknown"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cfg.ResolveTemplate(tt.state)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("ResolveTemplate() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolveFastPath(t *testing.T) {
	cfg := &ProviderWizardConfig{
		FastPaths: []FastPath{
			{When: "CLIOnly", SkipSteps: []string{"transports"}},
			{When: `ProjectStage == "prototype"`, SkipSteps: []string{"datastores"}, Defaults: map[string]interface{}{"HasHTTP": true}},
		},
	}

	t.Run("CLI only match", func(t *testing.T) {
		fp := cfg.ResolveFastPath(map[string]interface{}{"CLIOnly": true})
		if fp == nil {
			t.Fatal("expected fast path")
		}
		if len(fp.SkipSteps) != 1 || fp.SkipSteps[0] != "transports" {
			t.Fatalf("skip_steps = %v", fp.SkipSteps)
		}
	})

	t.Run("prototype match", func(t *testing.T) {
		fp := cfg.ResolveFastPath(map[string]interface{}{"CLIOnly": false, "ProjectStage": "prototype"})
		if fp == nil {
			t.Fatal("expected fast path")
		}
		if fp.Defaults["HasHTTP"] != true {
			t.Fatalf("defaults = %v", fp.Defaults)
		}
	})

	t.Run("no match", func(t *testing.T) {
		fp := cfg.ResolveFastPath(map[string]interface{}{"CLIOnly": false, "ProjectStage": "production"})
		if fp != nil {
			t.Fatal("expected nil fast path")
		}
	})
}
