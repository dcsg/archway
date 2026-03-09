package scaffold

import (
	"fmt"
	"io/fs"
	"path"
	"strings"

	"gopkg.in/yaml.v3"
)

// CapabilityManifest describes a composable capability module.
type CapabilityManifest struct {
	Name        string               `yaml:"name" json:"name"`
	Description string               `yaml:"description" json:"description"`
	Variables   []VariableDefinition `yaml:"variables,omitempty" json:"variables,omitempty"`
	Requires    []string             `yaml:"requires,omitempty" json:"requires,omitempty"`
	Suggests    []string             `yaml:"suggests,omitempty" json:"suggests,omitempty"`
	Conflicts   []string             `yaml:"conflicts,omitempty" json:"conflicts,omitempty"`
}

// CompositionPlan holds the merged result of architecture + capabilities.
type CompositionPlan struct {
	Architecture string
	Capabilities []string
	Manifest     *Manifest              // architecture manifest
	CapManifests []CapabilityManifest   // loaded capability manifests
	Vars         map[string]interface{} // merged variables
	Partials     map[string][]string    // partial_name → rendered snippets
	ArchDir      string                 // e.g. "templates/architectures/hexagonal"
	CapDirs      []string               // e.g. ["templates/capabilities/http-api", ...]
	Warnings     []string               // non-fatal warnings (e.g. capability conflicts)
}

// ParseCapabilityManifest parses a capability.yaml file.
func ParseCapabilityManifest(data []byte) (*CapabilityManifest, error) {
	m := &CapabilityManifest{}
	if err := yaml.Unmarshal(data, m); err != nil {
		return nil, fmt.Errorf("parse capability manifest: %w", err)
	}
	if strings.TrimSpace(m.Name) == "" {
		return nil, fmt.Errorf("capability manifest missing name")
	}
	return m, nil
}

// ComposeProject builds a CompositionPlan from an architecture and a set of capabilities.
func ComposeProject(templateFS fs.FS, architecture string, capabilities []string, vars map[string]interface{}) (*CompositionPlan, error) {
	if vars == nil {
		vars = map[string]interface{}{}
	}

	archDir := path.Join("templates", "architectures", architecture)

	// Load architecture manifest.
	archManifestData, err := fs.ReadFile(templateFS, path.Join(archDir, "manifest.yaml"))
	if err != nil {
		return nil, fmt.Errorf("read architecture manifest %q: %w", architecture, err)
	}
	archManifest, err := ParseManifest(archManifestData)
	if err != nil {
		return nil, fmt.Errorf("parse architecture manifest %q: %w", architecture, err)
	}

	// Apply architecture defaults.
	for k, v := range archManifest.Defaults() {
		if _, exists := vars[k]; !exists {
			vars[k] = v
		}
	}

	// Load capability manifests.
	capManifests := make([]CapabilityManifest, 0, len(capabilities))
	capDirs := make([]string, 0, len(capabilities))
	for _, cap := range capabilities {
		capDir := path.Join("templates", "capabilities", cap)
		data, err := fs.ReadFile(templateFS, path.Join(capDir, "capability.yaml"))
		if err != nil {
			return nil, fmt.Errorf("read capability manifest %q: %w", cap, err)
		}
		cm, err := ParseCapabilityManifest(data)
		if err != nil {
			return nil, fmt.Errorf("capability %q: %w", cap, err)
		}
		capManifests = append(capManifests, *cm)
		capDirs = append(capDirs, capDir)
	}

	// Auto-resolve transitive dependencies.
	capSet := make(map[string]bool, len(capabilities))
	for _, c := range capabilities {
		capSet[c] = true
	}
	resolved := resolveCapabilityDeps(templateFS, capabilities, capSet)
	if len(resolved) > len(capabilities) {
		// Reload manifests for newly added capabilities.
		capabilities = resolved
		capManifests = capManifests[:0]
		capDirs = capDirs[:0]
		for _, cap := range capabilities {
			capDir := path.Join("templates", "capabilities", cap)
			data, err := fs.ReadFile(templateFS, path.Join(capDir, "capability.yaml"))
			if err != nil {
				return nil, fmt.Errorf("read capability manifest %q: %w", cap, err)
			}
			cm, err := ParseCapabilityManifest(data)
			if err != nil {
				return nil, fmt.Errorf("capability %q: %w", cap, err)
			}
			capManifests = append(capManifests, *cm)
			capDirs = append(capDirs, capDir)
		}
	}

	// Validate requirements and detect conflicts.
	var warnings []string
	for _, cm := range capManifests {
		for _, req := range cm.Requires {
			if !capSet[req] {
				return nil, fmt.Errorf("capability %q requires %q which is not selected", cm.Name, req)
			}
		}
		for _, conflict := range cm.Conflicts {
			if capSet[conflict] {
				warnings = append(warnings, fmt.Sprintf("capability %q conflicts with %q — is this intentional?", cm.Name, conflict))
			}
		}
	}

	// Apply capability variable defaults.
	for _, cm := range capManifests {
		for _, v := range cm.Variables {
			if v.Default != "" {
				if _, exists := vars[v.Name]; !exists {
					if v.Type == "bool" {
						vars[v.Name] = strings.EqualFold(v.Default, "true")
					} else {
						vars[v.Name] = v.Default
					}
				}
			}
		}
	}

	// Coerce string booleans for architecture variables.
	for _, def := range archManifest.Variables {
		if def.Type == "bool" {
			if v, ok := vars[def.Name]; ok {
				if s, isStr := v.(string); isStr {
					vars[def.Name] = strings.EqualFold(s, "true")
				}
			}
		}
	}

	// Validate required architecture variables.
	for _, def := range archManifest.Variables {
		if def.Required {
			if v, ok := vars[def.Name]; !ok || strings.TrimSpace(fmt.Sprint(v)) == "" {
				return nil, fmt.Errorf("missing required variable %q", def.Name)
			}
		}
	}

	// Set Has* boolean flags for backward compatibility with conditional templates.
	capFlagMap := map[string]string{
		"http-api":       "HasHTTP",
		"grpc":           "HasGRPC",
		"kafka-consumer": "HasKafka",
		"mysql":          "HasMySQL",
		"redis":          "HasRedis",
		"platform":       "HasPlatform",
		"bootstrap":      "HasBootstrap",
		"postgres":        "HasPostgres",
		"health":          "HasHealth",
		"cors":            "HasCORS",
		"validation":      "HasValidation",
		"migrations":      "HasMigrations",
		"event-bus":       "HasEventBus",
		"circuit-breaker": "HasCircuitBreaker",
		"retry":           "HasRetry",
		"idempotency":     "HasIdempotency",
		"observability":   "HasObservability",
		"request-id":      "HasRequestID",
		"audit-log":       "HasAuditLog",
		"worker":          "HasWorker",
		"scheduler":       "HasScheduler",
		"websocket":       "HasWebSocket",
		"api-versioning":  "HasAPIVersioning",
		"cqrs":            "HasCQRS",
		"outbox":          "HasOutbox",
		"repository":      "HasRepository",
		"uuid":            "HasUUID",
		"i18n":            "HasI18n",
		"mailpit":         "HasMailpit",
	}
	for _, c := range capabilities {
		if flag, ok := capFlagMap[c]; ok {
			vars[flag] = true
		}
	}

	// Collect partials from each capability.
	partials, err := collectPartials(templateFS, capDirs, vars)
	if err != nil {
		return nil, err
	}

	return &CompositionPlan{
		Architecture: architecture,
		Capabilities: capabilities,
		Manifest:     archManifest,
		CapManifests: capManifests,
		Vars:         vars,
		Partials:     partials,
		ArchDir:      archDir,
		CapDirs:      capDirs,
		Warnings:     warnings,
	}, nil
}

