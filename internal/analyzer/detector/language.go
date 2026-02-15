package detector

import (
	"os"
	"path/filepath"
)

func DetectLanguage(path string) (string, float64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "unknown", 0, err
	}
	if !info.IsDir() {
		path = filepath.Dir(path)
	}

	checks := []struct {
		File       string
		Language   string
		Confidence float64
	}{
		{File: "go.mod", Language: "go", Confidence: 1.0},
		{File: "composer.json", Language: "php", Confidence: 1.0},
		{File: "package.json", Language: "node", Confidence: 0.9},
		{File: "pyproject.toml", Language: "python", Confidence: 1.0},
		{File: "requirements.txt", Language: "python", Confidence: 0.8},
		{File: "Cargo.toml", Language: "rust", Confidence: 1.0},
		{File: "pom.xml", Language: "java", Confidence: 1.0},
		{File: "build.gradle", Language: "java", Confidence: 1.0},
	}

	for _, check := range checks {
		if _, err := os.Stat(filepath.Join(path, check.File)); err == nil {
			return check.Language, check.Confidence, nil
		}
	}
	return "unknown", 0, nil
}
