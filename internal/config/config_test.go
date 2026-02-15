package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.DefaultLanguage != "go" {
		t.Fatalf("DefaultLanguage = %q, want go", cfg.DefaultLanguage)
	}
}

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("ARCHWAY_DEFAULT_LANGUAGE", "python")
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.DefaultLanguage != "python" {
		t.Fatalf("DefaultLanguage = %q, want python", cfg.DefaultLanguage)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	in := &AppConfig{
		DefaultLanguage: "go",
		Verbose:         true,
		LLM: LLMConfig{
			Provider: "ollama",
			Model:    "llama3",
			APIKey:   "secret",
			BaseURL:  "http://localhost:11434/v1",
		},
	}
	if err := Save(path, in); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	st, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat config: %v", err)
	}
	if st.Mode().Perm() != 0o600 {
		t.Fatalf("config permissions = %v, want 0600", st.Mode().Perm())
	}

	out, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if out.LLM.Provider != in.LLM.Provider || out.LLM.Model != in.LLM.Model {
		t.Fatalf("round-trip mismatch: %#v vs %#v", out.LLM, in.LLM)
	}
}
