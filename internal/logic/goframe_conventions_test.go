package logic_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestGoFrameORMConventions(t *testing.T) {
	root := findRepoRoot(t)

	checks := []struct {
		name    string
		pattern *regexp.Regexp
		paths   []string
	}{
		{
			name:    "database writes use DO objects instead of g.Map",
			pattern: regexp.MustCompile(`(?s)\.Data\(\s*(?:g\.Map|map\[string\]any)|\bg\.Map\s*\{`),
			paths:   []string{"internal/logic"},
		},
		{
			name:    "soft delete filtering is handled by GoFrame ORM",
			pattern: regexp.MustCompile(`(?i)deleted_at\s+IS\s+NULL|WhereNull\([^)]*DeletedAt`),
			paths:   []string{"internal/logic", "internal/cache"},
		},
		{
			name:    "automatic time fields are not manually maintained in writes",
			pattern: regexp.MustCompile(`(?is)\.Data\([^)]*(?:CreatedAt|UpdatedAt|DeletedAt)|INSERT\s+INTO[^` + "`" + `]{0,500}\b(?:created_at|updated_at|deleted_at)\b|UPDATE\s+[^` + "`" + `]{0,500}\b(?:created_at|updated_at|deleted_at)\b`),
			paths:   []string{"internal/logic", "internal/cache"},
		},
		{
			name:    "errors use gerror for stack trace preservation",
			pattern: regexp.MustCompile(`\b(?:fmt\.Errorf|errors\.New)\s*\(`),
			paths:   []string{"internal/logic", "utility"},
		},
	}

	for _, check := range checks {
		check := check
		t.Run(check.name, func(t *testing.T) {
			var matches []string
			for _, dir := range check.paths {
				matches = append(matches, scanGoFiles(t, filepath.Join(root, dir), check.pattern)...)
			}
			if len(matches) > 0 {
				t.Fatalf("found GoFrame convention violations:\n%s", strings.Join(matches, "\n"))
			}
		})
	}
}

func scanGoFiles(t *testing.T, root string, pattern *regexp.Regexp) []string {
	t.Helper()
	var matches []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if loc := pattern.FindIndex(content); loc != nil {
			line := 1 + strings.Count(string(content[:loc[0]]), "\n")
			matches = append(matches, filepath.ToSlash(path)+":"+itoa(line))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("scan go files: %v", err)
	}
	return matches
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("repository root with go.mod not found")
		}
		dir = parent
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits [20]byte
	i := len(digits)
	for n > 0 {
		i--
		digits[i] = byte('0' + n%10)
		n /= 10
	}
	return string(digits[i:])
}
