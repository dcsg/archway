package scaffold

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"
)

type Renderer struct {
	fs fs.FS
}

type RenderResult struct {
	FilesCreated []string `json:"files_created"`
}

func NewRenderer(fsys fs.FS) *Renderer {
	return &Renderer{fs: fsys}
}

func (r *Renderer) RenderTemplate(templateDir, outputDir string, vars map[string]interface{}) (*RenderResult, error) {
	if vars == nil {
		vars = map[string]interface{}{}
	}
	manifestData, err := fs.ReadFile(r.fs, path.Join(templateDir, "manifest.yaml"))
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}
	manifest, err := ParseManifest(manifestData)
	if err != nil {
		return nil, err
	}

	for key, value := range manifest.Defaults() {
		if _, exists := vars[key]; !exists {
			vars[key] = value
		}
	}

	// Coerce string booleans to actual bools so template conditionals work correctly.
	for _, def := range manifest.Variables {
		if def.Type == "bool" {
			if v, ok := vars[def.Name]; ok {
				if s, isStr := v.(string); isStr {
					vars[def.Name] = strings.EqualFold(s, "true")
				}
			}
		}
	}

	for _, def := range manifest.Variables {
		if def.Required {
			if v, ok := vars[def.Name]; !ok || strings.TrimSpace(fmt.Sprint(v)) == "" {
				return nil, fmt.Errorf("missing required variable %q", def.Name)
			}
		}
	}

	absOutputDir, err := filepath.Abs(outputDir)
	if err != nil {
		return nil, fmt.Errorf("resolve output dir: %w", err)
	}

	filesRoot := path.Join(templateDir, "files")
	result := &RenderResult{}
	if err := fs.WalkDir(r.fs, filesRoot, func(current string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if current == filesRoot {
			return nil
		}
		rel := strings.TrimPrefix(current, filesRoot+"/")

		renderedRel, err := renderPath(rel, vars)
		if err != nil {
			return fmt.Errorf("render path %q: %w", rel, err)
		}
		dstPath := filepath.Join(outputDir, filepath.FromSlash(renderedRel))

		if err := validatePathWithinDir(dstPath, absOutputDir); err != nil {
			return err
		}

		if d.IsDir() {
			return os.MkdirAll(dstPath, 0o755)
		}

		if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
			return err
		}

		srcBytes, err := fs.ReadFile(r.fs, current)
		if err != nil {
			return err
		}

		if strings.HasSuffix(dstPath, ".tmpl") {
			dstPath = strings.TrimSuffix(dstPath, ".tmpl")
			rendered, err := executeTemplate(string(srcBytes), vars)
			if err != nil {
				return fmt.Errorf("render template %q: %w", rel, err)
			}
			// Skip files whose rendered content is empty (allows conditional file
			// inclusion by wrapping entire templates in {{if}} blocks).
			if len(strings.TrimSpace(string(rendered))) == 0 {
				return nil
			}
			if err := os.WriteFile(dstPath, rendered, 0o644); err != nil {
				return err
			}
		} else {
			if err := os.WriteFile(dstPath, srcBytes, 0o644); err != nil {
				return err
			}
		}
		result.FilesCreated = append(result.FilesCreated, dstPath)
		return nil
	}); err != nil {
		return nil, err
	}

	// Remove empty directories left behind by conditionally-skipped files.
	removeEmptyDirs(outputDir)

	return result, nil
}

// RenderComposition renders an architecture + capability composition into the output directory.
func (r *Renderer) RenderComposition(plan *CompositionPlan, outputDir string) (*RenderResult, error) {
	// Inject partials into vars so templates can use {{range .Partials.main_imports}}.
	vars := plan.Vars
	vars["Partials"] = plan.Partials

	// Set boolean flags for each capability (e.g., HasHTTPAPI = true).
	capSet := map[string]bool{}
	for _, c := range plan.Capabilities {
		capSet[c] = true
	}
	vars["SelectedCapabilities"] = plan.Capabilities

	result := &RenderResult{}

	// Render architecture files.
	archResult, err := r.renderFilesDir(path.Join(plan.ArchDir, "files"), outputDir, vars)
	if err != nil {
		return nil, fmt.Errorf("render architecture: %w", err)
	}
	result.FilesCreated = append(result.FilesCreated, archResult.FilesCreated...)

	// Render each capability's files.
	for _, capDir := range plan.CapDirs {
		capResult, err := r.renderFilesDir(path.Join(capDir, "files"), outputDir, vars)
		if err != nil {
			return nil, fmt.Errorf("render capability %s: %w", capDir, err)
		}
		result.FilesCreated = append(result.FilesCreated, capResult.FilesCreated...)
	}

	removeEmptyDirs(outputDir)
	return result, nil
}

