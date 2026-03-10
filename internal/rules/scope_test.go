package rules

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpandScope_SimpleGlob(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"main.go":      "package main\n",
		"main_test.go": "package main\n",
		"readme.md":    "# Readme\n",
	})

	files, err := ExpandScope([]string{"*.go"}, nil, dir, nil)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"main.go", "main_test.go"}, files)
}

func TestExpandScope_DoublestarGlob(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"main.go":             "package main\n",
		"internal/foo/bar.go": "package foo\n",
		"internal/baz.go":     "package internal\n",
	})

	files, err := ExpandScope([]string{"**/*.go"}, nil, dir, nil)
	require.NoError(t, err)
	assert.Len(t, files, 3)
	assert.Contains(t, files, "main.go")
	assert.Contains(t, files, "internal/foo/bar.go")
	assert.Contains(t, files, "internal/baz.go")
}

func TestExpandScope_WithExclude(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"main.go":      "package main\n",
		"main_test.go": "package main\n",
	})

	files, err := ExpandScope([]string{"**/*.go"}, []string{"*_test.go"}, dir, nil)
	require.NoError(t, err)
	assert.Equal(t, []string{"main.go"}, files)
}

func TestExpandScope_DirectoryPrefix(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"domain/order.go":    "package domain\n",
		"domain/user.go":     "package domain\n",
		"adapter/handler.go": "package adapter\n",
	})

	files, err := ExpandScope([]string{"domain/**/*.go"}, nil, dir, nil)
	require.NoError(t, err)
	assert.Len(t, files, 2)
	assert.Contains(t, files, "domain/order.go")
	assert.Contains(t, files, "domain/user.go")
}

func TestExpandScope_SkipsGitAndVendor(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"main.go":           "package main\n",
		".git/config":       "[core]\n",
		"vendor/dep/dep.go": "package dep\n",
	})

	files, err := ExpandScope([]string{"**/*.go"}, nil, dir, nil)
	require.NoError(t, err)
	assert.Equal(t, []string{"main.go"}, files)
}

func TestExpandScope_EmptyResult(t *testing.T) {
	dir := t.TempDir()

	files, err := ExpandScope([]string{"**/*.go"}, nil, dir, nil)
	require.NoError(t, err)
	assert.Empty(t, files)
}

func TestExpandScope_AllowedFilesFilter(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"a.go": "package main\n",
		"b.go": "package main\n",
		"c.go": "package main\n",
	})

	files, err := ExpandScope([]string{"**/*.go"}, nil, dir, []string{"a.go", "c.go"})
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"a.go", "c.go"}, files)
}

func TestExpandScope_MultipleScopes(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"domain/order.go":    "package domain\n",
		"adapter/handler.go": "package adapter\n",
		"config/config.go":   "package config\n",
	})

	files, err := ExpandScope(
		[]string{"domain/**/*.go", "adapter/**/*.go"},
		nil, dir, nil,
	)
	require.NoError(t, err)
	assert.Len(t, files, 2)
	assert.Contains(t, files, "domain/order.go")
	assert.Contains(t, files, "adapter/handler.go")
}

func TestMatchGlob(t *testing.T) {
	tests := []struct {
		pattern string
		path    string
		want    bool
	}{
		{"*.go", "main.go", true},
		{"*.go", "main.txt", false},
		{"**/*.go", "main.go", true},
		{"**/*.go", "internal/foo.go", true},
		{"**/*.go", "a/b/c.go", true},
		{"domain/**/*.go", "domain/order.go", true},
		{"domain/**/*.go", "adapter/foo.go", false},
		{"*_test.go", "main_test.go", true},
		{"*_test.go", "main.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.path, func(t *testing.T) {
			got := matchGlob(tt.pattern, tt.path)
			assert.Equal(t, tt.want, got, "matchGlob(%q, %q)", tt.pattern, tt.path)
		})
	}
}

func TestExpandScope_EmptyProjectRoot(t *testing.T) {
	// Empty string projectRoot — filepath.Walk should fail gracefully.
	files, err := ExpandScope([]string{"**/*.go"}, nil, "", nil)
	// On most systems, walking "" returns an error.
	if err != nil {
		return // graceful: returned error
	}
	// If no error, should return empty.
	assert.Empty(t, files)
}

func TestExpandScope_AllowedFilesEmptySliceVsNil(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"a.go": "package main\n",
		"b.go": "package main\n",
	})

	// nil allowedFiles = all files allowed.
	filesNil, err := ExpandScope([]string{"**/*.go"}, nil, dir, nil)
	require.NoError(t, err)
	assert.Len(t, filesNil, 2)

	// Empty slice [] = treated same as nil (len check, not nil check).
	filesEmpty, err := ExpandScope([]string{"**/*.go"}, nil, dir, []string{})
	require.NoError(t, err)
	assert.Len(t, filesEmpty, 2, "empty allowedFiles slice behaves like nil — no filtering")
}

func TestExpandScope_DeeplyNestedDoublestar(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"a/b/c/d/e/deep.go": "package deep\n",
	})

	files, err := ExpandScope([]string{"**/*.go"}, nil, dir, nil)
	require.NoError(t, err)
	assert.Equal(t, []string{"a/b/c/d/e/deep.go"}, files)
}

func TestExpandScope_ExcludeMatchesAll(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"main.go": "package main\n",
		"lib.go":  "package main\n",
	})

	files, err := ExpandScope([]string{"**/*.go"}, []string{"**/*.go"}, dir, nil)
	require.NoError(t, err)
	assert.Empty(t, files)
}

func TestExpandScope_NonExistentProjectRoot(t *testing.T) {
	files, err := ExpandScope([]string{"**/*.go"}, nil, "/nonexistent/path/that/does/not/exist", nil)
	// filepath.Walk returns an error for non-existent roots, but ExpandScope may
	// propagate or swallow it. Either way, no files should be returned.
	if err != nil {
		return // error path is fine
	}
	assert.Empty(t, files, "non-existent root should yield no files")
}

// setupTestProject is defined in grep_test.go — reused here via same package.
