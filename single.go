package guts

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"

	"golang.org/x/xerrors"

	"github.com/coder/guts/bindings"
)

// parseExpression feels a bit janky, however it enables the caller to send in
// a golang expression, eg `map[string]string`, and get back a Typescript type.
func parseExpression(expr string) (bindings.ExpressionType, error) {
	fs := token.NewFileSet()
	// This line means the expression must be an expression type, not a statement.
	// This removes the ability for things like generics. If there is a reason
	// to allow a larger subset of golang expressions and statements, this
	// can be changed.
	src := fmt.Sprintf(`package main; type check = %s;`, expr)

	asFile, err := parser.ParseFile(fs, "main.go", []byte(src), 0)
	if err != nil {
		return nil, xerrors.Errorf("parse expression: %w", err)
	}

	config := types.Config{}
	pkg, err := config.Check("main.go", fs, []*ast.File{asFile}, nil)
	if err != nil {
		return nil, xerrors.Errorf("check types: %w", err)
	}

	goParser, _ := NewGolangParser()
	goParser.fileSet = fs
	ts := Typescript{
		typescriptNodes: make(map[string]*typescriptNode),
		parsed:          goParser, // Intentionally empty
		serialized:      false,
	}
	err = ts.parse(pkg.Scope().Lookup("check"))
	if err != nil {
		return nil, xerrors.Errorf("parse: %w", err)
	}

	check, ok := ts.typescriptNodes["check"]
	if !ok {
		return nil, xerrors.Errorf("no check node")
	}

	alias, ok := check.Node.(*bindings.Alias)
	if !ok {
		return nil, xerrors.Errorf("expected alias, got %T", check)
	}

	return alias.Type, nil
}
