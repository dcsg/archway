package golang

import (
	"io/fs"
	"path"
	"testing"

	"github.com/dcsg/archway/internal/scaffold"
)

func TestAllCapabilityManifestsParse(t *testing.T) {
	capDir := path.Join("templates", "capabilities")
	entries, err := fs.ReadDir(templatesFS, capDir)
	if err != nil {
		t.Fatalf("read capabilities dir: %v", err)
	}

	if len(entries) == 0 {
		t.Fatal("expected at least one capability directory")
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		t.Run(entry.Name(), func(t *testing.T) {
			yamlPath := path.Join(capDir, entry.Name(), "capability.yaml")
			data, err := fs.ReadFile(templatesFS, yamlPath)
			if err != nil {
				t.Fatalf("read capability.yaml: %v", err)
			}

			manifest, err := scaffold.ParseCapabilityManifest(data)
			if err != nil {
				t.Fatalf("parse capability.yaml: %v", err)
			}

			if manifest.Name == "" {
				t.Error("capability name must not be empty")
			}

			if manifest.Name != entry.Name() {
				t.Errorf("capability name %q does not match directory name %q", manifest.Name, entry.Name())
			}
		})
	}
}

func TestCIGitHubAndGitLabConflict(t *testing.T) {
	capDir := path.Join("templates", "capabilities")

	// Load ci-gitlab manifest and check it conflicts with ci-github.
	data, err := fs.ReadFile(templatesFS, path.Join(capDir, "ci-gitlab", "capability.yaml"))
	if err != nil {
		t.Fatalf("read ci-gitlab capability.yaml: %v", err)
	}

	manifest, err := scaffold.ParseCapabilityManifest(data)
	if err != nil {
		t.Fatalf("parse ci-gitlab capability.yaml: %v", err)
	}

	found := false
	for _, c := range manifest.Conflicts {
		if c == "ci-github" {
			found = true
			break
		}
	}
	if !found {
		t.Error("ci-gitlab should declare conflict with ci-github")
	}
}
