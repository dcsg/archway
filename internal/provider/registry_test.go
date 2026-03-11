package provider

import (
	"context"
	"io/fs"
	"testing"
	"testing/fstest"
)

type fakeProvider struct{}

func (fakeProvider) Scaffold(_ context.Context, _ ScaffoldRequest) (*ScaffoldResponse, error) {
	return &ScaffoldResponse{}, nil
}

func (fakeProvider) Analyze(_ context.Context, _ AnalyzeRequest) (*AnalyzeResponse, error) {
	return &AnalyzeResponse{}, nil
}

func (fakeProvider) Migrate(_ context.Context, _ MigrateRequest) (*MigrateResponse, error) {
	return &MigrateResponse{}, nil
}

func (fakeProvider) GetInfo(_ context.Context) (*ProviderInfo, error) {
	return &ProviderInfo{Name: "fake", Language: "fake"}, nil
}

func (fakeProvider) GetTemplateFS() fs.FS {
	return fstest.MapFS{}
}

func TestRegistry_RegisterGetList(t *testing.T) {
	r := NewRegistry()
	p := fakeProvider{}
	r.Register("Go", p)

	got, err := r.Get("go")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got == nil {
		t.Fatal("Get() returned nil provider")
	}

	langs := r.List()
	if len(langs) != 1 || langs[0] != "go" {
		t.Fatalf("List() = %#v, want [go]", langs)
	}
}

func TestRegistry_GetUnknown(t *testing.T) {
	r := NewRegistry()
	if _, err := r.Get("unknown"); err == nil {
		t.Fatal("expected error for unknown provider")
	}
}

func TestList_ReturnsRegistered(t *testing.T) {
	r := NewRegistry()
	r.Register("python", fakeProvider{})
	r.Register("go", fakeProvider{})

	langs := r.List()
	if len(langs) != 2 {
		t.Fatalf("List() returned %d providers, want 2", len(langs))
	}
	// List returns sorted order.
	if langs[0] != "go" || langs[1] != "python" {
		t.Errorf("List() = %v, want [go python]", langs)
	}
}

func TestRegister_NilProvider(t *testing.T) {
	r := NewRegistry()
	r.Register("go", nil)
	if langs := r.List(); len(langs) != 0 {
		t.Errorf("expected empty list after nil register, got %v", langs)
	}
}

func TestRegister_EmptyLanguage(t *testing.T) {
	r := NewRegistry()
	r.Register("", fakeProvider{})
	r.Register("  ", fakeProvider{})
	if langs := r.List(); len(langs) != 0 {
		t.Errorf("expected empty list after empty language register, got %v", langs)
	}
}

func TestPackageLevelFunctions(t *testing.T) {
	// Test the package-level Register/Get/List functions that use defaultRegistry.
	Register("testlang", fakeProvider{})
	got, err := Get("testlang")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got == nil {
		t.Fatal("Get() returned nil")
	}

	langs := List()
	found := false
	for _, l := range langs {
		if l == "testlang" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("List() = %v, expected to contain 'testlang'", langs)
	}

	_, err = Get("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent provider")
	}
}

func TestRegister_Overwrite(t *testing.T) {
	r := NewRegistry()
	p1 := fakeProvider{}
	p2 := fakeProvider{}
	r.Register("go", p1)
	r.Register("go", p2)

	got, err := r.Get("go")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got == nil {
		t.Fatal("Get() returned nil after overwrite")
	}
	// Verify only one entry exists.
	langs := r.List()
	if len(langs) != 1 {
		t.Errorf("List() returned %d providers after overwrite, want 1", len(langs))
	}
}
