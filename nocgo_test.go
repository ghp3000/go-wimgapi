package go_wimgapi_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"
)

func TestNoCgoImports(t *testing.T) {
	root := "."
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" {
			return nil
		}
		fset := token.NewFileSet()
		f, parseErr := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if parseErr != nil {
			return parseErr
		}
		for _, imp := range f.Imports {
			if imp.Path != nil && imp.Path.Value == "\"C\"" {
				t.Fatalf("cgo import found in %s", path)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
