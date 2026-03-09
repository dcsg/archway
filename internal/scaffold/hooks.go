package scaffold

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"text/template"
)

// safeShellValue matches values safe for shell interpolation:
// alphanumeric, hyphens, underscores, dots, forward slashes, colons, and @.
var safeShellValue = regexp.MustCompile(`^[a-zA-Z0-9._\-/:@,=]+$`)

func DefaultGoHooks() []string {
	return []string{
		"go mod tidy",
		"git init",
	}
}

// validateHookVars checks that variable values used in hook rendering
// do not contain shell metacharacters that could lead to command injection.
func validateHookVars(vars map[string]interface{}) error {
	for key, val := range vars {
		s := fmt.Sprint(val)
		if s == "" {
			continue
		}
		// Skip non-scalar values (bools, maps, slices, etc.) — only validate
		// values that would render as simple strings in shell commands.
		if val == nil {
			continue
		}
		kind := reflect.TypeOf(val).Kind()
		switch kind {
		case reflect.Map, reflect.Slice, reflect.Array, reflect.Struct, reflect.Ptr:
			continue
		}
		if _, ok := val.(bool); ok {
			continue
		}
		if !safeShellValue.MatchString(s) {
			return fmt.Errorf("unsafe variable %q for shell hook: %q contains shell metacharacters", key, s)
		}
	}
	return nil
}

func RunPostScaffoldHooks(outputDir string, hooks []string, vars map[string]interface{}) error {
	if err := validateHookVars(vars); err != nil {
		return err
	}
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
