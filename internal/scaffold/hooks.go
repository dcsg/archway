package scaffold

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

func DefaultGoHooks() []string {
	return []string{
		"go mod tidy",
		"git init",
	}
}

func RunPostScaffoldHooks(outputDir string, hooks []string, vars map[string]interface{}) error {
	for _, hook := range hooks {
		hook = strings.TrimSpace(hook)
		if hook == "" {
			continue
		}

		renderedHook, err := renderHook(hook, vars)
		if err != nil {
			return err
		}
		fmt.Printf("Running: %s\n", renderedHook)

		if renderedHook == "git init" {
			if _, err := os.Stat(filepath.Join(outputDir, ".git")); err == nil {
				continue
			}
		}

		cmd := exec.Command("sh", "-c", renderedHook)
		cmd.Dir = outputDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("run hook %q: %w", renderedHook, err)
		}
	}
	return nil
}

func renderHook(hook string, vars map[string]interface{}) (string, error) {
	if vars == nil {
		vars = map[string]interface{}{}
	}
	t, err := template.New("hook").Parse(hook)
	if err != nil {
		return "", fmt.Errorf("parse hook template: %w", err)
	}
	buf := &bytes.Buffer{}
	if err := t.Execute(buf, vars); err != nil {
		return "", fmt.Errorf("render hook template: %w", err)
	}
	return buf.String(), nil
}
