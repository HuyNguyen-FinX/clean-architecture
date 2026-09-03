package architecture_test

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const modulePath = "github.com/huynguyen/clean-architecture-golang/examples/mini-banking"

func TestCorePackagesDoNotDependOnOuterLayers(t *testing.T) {
	rules := []struct {
		name    string
		dir     string
		allowed func(string) bool
	}{
		{
			name: "domain only imports the standard library",
			dir:  "../account/domain",
			allowed: func(importPath string) bool {
				return isStandardLibrary(importPath)
			},
		},
		{
			name: "application imports only the standard library and domain",
			dir:  "../account/application",
			allowed: func(importPath string) bool {
				return isStandardLibrary(importPath) ||
					strings.HasPrefix(importPath, modulePath+"/internal/account/domain")
			},
		},
	}

	for _, rule := range rules {
		rule := rule
		t.Run(rule.name, func(t *testing.T) {
			imports, err := productionImports(rule.dir)
			if err != nil {
				t.Fatal(err)
			}

			for file, paths := range imports {
				for _, importPath := range paths {
					if !rule.allowed(importPath) {
						t.Errorf("%s imports forbidden package %q", file, importPath)
					}
				}
			}
		})
	}
}

func productionImports(dir string) (map[string][]string, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		return nil, err
	}

	imports := make(map[string][]string, len(files))
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}

		parsed, err := parser.ParseFile(token.NewFileSet(), file, nil, parser.ImportsOnly)
		if err != nil {
			return nil, err
		}

		for _, spec := range parsed.Imports {
			importPath, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				return nil, err
			}
			imports[file] = append(imports[file], importPath)
		}
	}

	return imports, nil
}

func isStandardLibrary(importPath string) bool {
	firstSegment, _, _ := strings.Cut(importPath, "/")
	return !strings.Contains(firstSegment, ".")
}
