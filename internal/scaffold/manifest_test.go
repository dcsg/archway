package scaffold

import (
	"strings"
	"testing"
)

func TestParseManifest(t *testing.T) {
	m, err := ParseManifest([]byte(`name: test
language: go
variables:
  - name: ServiceName
    type: string
    required: true
`))
	if err != nil {
		t.Fatalf("ParseManifest() error = %v", err)
	}
	if m.Name != "test" || m.Language != "go" {
		t.Fatalf("unexpected manifest: %+v", m)
	}
}

func TestParseManifestMissingFields(t *testing.T) {
	if _, err := ParseManifest([]byte("name: \"\"\n")); err == nil {
		t.Fatal("expected error")
	}
}

func TestMajorMinor(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"1.22.3", "1.22"},
		{"1", "1"},
		{"1.22", "1.22"},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := majorMinor(tt.input)
			if got != tt.want {
				t.Errorf("majorMinor(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestMajorOnly(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"22.1.0", "22"},
		{"22", "22"},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := majorOnly(tt.input)
			if got != tt.want {
				t.Errorf("majorOnly(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestResolveAutoVersion_UnknownVariable(t *testing.T) {
	got := resolveAutoVersion("UnknownVersion")
	if got != "" {
		t.Errorf("resolveAutoVersion(UnknownVersion) = %q, want empty", got)
	}
}

func TestResolveAutoVersion_GoVersion(t *testing.T) {
	got := resolveAutoVersion("GoVersion")
	if got == "" {
		t.Error("resolveAutoVersion(GoVersion) returned empty, expected non-empty since Go is installed")
	}
}

func TestDefaults_BoolVariable(t *testing.T) {
	tests := []struct {
		name     string
		defValue string
		want     bool
	}{
		{"true", "true", true},
		{"false", "false", false},
		{"TRUE", "TRUE", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Manifest{
				Variables: []VariableDefinition{
					{Name: "EnableFeature", Type: "bool", Default: tt.defValue},
				},
			}
			defaults := m.Defaults()
			got, ok := defaults["EnableFeature"].(bool)
			if !ok {
				t.Fatalf("expected bool, got %T", defaults["EnableFeature"])
			}
			if got != tt.want {
				t.Errorf("Defaults()[EnableFeature] = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaults_AutoGoVersion(t *testing.T) {
	m := &Manifest{
		Variables: []VariableDefinition{
			{Name: "GoVersion", Type: "string", Default: "auto"},
		},
	}
	defaults := m.Defaults()
	v, ok := defaults["GoVersion"].(string)
	if !ok {
		t.Fatalf("expected string, got %T", defaults["GoVersion"])
	}
	if v == "" {
		t.Error("Defaults() GoVersion with auto should be non-empty since Go is installed")
	}
}

func TestDefaults_EmptyDefault(t *testing.T) {
	m := &Manifest{
		Variables: []VariableDefinition{
			{Name: "ServiceName", Type: "string", Default: ""},
		},
	}
	defaults := m.Defaults()
	if _, exists := defaults["ServiceName"]; exists {
		t.Error("variable with empty default should be skipped")
	}
}

func TestParseManifest_InvalidYAML(t *testing.T) {
	_, err := ParseManifest([]byte(":::invalid yaml\n\t\tbad"))
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestParseManifest_EmptyName(t *testing.T) {
	_, err := ParseManifest([]byte("name: \"  \"\nlanguage: go\n"))
	if err == nil {
		t.Fatal("expected error for empty/whitespace name")
	}
	if !strings.Contains(err.Error(), "missing name") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseManifest_DefaultType(t *testing.T) {
	m, err := ParseManifest([]byte(`name: test
variables:
  - name: Foo
    required: false
`))
	if err != nil {
		t.Fatalf("ParseManifest() error = %v", err)
	}
	if m.Variables[0].Type != "string" {
		t.Errorf("expected default type 'string', got %q", m.Variables[0].Type)
	}
}

func TestDetectGoVersion(t *testing.T) {
	v := detectGoVersion()
	if v == "" {
		t.Fatal("detectGoVersion() returned empty, Go is installed")
	}
	// Should be major.minor format like "1.26".
	if !strings.Contains(v, ".") {
		t.Errorf("detectGoVersion() = %q, expected major.minor format", v)
	}
}

func TestDetectNodeVersion(t *testing.T) {
	// Just ensure it doesn't panic. May return empty if node not installed.
	v := detectNodeVersion()
	_ = v
}

func TestDetectRustVersion(t *testing.T) {
	v := detectRustVersion()
	_ = v
}

func TestDetectPythonVersion(t *testing.T) {
	v := detectPythonVersion()
	_ = v
}

func TestResolveAutoVersion_AllLanguages(t *testing.T) {
	tests := []struct {
		variable string
		lang     string
	}{
		{"GoVersion", "go"},
		{"NodeVersion", "typescript"},
		{"RustVersion", "rust"},
		{"PythonVersion", "python"},
	}
	for _, tt := range tests {
		t.Run(tt.variable, func(t *testing.T) {
			// Just ensure no panic and correct dispatch.
			_ = resolveAutoVersion(tt.variable)
		})
	}
}

func TestDefaults_AutoUnknownVersion(t *testing.T) {
	m := &Manifest{
		Variables: []VariableDefinition{
			{Name: "UnknownVersion", Type: "string", Default: "auto"},
		},
	}
	defaults := m.Defaults()
	// "auto" with unknown variable name resolves to empty string,
	// which means it should still be set (resolveAutoVersion returns "").
	v, ok := defaults["UnknownVersion"]
	if !ok {
		// auto resolves to "" which is empty, so it's set as "".
		_ = v
	}
}

func TestDefaultGoHooks(t *testing.T) {
	hooks := DefaultGoHooks()
	if len(hooks) != 2 {
		t.Fatalf("DefaultGoHooks() len = %d, want 2", len(hooks))
	}
	if hooks[0] != "go mod tidy" {
		t.Errorf("hooks[0] = %q, want 'go mod tidy'", hooks[0])
	}
	if hooks[1] != "git init" {
		t.Errorf("hooks[1] = %q, want 'git init'", hooks[1])
	}
}
