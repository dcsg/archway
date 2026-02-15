package scaffold

import "testing"

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
