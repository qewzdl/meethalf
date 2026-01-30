package architecture

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	modulePath      = "meethalf-api"
	otherModulePath = "meethalf-telegram-bot"
)

type layer string

const (
	layerApp       layer = "app"
	layerConfig    layer = "config"
	layerDomain    layer = "domain"
	layerLogger    layer = "logger"
	layerRateLimit layer = "ratelimit"
	layerStorage   layer = "storage"
	layerTransport layer = "transport"
	layerUsecase   layer = "usecase"
)

var allowedInternalImports = map[layer]map[layer]bool{
	layerDomain: {
		layerDomain: true,
	},
	layerUsecase: {
		layerDomain:  true,
		layerUsecase: true,
	},
	layerStorage: {
		layerConfig:  true,
		layerDomain:  true,
		layerStorage: true,
	},
	layerTransport: {
		layerDomain:    true,
		layerRateLimit: true,
		layerTransport: true,
		layerUsecase:   true,
	},
	layerConfig: {
		layerConfig: true,
	},
	layerLogger: {
		layerLogger: true,
	},
	layerRateLimit: {
		layerRateLimit: true,
	},
	layerApp: {
		layerApp:       true,
		layerConfig:    true,
		layerDomain:    true,
		layerLogger:    true,
		layerRateLimit: true,
		layerStorage:   true,
		layerTransport: true,
		layerUsecase:   true,
	},
}

func TestInternalBoundaries(t *testing.T) {
	root, err := findModuleRoot()
	if err != nil {
		t.Fatalf("find module root: %v", err)
	}

	internalRoot := filepath.Join(root, "internal")
	var violations []string

	walkErr := filepath.WalkDir(internalRoot, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, relErr := filepath.Rel(internalRoot, path)
		if relErr != nil {
			return relErr
		}
		relPath = filepath.ToSlash(relPath)

		if entry.IsDir() {
			if relPath == "architecture" || strings.HasPrefix(relPath, "architecture/") {
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		layerName, _ := layerFromRel(relPath)
		fileLayer := layer(layerName)
		rules, ok := allowedInternalImports[fileLayer]
		if !ok {
			return nil
		}

		imports, parseErr := parseImports(path)
		if parseErr != nil {
			return parseErr
		}

		for _, importPath := range imports {
			if isOtherModuleImport(importPath) {
				violations = append(violations, fmt.Sprintf("%s: %s imports %s", relPath, fileLayer, importPath))
				continue
			}

			importLayer, _, ok := layerForImport(importPath)
			if !ok {
				continue
			}

			if !rules[importLayer] {
				violations = append(violations, fmt.Sprintf("%s: %s imports %s (%s)", relPath, fileLayer, importPath, importLayer))
			}
		}

		return nil
	})

	if walkErr != nil {
		t.Fatalf("walk internal: %v", walkErr)
	}

	if len(violations) > 0 {
		t.Fatalf("internal boundaries violated:\n%s", strings.Join(violations, "\n"))
	}
}

func parseImports(path string) ([]string, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}

	imports := make([]string, 0, len(file.Imports))
	for _, spec := range file.Imports {
		if spec.Path == nil {
			continue
		}
		imports = append(imports, strings.Trim(spec.Path.Value, `"`))
	}

	return imports, nil
}

func layerFromRel(rel string) (string, string) {
	rel = strings.TrimPrefix(rel, "./")
	parts := strings.Split(rel, "/")
	if len(parts) == 0 {
		return "", ""
	}
	layerName := parts[0]
	sub := ""
	if len(parts) > 1 {
		sub = parts[1]
	}
	return layerName, sub
}

func layerForImport(importPath string) (layer, string, bool) {
	prefix := modulePath + "/internal/"
	if !strings.HasPrefix(importPath, prefix) {
		return "", "", false
	}

	rel := strings.TrimPrefix(importPath, prefix)
	parts := strings.Split(rel, "/")
	if len(parts) == 0 {
		return "", "", false
	}

	importLayer := layer(parts[0])
	sub := ""
	if len(parts) > 1 {
		sub = parts[1]
	}
	return importLayer, sub, true
}

func isOtherModuleImport(importPath string) bool {
	return importPath == otherModulePath || strings.HasPrefix(importPath, otherModulePath+"/")
}

func findModuleRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}
