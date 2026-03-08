package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadArchwayYAML(t *testing.T) {
	path := filepath.Join("testdata", "archway.yaml")
	cfg, err := LoadArchwayYAML(path)
	if err != nil {
		t.Fatalf("LoadArchwayYAML() error = %v", err)
	}
	if cfg.Language != "go" {
		t.Fatalf("Language = %q, want go", cfg.Language)
	}
	if cfg.Architecture != "hexagonal" {
		t.Fatalf("Architecture = %q, want hexagonal", cfg.Architecture)
	}
	if len(cfg.Components) != 5 {
		t.Fatalf("Components len = %d, want 5", len(cfg.Components))
	}
	if cfg.Components[0].Name != "domain" {
		t.Fatalf("Components[0].Name = %q, want domain", cfg.Components[0].Name)
	}
}

func TestValidateArchwayYAML(t *testing.T) {
	errs := ValidateArchwayYAML(&ArchwayConfig{})
	if len(errs) == 0 {
		t.Fatal("expected validation errors")
	}
	valid := DefaultArchwayConfig("go", "hexagonal")
	if got := ValidateArchwayYAML(valid); len(got) != 0 {
		t.Fatalf("unexpected validation errors: %v", got)
	}
}

func TestSaveLoadArchwayYAMLRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "archway.yaml")
	in := DefaultArchwayConfig("go", "hexagonal")
	if err := SaveArchwayYAML(path, in); err != nil {
		t.Fatalf("SaveArchwayYAML() error = %v", err)
	}
	out, err := LoadArchwayYAML(path)
	if err != nil {
		t.Fatalf("LoadArchwayYAML() error = %v", err)
	}
	if out.Language != in.Language || out.Architecture != in.Architecture {
		t.Fatalf("round-trip mismatch: in=%+v out=%+v", in, out)
	}
}

func TestFindArchwayYAML(t *testing.T) {
	dir := t.TempDir()
	project := filepath.Join(dir, "project")
	nested := filepath.Join(project, "internal", "adapter")
	if err := mkdirAll(nested); err != nil {
		t.Fatalf("mkdirAll: %v", err)
	}
	cfg := DefaultArchwayConfig("go", "hexagonal")
	path := filepath.Join(project, "archway.yaml")
	if err := SaveArchwayYAML(path, cfg); err != nil {
		t.Fatalf("SaveArchwayYAML: %v", err)
	}
	found, err := FindArchwayYAML(nested)
	if err != nil {
		t.Fatalf("FindArchwayYAML: %v", err)
	}
	if found != path {
		t.Fatalf("FindArchwayYAML = %q, want %q", found, path)
	}
}

func TestValidateComponentUniqueNames(t *testing.T) {
	cfg := DefaultArchwayConfig("go", "hexagonal")
	cfg.Components = append(cfg.Components, Component{Name: "domain", In: []string{"other/**"}, MayDependOn: []string{}})
	errs := ValidateArchwayYAML(cfg)
	if len(errs) == 0 {
		t.Fatal("expected duplicate name error")
	}
}

func TestValidateComponentNoSelfReference(t *testing.T) {
	cfg := DefaultArchwayConfig("go", "hexagonal")
	cfg.Components = []Component{
		{Name: "domain", In: []string{"domain/**"}, MayDependOn: []string{"domain"}},
	}
	errs := ValidateArchwayYAML(cfg)
	if len(errs) == 0 {
		t.Fatal("expected self-reference error")
	}
}

func TestValidateComponentUnknownDependency(t *testing.T) {
	cfg := DefaultArchwayConfig("go", "hexagonal")
	cfg.Components = []Component{
		{Name: "domain", In: []string{"domain/**"}, MayDependOn: []string{"nonexistent"}},
	}
	errs := ValidateArchwayYAML(cfg)
	if len(errs) == 0 {
		t.Fatal("expected unknown dependency error")
	}
}

func mkdirAll(path string) error {
	return os.MkdirAll(path, 0o755)
}
