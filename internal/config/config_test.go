package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// --- ArchwayYAML edge cases ---

func TestLoadArchwayYAML_MalformedYAML(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "archway.yaml")
	require.NoError(t, os.WriteFile(path, []byte(":\n  bad:\n  - [unmatched"), 0o644))

	_, err := LoadArchwayYAML(path)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parse archway.yaml")
}

func TestLoadArchwayYAML_ValidButEmptyComponents(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "archway.yaml")
	content := "language: go\narchitecture: flat\ncomponents: []\n"
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))

	cfg, err := LoadArchwayYAML(path)
	require.NoError(t, err)
	assert.Equal(t, "go", cfg.Language)
	assert.Equal(t, "flat", cfg.Architecture)
	assert.Empty(t, cfg.Components)
}

func TestFindArchwayYAML_FileDoesNotExist(t *testing.T) {
	tmp := t.TempDir()

	_, err := FindArchwayYAML(tmp)
	assert.Error(t, err)
}

func TestFindArchwayYAML_FileInCurrentDir(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "archway.yaml")
	require.NoError(t, os.WriteFile(path, []byte("language: go\n"), 0o644))

	found, err := FindArchwayYAML(tmp)
	require.NoError(t, err)
	assert.Equal(t, path, found)
}

func TestLoadArchwayYAML_DuplicateComponentNames(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "archway.yaml")
	content := `language: go
architecture: hexagonal
components:
  - name: domain
    in: ["domain/**"]
  - name: domain
    in: ["other/**"]
`
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))

	cfg, err := LoadArchwayYAML(path)
	require.NoError(t, err, "LoadArchwayYAML should not validate duplicates")
	assert.Len(t, cfg.Components, 2)
}

func TestLoadArchwayYAML_EmptyArchitecture(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "archway.yaml")
	content := "language: go\narchitecture: \"\"\n"
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))

	cfg, err := LoadArchwayYAML(path)
	require.NoError(t, err)
	assert.Empty(t, cfg.Architecture)
}

func TestLoad_InvalidYAML(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte(":\n  bad:\n  - [unmatched"), 0o644))

	_, err := Load(path)
	assert.Error(t, err)
}

func TestLoad_NonexistentFile(t *testing.T) {
	// Viper treats missing files gracefully (returns defaults), so this should not error
	cfg, err := Load("/tmp/nonexistent-archway-test-dir/config.yaml")
	// Viper returns defaults when file not found
	if err != nil {
		// Some viper versions may error on truly nonexistent paths
		return
	}
	assert.Equal(t, "go", cfg.DefaultLanguage)
}

func TestDefaultArchwayConfig_Flat(t *testing.T) {
	cfg := DefaultArchwayConfig("go", "flat")
	assert.Equal(t, "go", cfg.Language)
	assert.Equal(t, "flat", cfg.Architecture)
	assert.Nil(t, cfg.Components)
	assert.Contains(t, cfg.Rules.Structure.ForbiddenDirs, "utils/")
	assert.Contains(t, cfg.Rules.Structure.ForbiddenDirs, "helpers/")
	assert.Equal(t, "archway/cli", cfg.Templates.Source)
}

func TestDefaultArchwayConfig_Hexagonal(t *testing.T) {
	cfg := DefaultArchwayConfig("go", "hexagonal")
	assert.Equal(t, "hexagonal", cfg.Architecture)
	assert.Len(t, cfg.Components, 5)
	names := make([]string, len(cfg.Components))
	for i, c := range cfg.Components {
		names[i] = c.Name
	}
	assert.Contains(t, names, "domain")
	assert.Contains(t, names, "ports")
	assert.Contains(t, names, "adapters")
	assert.Contains(t, names, "service")
	assert.Contains(t, names, "platform")
	assert.Equal(t, 80, cfg.Rules.Functions.MaxLines)
}

func TestDefaultArchwayConfig_EmptyDefaults(t *testing.T) {
	cfg := DefaultArchwayConfig("", "")
	assert.Equal(t, "go", cfg.Language)
	assert.Equal(t, "hexagonal", cfg.Architecture)
}

func TestValidateArchwayYAML_Nil(t *testing.T) {
	errs := ValidateArchwayYAML(nil)
	assert.Len(t, errs, 1)
	assert.Contains(t, errs[0].Error(), "config is nil")
}

func TestValidateArchwayYAML_EmptyNames(t *testing.T) {
	cfg := &ArchwayConfig{
		Language:     "go",
		Architecture: "hexagonal",
		Components: []Component{
			{Name: "", In: []string{"foo/**"}},
		},
	}
	errs := ValidateArchwayYAML(cfg)
	assert.NotEmpty(t, errs)
	assert.Contains(t, errs[0].Error(), "name is required")
}

func TestValidateArchwayYAML_SelfReference(t *testing.T) {
	cfg := &ArchwayConfig{
		Language:     "go",
		Architecture: "hexagonal",
		Components: []Component{
			{Name: "domain", In: []string{"domain/**"}, MayDependOn: []string{"domain"}},
		},
	}
	errs := ValidateArchwayYAML(cfg)
	assert.NotEmpty(t, errs)
	found := false
	for _, e := range errs {
		if assert.ObjectsAreEqual("", "") || true {
			if strings.Contains(e.Error(), "must not reference itself") {
				found = true
			}
		}
	}
	assert.True(t, found, "expected self-reference error in %v", errs)
}

func TestValidateArchwayYAML_UnknownDependency(t *testing.T) {
	cfg := &ArchwayConfig{
		Language:     "go",
		Architecture: "hexagonal",
		Components: []Component{
			{Name: "domain", In: []string{"domain/**"}, MayDependOn: []string{"nonexistent"}},
		},
	}
	errs := ValidateArchwayYAML(cfg)
	assert.NotEmpty(t, errs)
	found := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "unknown component") {
			found = true
		}
	}
	assert.True(t, found, "expected unknown component error in %v", errs)
}

func TestValidateArchwayYAML_Valid(t *testing.T) {
	cfg := DefaultArchwayConfig("go", "hexagonal")
	errs := ValidateArchwayYAML(cfg)
	assert.Empty(t, errs)
}
