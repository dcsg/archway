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
