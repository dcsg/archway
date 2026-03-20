package typescript

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"sort"
)

// capabilityDeps maps capabilities to the npm packages they require.
// Key format: "capability" or "capability:variant" for framework-specific deps.
var capabilityDeps = map[string]struct {
	deps    map[string]string
	devDeps map[string]string
}{
	"platform": {
		deps:    map[string]string{"pino": "^9", "envalid": "^8"},
		devDeps: map[string]string{"pino-pretty": "^11"},
	},
	"http-api:express": {
		deps:    map[string]string{"express": "^4"},
		devDeps: map[string]string{"@types/express": "^4"},
	},
	"http-api:fastify": {
		deps:    map[string]string{"fastify": "^5", "@fastify/sensible": "^6"},
		devDeps: map[string]string{},
	},
	"http-api:hono": {
		deps:    map[string]string{"hono": "^4", "@hono/node-server": "^1"},
		devDeps: map[string]string{},
	},
	"request-id": {
		deps:    map[string]string{},
		devDeps: map[string]string{},
	},
	"linting": {
		deps: map[string]string{},
		devDeps: map[string]string{
			"eslint":                           "^9",
			"@typescript-eslint/eslint-plugin": "^8",
			"@typescript-eslint/parser":        "^8",
			"prettier":                         "^3",
		},
	},
	"testing": {
		deps:    map[string]string{},
		devDeps: map[string]string{"@vitest/coverage-v8": "^2"},
	},
	"postgres": {
		deps:    map[string]string{"@prisma/client": "^6"},
		devDeps: map[string]string{"prisma": "^6"},
	},
	"mysql": {
		deps:    map[string]string{"@prisma/client": "^6"},
		devDeps: map[string]string{"prisma": "^6"},
	},
	"mongodb": {
		deps:    map[string]string{"mongoose": "^8"},
		devDeps: map[string]string{},
	},
	"redis": {
		deps:    map[string]string{"ioredis": "^5"},
		devDeps: map[string]string{"@types/ioredis": "^4"},
	},
	"sqlite": {
		deps:    map[string]string{"better-sqlite3": "^9"},
		devDeps: map[string]string{"@types/better-sqlite3": "^7"},
	},
	"auth-jwt": {
		deps:    map[string]string{"jose": "^5"},
		devDeps: map[string]string{},
	},
	"circuit-breaker": {
		deps:    map[string]string{"opossum": "^8"},
		devDeps: map[string]string{"@types/opossum": "^8"},
	},
	"retry": {
		deps:    map[string]string{"p-retry": "^6"},
		devDeps: map[string]string{},
	},
	"event-bus": {
		deps:    map[string]string{"eventemitter3": "^5"},
		devDeps: map[string]string{},
	},
	"worker": {
		deps:    map[string]string{"bullmq": "^5"},
		devDeps: map[string]string{},
	},
	"scheduler": {
		deps:    map[string]string{"node-cron": "^3"},
		devDeps: map[string]string{"@types/node-cron": "^3"},
	},
	"observability": {
		deps: map[string]string{
			"@opentelemetry/sdk-node":                   "^0.57",
			"@opentelemetry/auto-instrumentations-node": "^0.57",
		},
		devDeps: map[string]string{},
	},
	"cors:express": {
		deps:    map[string]string{"cors": "^2"},
		devDeps: map[string]string{"@types/cors": "^2"},
	},
	"cors:fastify": {
		deps:    map[string]string{"@fastify/cors": "^10"},
		devDeps: map[string]string{},
	},
	"rate-limiting:express": {
		deps:    map[string]string{"express-rate-limit": "^7"},
		devDeps: map[string]string{},
	},
	"rate-limiting:fastify": {
		deps:    map[string]string{"@fastify/rate-limit": "^10"},
		devDeps: map[string]string{},
	},
	"kafka-consumer": {
		deps:    map[string]string{"kafkajs": "^2"},
		devDeps: map[string]string{},
	},
	"uuid": {
		deps:    map[string]string{"uuid": "^10"},
		devDeps: map[string]string{"@types/uuid": "^10"},
	},
	"email-gateway": {
		deps:    map[string]string{"nodemailer": "^6"},
		devDeps: map[string]string{"@types/nodemailer": "^6"},
	},
}

// packageJSON is the minimal structure needed to inject dependencies.
type packageJSON struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	Private         bool              `json:"private"`
	Scripts         map[string]string `json:"scripts,omitempty"`
	Engines         map[string]string `json:"engines,omitempty"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

// injectPackageDependencies reads the rendered package.json, adds capability-specific
// deps, and writes it back. This avoids JSON templating complexity.
func injectPackageDependencies(outputDir string, capabilities []string, opts map[string]string) error {
	pkgPath := filepath.Join(outputDir, "package.json")
	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return fmt.Errorf("read package.json: %w", err)
	}

	var pkg packageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return fmt.Errorf("parse package.json: %w", err)
	}
	if pkg.Dependencies == nil {
		pkg.Dependencies = map[string]string{}
	}
	if pkg.DevDependencies == nil {
		pkg.DevDependencies = map[string]string{}
	}

	framework := opts["HttpFramework"]
	if framework == "" {
		framework = "express"
	}

	for _, cap := range capabilities {
		// Try framework-specific variant first, then generic.
		keys := []string{cap + ":" + framework, cap}
		for _, key := range keys {
			if entry, ok := capabilityDeps[key]; ok {
				maps.Copy(pkg.Dependencies, entry.deps)
				maps.Copy(pkg.DevDependencies, entry.devDeps)
				break
			}
		}
	}

	// Remove empty dependencies block if no deps were added.
	if len(pkg.Dependencies) == 0 {
		pkg.Dependencies = nil
	}

	out, err := marshalPackageJSON(&pkg)
	if err != nil {
		return fmt.Errorf("marshal package.json: %w", err)
	}
	if err := os.WriteFile(pkgPath, out, 0o644); err != nil {
		return fmt.Errorf("write package.json: %w", err)
	}
	return nil
}

// marshalPackageJSON produces a sorted, human-readable package.json.
func marshalPackageJSON(pkg *packageJSON) ([]byte, error) {
	// Sort dependency keys for reproducible output.
	sortedDeps := func(m map[string]string) map[string]string {
		if m == nil {
			return nil
		}
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		out := make(map[string]string, len(m))
		for _, k := range keys {
			out[k] = m[k]
		}
		return out
	}

	pkg.Dependencies = sortedDeps(pkg.Dependencies)
	pkg.DevDependencies = sortedDeps(pkg.DevDependencies)

	data, err := json.MarshalIndent(pkg, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}
