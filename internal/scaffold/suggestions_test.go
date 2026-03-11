package scaffold

import "testing"

func TestComputeSuggestions_HTTPAPISelected(t *testing.T) {
	suggestions := ComputeSuggestions([]string{"http-api"})

	// Should suggest platform, rate-limiting, auth-jwt, testing, ci-github, linting, docker.
	if len(suggestions) < 5 {
		t.Errorf("expected at least 5 suggestions for http-api, got %d", len(suggestions))
	}

	found := map[string]bool{}
	for _, s := range suggestions {
		found[s.Capability] = true
	}
	for _, expected := range []string{"platform", "rate-limiting", "auth-jwt", "testing"} {
		if !found[expected] {
			t.Errorf("expected suggestion for %q", expected)
		}
	}
}

func TestComputeSuggestions_AlreadySelected(t *testing.T) {
	suggestions := ComputeSuggestions([]string{"http-api", "rate-limiting", "auth-jwt", "platform", "bootstrap", "testing", "ci-github", "linting", "docker", "uuid", "migrations", "cors", "health", "observability", "request-id"})

	// Everything is already selected — no suggestions.
	if len(suggestions) != 0 {
		t.Errorf("expected 0 suggestions when all are selected, got %d", len(suggestions))
		for _, s := range suggestions {
			t.Logf("  unexpected: %s", s.Capability)
		}
	}
}

func TestComputeSuggestions_MySQLSelected(t *testing.T) {
	suggestions := ComputeSuggestions([]string{"mysql"})

	found := map[string]bool{}
	for _, s := range suggestions {
		found[s.Capability] = true
	}
	if !found["platform"] {
		t.Error("expected platform suggestion for mysql")
	}
	if !found["docker"] {
		t.Error("expected docker suggestion for mysql")
	}
}

func TestComputeSuggestions_NoDuplicates(t *testing.T) {
	// http-api and mysql both suggest platform.
	suggestions := ComputeSuggestions([]string{"http-api", "mysql"})

	counts := map[string]int{}
	for _, s := range suggestions {
		counts[s.Capability]++
	}
	for cap, count := range counts {
		if count > 1 {
			t.Errorf("duplicate suggestion for %q: %d times", cap, count)
		}
	}
}

func TestComputeSuggestions_Empty(t *testing.T) {
	suggestions := ComputeSuggestions(nil)
	if len(suggestions) != 0 {
		t.Errorf("expected 0 suggestions for nil input, got %d", len(suggestions))
	}
}

func TestCapabilityWarnings(t *testing.T) {
	tests := []struct {
		name     string
		caps     []string
		wantMin  int // minimum expected warnings
		wantZero bool
	}{
		{"postgres without uuid warns", []string{"postgres"}, 1, false},
		{"postgres with uuid no warn", []string{"postgres", "uuid"}, 0, true},
		{"http-api without health warns", []string{"http-api"}, 1, false},
		{"http-api with health fewer warns", []string{"http-api", "health"}, 0, false},
		{"event-bus without outbox warns", []string{"event-bus"}, 1, false},
		{"event-bus with outbox no warn", []string{"event-bus", "outbox"}, 0, true},
		{"empty caps no warnings", []string{}, 0, true},
		{"full stack fewer warnings", []string{"http-api", "postgres", "uuid", "health", "cors", "observability", "request-id", "rate-limiting"}, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warnings := CapabilityWarnings(tt.caps)
			if tt.wantZero && len(warnings) != 0 {
				t.Errorf("expected no warnings, got %d: %v", len(warnings), warnings)
			}
			if tt.wantMin > 0 && len(warnings) < tt.wantMin {
				t.Errorf("expected at least %d warnings, got %d", tt.wantMin, len(warnings))
			}
		})
	}
}

func TestComputeSuggestions(t *testing.T) {
	tests := []struct {
		name    string
		caps    []string
		wantMin int
	}{
		{"http-api suggests platform", []string{"http-api"}, 1},
		{"platform suggests bootstrap", []string{"platform"}, 1},
		{"already has suggestion no dupe", []string{"http-api", "platform", "bootstrap", "rate-limiting", "auth-jwt", "testing", "docker", "ci-github", "linting"}, 0},
		{"empty no suggestions", []string{}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestions := ComputeSuggestions(tt.caps)
			if len(suggestions) < tt.wantMin {
				t.Errorf("expected at least %d suggestions, got %d", tt.wantMin, len(suggestions))
			}
		})
	}
}
