package bindings

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"strings"
)

type Parsed struct {
	FileSet *token.FileSet
	Pkgs    map[string]*ast.Package
}

func NewGolangParser() *Parsed {
	return &Parsed{
		FileSet: token.NewFileSet(),
		Pkgs:    make(map[string]*ast.Package),
	}
}

func (p *Parsed) Include(directory string) error {
	pkgs, err := parser.ParseDir(p.FileSet, directory, onlyGoFiles, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse directory %s: %w", directory, err)
	}
	for k, v := range pkgs {
		if _, ok := p.Pkgs[k]; ok {
			return fmt.Errorf("package %s already exists", k)
		}
		p.Pkgs[k] = v
	}
	return nil
}

type Typescript struct {
	Interfaces []Interface
}

func onlyGoFiles(fi fs.FileInfo) bool {
	return strings.HasSuffix(fi.Name(), ".go") && !strings.HasSuffix(fi.Name(), "_test.go")
}
