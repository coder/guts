package guts

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/packages"
)

func (p *GoParser) CommentForObject(obj types.Object) *ast.CommentGroup {
	for _, pkg := range p.Pkgs {
		if obj.Pkg() != nil && pkg.PkgPath == obj.Pkg().Path() {
			return CommentForObject(obj, pkg)
		}
	}
	return &ast.CommentGroup{List: []*ast.Comment{}}
}

// CommentForObject returns the *ast.CommentGroup associated with the object's declaration.
// It looks up the syntax node that defines the object, and returns its Doc comment (if any).
func CommentForObject(obj types.Object, pkg *packages.Package) *ast.CommentGroup {
	if obj == nil || pkg == nil {
		return &ast.CommentGroup{List: []*ast.Comment{}}
	}

	// obj.Pos() gives us the token.Pos of the declaration
	pos := obj.Pos()
	for _, f := range pkg.Syntax {
		// File covers the object position?
		if f.Pos() <= pos && pos <= f.End() {
			// Walk the file to find the node at that position
			var found *ast.CommentGroup
			ast.Inspect(f, func(n ast.Node) bool {
				if n == nil {
					return false
				}
				if n.Pos() <= pos && pos <= n.End() {
					switch decl := n.(type) {
					case *ast.FuncDecl:
						if decl.Name != nil && decl.Name.Pos() == pos {
							found = decl.Doc
							return false
						}
					case *ast.GenDecl:
						for _, spec := range decl.Specs {
							if spec.Pos() <= pos && pos <= spec.End() {
								found = decl.Doc
								return false
							}
						}
					}
				}
				return true
			})
			return found
		}
	}
	return &ast.CommentGroup{List: []*ast.Comment{}}
}
