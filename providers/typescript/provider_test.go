package typescript

import (
	"context"
	"testing"

	"github.com/dcsg/archway/internal/provider"
)

func TestImplementsLanguageProvider(t *testing.T) {
	var _ provider.LanguageProvider = (*TypeScriptProvider)(nil)
}

func TestImplementsVersionDetector(t *testing.T) {
	var _ provider.VersionDetector = (*TypeScriptProvider)(nil)
}

func TestGetInfo(t *testing.T) {
	p := &TypeScriptProvider{}
	info, err := p.GetInfo(context.Background())
	if err != nil {
		t.Fatalf("GetInfo() error = %v", err)
	}
	if info.Language != "typescript" {
		t.Fatalf("Language = %q, want typescript", info.Language)
	}
	if len(info.SupportedArchitectures) == 0 {
		t.Fatal("expected supported architectures")
	}

	wantArchs := map[string]bool{"hexagonal": true, "flat": true}
	for _, arch := range info.SupportedArchitectures {
		if !wantArchs[arch] {
			t.Errorf("unexpected architecture %q", arch)
		}
	}
}

func TestGetInfoTemplates(t *testing.T) {
	p := &TypeScriptProvider{}
	info, err := p.GetInfo(context.Background())
	if err != nil {
		t.Fatalf("GetInfo() error = %v", err)
	}
	// Templates list should contain at least hexagonal and flat.
	if len(info.Templates) < 2 {
		t.Fatalf("expected at least 2 templates, got %d", len(info.Templates))
	}
}

func TestRegistration(t *testing.T) {
	p, err := provider.Get("typescript")
	if err != nil {
		t.Fatalf("provider.Get(typescript) error = %v", err)
	}
	if p == nil {
		t.Fatal("expected non-nil provider")
	}
}

func TestGetTemplateFS(t *testing.T) {
	p := &TypeScriptProvider{}
	fsys := p.GetTemplateFS()
	if fsys == nil {
		t.Fatal("GetTemplateFS() returned nil")
	}
}

func TestParseNodeVersion(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"v20.11.0", "20"},
		{"v18.19.1", "18"},
		{"v22.0.0", "22"},
		{"20.0.0", "20"},
		{"", ""},
	}
	for _, tt := range tests {
		got := parseNodeVersion(tt.input)
		if got != tt.want {
			t.Errorf("parseNodeVersion(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
