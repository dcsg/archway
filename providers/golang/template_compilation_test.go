package golang

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path"
	"strings"
	"testing"
	"text/template"

	"github.com/dcsg/archway/internal/scaffold"
)

// sampleVars returns a full set of template variables for rendering tests.
func sampleVars() map[string]interface{} {
	return map[string]interface{}{
		"ServiceName": "testservice",
		"ModulePath":  "github.com/example/testservice",
		"GoVersion":   "1.26",
		"HasHTTP":     true,
		"HasHealth":   true,
		"HasPlatform": true,
		// Additional Has* flags to avoid empty conditional output.
		"HasGRPC":           false,
		"HasKafka":          false,
		"HasMySQL":          false,
		"HasRedis":          false,
		"HasBootstrap":      false,
		"HasPostgres":       false,
		"HasCORS":           false,
		"HasValidation":     false,
		"HasMigrations":     false,
		"HasEventBus":       false,
		"HasCircuitBreaker": false,
		"HasRetry":          false,
		"HasIdempotency":    false,
		"HasObservability":  false,
		"HasRequestID":      false,
		"HasAuditLog":       false,
		"HasWorker":         false,
		"HasScheduler":      false,
		"HasWebSocket":      false,
		"HasAPIVersioning":  false,
		"HasCQRS":           false,
		"HasOutbox":         false,
		"HasRepository":     false,
		"HasUUID":           false,
		"HasI18n":           false,
		"HasMailpit":        false,
		"Partials":          map[string][]string{},
	}
}

// templateFuncsForParsing returns the same FuncMap used by the scaffold renderer,
// so template parsing succeeds for templates that call custom functions.
func templateFuncsForParsing() template.FuncMap {
	return template.FuncMap{
		"camelCase":  func(s string) string { return s },
		"snakeCase":  func(s string) string { return s },
		"pascalCase": func(s string) string { return s },
		"kebabCase":  func(s string) string { return s },
		"lower":      strings.ToLower,
		"upper":      strings.ToUpper,
		"title":      strings.Title, //nolint:staticcheck
		"contains":   strings.Contains,
		"hasPrefix":  strings.HasPrefix,
		"hasSuffix":  strings.HasSuffix,
		"join":       strings.Join,
		"split":      strings.Split,
		"now":        func() string { return "" },
		"date":       func(string, interface{}) string { return "" },
	}
}

func TestAllTemplatesParse(t *testing.T) {
	var count int
	err := fs.WalkDir(templatesFS, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(p, ".tmpl") {
			return nil
		}

		count++
		t.Run(p, func(t *testing.T) {
			data, readErr := fs.ReadFile(templatesFS, p)
			if readErr != nil {
				t.Fatalf("read %s: %v", p, readErr)
			}
			_, parseErr := template.New(p).Funcs(templateFuncsForParsing()).Option("missingkey=zero").Parse(string(data))
			if parseErr != nil {
				t.Fatalf("parse %s: %v", p, parseErr)
			}
		})
		return nil
	})
	if err != nil {
		t.Fatalf("walk templates: %v", err)
	}
	if count == 0 {
		t.Fatal("no .tmpl files found in embedded FS")
	}
	t.Logf("validated %d .tmpl files parse successfully", count)
}

func TestArchitectureManifestsParse(t *testing.T) {
	architectures := []string{"hexagonal", "flat", "layered", "clean"}
	for _, arch := range architectures {
		t.Run(arch, func(t *testing.T) {
			manifestPath := path.Join("templates", "architectures", arch, "manifest.yaml")
			data, err := fs.ReadFile(templatesFS, manifestPath)
			if err != nil {
				t.Fatalf("read manifest: %v", err)
			}
			m, err := scaffold.ParseManifest(data)
			if err != nil {
				t.Fatalf("parse manifest: %v", err)
			}
			if m.Name == "" {
				t.Error("manifest name must not be empty")
			}
		})
	}
}

func TestCapabilityManifestsParse(t *testing.T) {
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
			data, readErr := fs.ReadFile(templatesFS, path.Join(capDir, entry.Name(), "capability.yaml"))
			if readErr != nil {
				t.Fatalf("read capability.yaml: %v", readErr)
			}
			cm, parseErr := scaffold.ParseCapabilityManifest(data)
			if parseErr != nil {
				t.Fatalf("parse capability.yaml: %v", parseErr)
			}
			if cm.Name == "" {
				t.Error("capability name must not be empty")
			}
		})
	}
}

func TestArchitectureTemplatesRender(t *testing.T) {
	architectures := []string{"hexagonal", "flat", "layered", "clean"}
	for _, arch := range architectures {
		t.Run(arch, func(t *testing.T) {
			filesDir := path.Join("templates", "architectures", arch, "files")
			vars := sampleVars()

			var rendered int
			err := fs.WalkDir(templatesFS, filesDir, func(p string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() || !strings.HasSuffix(p, ".tmpl") {
					return nil
				}
				data, readErr := fs.ReadFile(templatesFS, p)
				if readErr != nil {
					return readErr
				}
				tmpl, parseErr := template.New(p).Funcs(templateFuncsForParsing()).Option("missingkey=zero").Parse(string(data))
				if parseErr != nil {
					t.Errorf("parse %s: %v", p, parseErr)
					return nil
				}
				var buf strings.Builder
				if execErr := tmpl.Execute(&buf, vars); execErr != nil {
					t.Errorf("execute %s: %v", p, execErr)
					return nil
				}
				rendered++
				return nil
			})
			if err != nil {
				t.Fatalf("walk %s: %v", arch, err)
			}
			if rendered == 0 {
				t.Errorf("no templates rendered for architecture %s", arch)
			}
			t.Logf("rendered %d templates for %s", rendered, arch)
		})
	}
}

func TestGoTemplatesRenderValidGo(t *testing.T) {
	vars := sampleVars()
	fset := token.NewFileSet()

	var checked int
	err := fs.WalkDir(templatesFS, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(p, ".go.tmpl") {
			return nil
		}
		// Skip partials — they are code fragments, not complete Go files.
		if strings.Contains(p, "_partials") {
			return nil
		}

		data, readErr := fs.ReadFile(templatesFS, p)
		if readErr != nil {
			return readErr
		}

		tmpl, parseErr := template.New(p).Funcs(templateFuncsForParsing()).Option("missingkey=zero").Parse(string(data))
		if parseErr != nil {
			t.Errorf("template parse %s: %v", p, parseErr)
			return nil
		}

		var buf strings.Builder
		if execErr := tmpl.Execute(&buf, vars); execErr != nil {
			t.Errorf("template execute %s: %v", p, execErr)
			return nil
		}

		rendered := strings.TrimSpace(buf.String())
		// Some templates render to empty when their conditional is false; skip those.
		if rendered == "" {
			return nil
		}

		_, goErr := parser.ParseFile(fset, p, rendered, parser.AllErrors)
		if goErr != nil {
			t.Errorf("invalid Go in %s: %v\n--- rendered output ---\n%s", p, goErr, rendered)
		} else {
			checked++
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk templates: %v", err)
	}
	if checked == 0 {
		t.Fatal("no .go.tmpl files produced valid Go output")
	}
	t.Logf("validated %d .go.tmpl files render valid Go", checked)
}