// collectPartials reads _partials/ directories from each capability and renders them.
func collectPartials(templateFS fs.FS, capDirs []string, vars map[string]interface{}) (map[string][]string, error) {
	partials := map[string][]string{}
	for _, capDir := range capDirs {
		partialsDir := path.Join(capDir, "_partials")
		entries, err := fs.ReadDir(templateFS, partialsDir)
		if err != nil {
			continue // no partials for this capability
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			// Strip .go.tmpl or .tmpl extension to get the partial key.
			key := strings.TrimSuffix(strings.TrimSuffix(name, ".tmpl"), ".go")
			data, err := fs.ReadFile(templateFS, path.Join(partialsDir, name))
			if err != nil {
				return nil, fmt.Errorf("read partial %s/%s: %w", capDir, name, err)
			}
			rendered, err := executeTemplate(string(data), vars)
			if err != nil {
				return nil, fmt.Errorf("render partial %s/%s: %w", capDir, name, err)
			}
			content := strings.TrimSpace(string(rendered))
			if content != "" {
				partials[key] = append(partials[key], content)
			}
		}
	}
	return partials, nil
}

// resolveCapabilityDeps walks the requires graph and adds missing transitive dependencies.
// Returns the expanded list in dependency-first order.
func resolveCapabilityDeps(templateFS fs.FS, capabilities []string, capSet map[string]bool) []string {
	// BFS to discover all transitive requires.
	queue := make([]string, len(capabilities))
	copy(queue, capabilities)
	ordered := make([]string, 0, len(capabilities))
	visited := map[string]bool{}

	for len(queue) > 0 {
		cap := queue[0]
		queue = queue[1:]
		if visited[cap] {
			continue
		}
		visited[cap] = true

		capDir := path.Join("templates", "capabilities", cap)
		data, err := fs.ReadFile(templateFS, path.Join(capDir, "capability.yaml"))
		if err != nil {
			ordered = append(ordered, cap)
			continue
		}
		cm, err := ParseCapabilityManifest(data)
		if err != nil {
			ordered = append(ordered, cap)
			continue
		}

		// Add requires that aren't yet selected.
		for _, req := range cm.Requires {
			if !capSet[req] {
				capSet[req] = true
				queue = append(queue, req)
			}
		}
		ordered = append(ordered, cap)
	}

	// Put dependencies before dependents: move auto-added deps to front.
	var deps, orig []string
	origSet := map[string]bool{}
	for _, c := range capabilities {
		origSet[c] = true
	}
	for _, c := range ordered {
		if origSet[c] {
			orig = append(orig, c)
		} else {
			deps = append(deps, c)
		}
	}
	return append(deps, orig...)
}

// Suggestions returns capabilities suggested by the selected set but not yet included.
func Suggestions(templateFS fs.FS, capabilities []string) []string {
	capSet := make(map[string]bool, len(capabilities))
	for _, c := range capabilities {
		capSet[c] = true
	}
	seen := map[string]bool{}
	var suggestions []string
	for _, cap := range capabilities {
		capDir := path.Join("templates", "capabilities", cap)
		data, err := fs.ReadFile(templateFS, path.Join(capDir, "capability.yaml"))
		if err != nil {
			continue
		}
		cm, err := ParseCapabilityManifest(data)
		if err != nil {
			continue
		}
		for _, s := range cm.Suggests {
			if !capSet[s] && !seen[s] {
				seen[s] = true
				suggestions = append(suggestions, s)
			}
		}
	}
	return suggestions
}
