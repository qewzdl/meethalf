package architecture

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

var boundedContexts = map[string]struct{}{
	"matching":   {},
	"profile":    {},
	"moderation": {},
	"ads":        {},
	"payments":   {},
	"analytics":  {},
}

func TestUsecaseContextBoundaries(t *testing.T) {
	root, err := findModuleRoot()
	if err != nil {
		t.Fatalf("find module root: %v", err)
	}

	usecaseRoot := filepath.Join(root, "internal", "usecase")
	var violations []string

	walkErr := filepath.WalkDir(usecaseRoot, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		relPath, relErr := filepath.Rel(usecaseRoot, path)
		if relErr != nil {
			return relErr
		}
		relPath = filepath.ToSlash(relPath)
		parts := strings.Split(relPath, "/")
		if len(parts) == 0 {
			return nil
		}

		context := parts[0]
		if _, ok := boundedContexts[context]; !ok {
			return nil
		}

		imports, parseErr := parseImports(path)
		if parseErr != nil {
			return parseErr
		}

		for _, importPath := range imports {
			if !strings.HasPrefix(importPath, modulePath+"/internal/usecase/") {
				continue
			}
			trimmed := strings.TrimPrefix(importPath, modulePath+"/internal/usecase/")
			importContext := strings.Split(trimmed, "/")[0]
			if importContext == "" {
				continue
			}
			if _, ok := boundedContexts[importContext]; !ok {
				continue
			}
			if importContext != context {
				violations = append(violations, fmt.Sprintf("%s: %s imports usecase/%s", relPath, context, importContext))
			}
		}

		return nil
	})

	if walkErr != nil {
		t.Fatalf("walk usecase: %v", walkErr)
	}

	if len(violations) > 0 {
		t.Fatalf("usecase context boundaries violated:\n%s", strings.Join(violations, "\n"))
	}
}
