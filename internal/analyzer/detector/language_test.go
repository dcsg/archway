package detector

import (
	"path/filepath"
	"testing"
)

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		name     string
		dir      string
		wantLang string
		wantConf float64
	}{
		{name: "go", dir: "go", wantLang: "go", wantConf: 1.0},
		{name: "php", dir: "php", wantLang: "php", wantConf: 1.0},
		{name: "node", dir: "node", wantLang: "node", wantConf: 0.9},
		{name: "python pyproject", dir: "python-pyproject", wantLang: "python", wantConf: 1.0},
		{name: "python req", dir: "python-req", wantLang: "python", wantConf: 0.8},
		{name: "rust", dir: "rust", wantLang: "rust", wantConf: 1.0},
		{name: "java", dir: "java", wantLang: "java", wantConf: 1.0},
		{name: "unknown", dir: "unknown", wantLang: "unknown", wantConf: 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			lang, conf, err := DetectLanguage(filepath.Join("testdata", tc.dir))
			if err != nil {
				t.Fatalf("DetectLanguage() error = %v", err)
			}
			if lang != tc.wantLang {
				t.Fatalf("language = %q, want %q", lang, tc.wantLang)
			}
			if conf != tc.wantConf {
				t.Fatalf("confidence = %v, want %v", conf, tc.wantConf)
			}
		})
	}
}

func TestDetectLanguageInvalidPath(t *testing.T) {
	_, _, err := DetectLanguage("testdata/does-not-exist")
	if err == nil {
		t.Fatal("expected error")
	}
}
