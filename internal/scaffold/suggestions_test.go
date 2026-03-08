package scaffold

import "testing"

func TestComputeSuggestions_HTTPAPISelected(t *testing.T) {
	suggestions := ComputeSuggestions([]string{"http-api"})

	// Should suggest rate-limiting, auth-jwt, observability, testing, ci-github, linting, docker.
	if len(suggestions) < 5 {
		t.Errorf("expected at least 5 suggestions for http-api, got %d", len(suggestions))
	}

	found := map[string]bool{}
	for _, s := range suggestions {
		found[s.Capability] = true
	}
	for _, expected := range []string{"rate-limiting", "auth-jwt", "observability", "testing"} {
		if !found[expected] {
			t.Errorf("expected suggestion for %q", expected)
		}
	}
}

func TestComputeSuggestions_AlreadySelected(t *testing.T) {
	suggestions := ComputeSuggestions([]string{"http-api", "rate-limiting", "auth-jwt", "observability", "testing", "ci-github", "linting", "docker"})

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
	if !found["observability"] {
		t.Error("expected observability suggestion for mysql")
	}
	if !found["docker"] {
		t.Error("expected docker suggestion for mysql")
	}
}

func TestComputeSuggestions_NoDuplicates(t *testing.T) {
	// http-api and mysql both suggest observability.
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