// renderFilesDir walks a files/ directory and renders templates into outputDir.
func (r *Renderer) renderFilesDir(filesRoot, outputDir string, vars map[string]interface{}) (*RenderResult, error) {
	result := &RenderResult{}

	// Check if the directory exists.
	if _, err := fs.Stat(r.fs, filesRoot); err != nil {
		return result, nil // no files directory
	}

	absOutputDir, err := filepath.Abs(outputDir)
	if err != nil {
		return nil, fmt.Errorf("resolve output dir: %w", err)
	}

	if err := fs.WalkDir(r.fs, filesRoot, func(current string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if current == filesRoot {
			return nil
		}
		rel := strings.TrimPrefix(current, filesRoot+"/")

		renderedRel, err := renderPath(rel, vars)
		if err != nil {
			return fmt.Errorf("render path %q: %w", rel, err)
		}
		dstPath := filepath.Join(outputDir, filepath.FromSlash(renderedRel))

		if err := validatePathWithinDir(dstPath, absOutputDir); err != nil {
			return err
		}

		if d.IsDir() {
			return os.MkdirAll(dstPath, 0o755)
		}

		if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
			return err
		}

		srcBytes, err := fs.ReadFile(r.fs, current)
		if err != nil {
			return err
		}

		if strings.HasSuffix(dstPath, ".tmpl") {
			dstPath = strings.TrimSuffix(dstPath, ".tmpl")
			rendered, err := executeTemplate(string(srcBytes), vars)
			if err != nil {
				return fmt.Errorf("render template %q: %w", rel, err)
			}
			if len(strings.TrimSpace(string(rendered))) == 0 {
				return nil
			}
			if err := os.WriteFile(dstPath, rendered, 0o644); err != nil {
				return err
			}
		} else {
			if err := os.WriteFile(dstPath, srcBytes, 0o644); err != nil {
				return err
			}
		}
		result.FilesCreated = append(result.FilesCreated, dstPath)
		return nil
	}); err != nil {
		return nil, err
	}
	return result, nil
}

// removeEmptyDirs walks bottom-up and removes directories that contain no files.
func removeEmptyDirs(root string) {
	// Collect dirs in reverse depth order (deepest first).
	var dirs []string
	_ = filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			dirs = append(dirs, p)
		}
		return nil
	})
	for i := len(dirs) - 1; i >= 0; i-- {
		if dirs[i] == root {
			continue
		}
		entries, err := os.ReadDir(dirs[i])
		if err == nil && len(entries) == 0 {
			_ = os.Remove(dirs[i])
		}
	}
}

func executeTemplate(content string, vars map[string]interface{}) ([]byte, error) {
	tmpl, err := template.New("file").Funcs(templateFunctions()).Option("missingkey=zero").Parse(content)
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, vars); err != nil {
		return nil, err
	}
	return io.ReadAll(buf)
}

// validatePathWithinDir ensures the resolved path stays within the output directory.
// Prevents path traversal via template variables containing "../".
func validatePathWithinDir(target string, absDir string) error {
	absTarget, err := filepath.Abs(target)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}
	// Ensure the target is within the directory (equal to or a child of absDir).
	if !strings.HasPrefix(absTarget, absDir+string(filepath.Separator)) && absTarget != absDir {
		return fmt.Errorf("path traversal detected: %q escapes output directory", target)
	}
	return nil
}

func renderPath(rel string, vars map[string]interface{}) (string, error) {
	segments := strings.Split(rel, "/")
	for i := range segments {
		for key, value := range vars {
			token := "__" + key + "__"
			segments[i] = strings.ReplaceAll(segments[i], token, fmt.Sprint(value))
		}
		if strings.Contains(segments[i], "{{") {
			rendered, err := executeTemplate(segments[i], vars)
			if err != nil {
				return "", err
			}
			segments[i] = string(rendered)
		}
	}
	return path.Join(segments...), nil
}

func templateFunctions() template.FuncMap {
	return template.FuncMap{
		"camelCase":  camelCase,
		"snakeCase":  snakeCase,
		"pascalCase": pascalCase,
		"kebabCase":  kebabCase,
		"lower":      strings.ToLower,
		"upper":      strings.ToUpper,
		"title":      strings.Title, //nolint:staticcheck // acceptable for template helper
		"contains":   strings.Contains,
		"hasPrefix":  strings.HasPrefix,
		"hasSuffix":  strings.HasSuffix,
		"join":  strings.Join,
		"split": strings.Split,
		"now":   time.Now,
		"date": func(layout string, t time.Time) string {
			return t.Format(layout)
		},
	}
}

var wordRegexp = regexp.MustCompile(`[A-Za-z0-9]+`)

func words(value string) []string {
	chunks := wordRegexp.FindAllString(value, -1)
	for i := range chunks {
		chunks[i] = strings.ToLower(chunks[i])
	}
	return chunks
}

func pascalCase(value string) string {
	parts := words(value)
	for i := range parts {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}

func camelCase(value string) string {
	parts := words(value)
	if len(parts) == 0 {
		return ""
	}
	for i := 1; i < len(parts); i++ {
		parts[i] = strings.Title(parts[i]) //nolint:staticcheck // template helper
	}
	return strings.Join(parts, "")
}

func snakeCase(value string) string {
	return strings.Join(words(value), "_")
}

func kebabCase(value string) string {
	return strings.Join(words(value), "-")
}
