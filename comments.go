package guts

import (
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/packages"

	"github.com/coder/guts/bindings"
)

func (p *GoParser) CommentForObject(obj types.Object) []bindings.SyntheticComment {
	for _, pkg := range p.Pkgs {
		if obj.Pkg() != nil && pkg.PkgPath == obj.Pkg().Path() {
			return CommentForObject(obj, pkg)
		}
	}
	return []bindings.SyntheticComment{}
}

// CommentForObject returns the comment group associated with the object's declaration.
// For functions/methods it returns FuncDecl.Doc.
// For const/var/type it prefers Spec.Doc, else GenDecl.Doc.
// For struct/interface members it returns Field.Doc, else Field.Comment (trailing).
// Disclaimer: A lot of this code was AI generated. Feel free to improve it!
func CommentForObject(obj types.Object, pkg *packages.Package) []bindings.SyntheticComment {
	if obj == nil || pkg == nil {
		return nil
	}
	pos := obj.Pos()

	for _, f := range pkg.Syntax {
		if !covers(f, pos) {
			continue
		}

		var found []bindings.SyntheticComment
		ast.Inspect(f, func(n ast.Node) bool {
			// The decl nodes "cover" the types they comment on. So we can check quickly if
			// the node is relevant.
			if n == nil || !covers(n, pos) {
				return false
			}

			switch nd := n.(type) {
			case *ast.FuncDecl:
				// Match function/method name token exactly.
				if nd.Name != nil && nd.Name.Pos() == pos {
					found = syntheticComments(true, nd.Doc)
					return false
				}

			case *ast.GenDecl:
				// Walk specs to prefer per-spec docs over decl docs.
				for _, sp := range nd.Specs {
					if !covers(sp, pos) {
						continue
					}

					// nd.Doc are the comments for the entire type/const/var block.
					if nd.Doc != nil {
						found = append(found, syntheticComments(true, nd.Doc)...)
					}

					switch spec := sp.(type) {
					case *ast.ValueSpec:
						// const/var
						for _, name := range spec.Names {
							if name.Pos() == pos {
								found = append(found, syntheticComments(true, spec.Doc)...)
								found = append(found, syntheticComments(false, spec.Comment)...)
								return false
							}
						}

					case *ast.TypeSpec:
						// type declarations (struct/interface/alias)
						if spec.Name != nil && spec.Name.Pos() == pos {
							// comment on the type itself
							found = append(found, syntheticComments(true, spec.Doc)...)
							found = append(found, syntheticComments(false, spec.Comment)...)
							return false
						}

						// dive into members for struct/interface fields
						switch t := spec.Type.(type) {
						case *ast.StructType:
							if cg := commentForFieldList(t.Fields, pos); cg != nil {
								found = cg
								return false
							}
						case *ast.InterfaceType:
							if cg := commentForFieldList(t.Methods, pos); cg != nil {
								found = cg
								return false
							}
						}
					}
				}

				// If we saw the decl but not a more specific match, keep walking.
				return true
			}

			// Keep drilling down until we either match or run out.
			return true
		})

		return found
	}

	return nil
}

func commentForFieldList(fl *ast.FieldList, pos token.Pos) []bindings.SyntheticComment {
	if fl == nil {
		return nil
	}
	cmts := []bindings.SyntheticComment{}
	for _, fld := range fl.List {
		if !covers(fld, pos) {
			continue
		}
		// Named field or interface method: match any of the Names.
		if len(fld.Names) > 0 {
			for _, nm := range fld.Names {
				if nm.Pos() == pos {
					cmts = append(cmts, syntheticComments(true, fld.Doc)...)
					cmts = append(cmts, syntheticComments(false, fld.Comment)...)
					return cmts
				}
			}
		} else {
			// Embedded field (anonymous): no Names; match on the Type span.
			if covers(fld.Type, pos) {
				cmts = append(cmts, syntheticComments(true, fld.Doc)...)
				cmts = append(cmts, syntheticComments(false, fld.Comment)...)
				return cmts
			}
		}
	}
	return nil
}

func covers(n ast.Node, p token.Pos) bool {
	return n != nil && n.Pos() <= p && p <= n.End()
}

func syntheticComments(leading bool, grp *ast.CommentGroup) []bindings.SyntheticComment {
	cmts := []bindings.SyntheticComment{}
	if grp == nil {
		return cmts
	}
	for _, c := range grp.List {
		cmts = append(cmts, bindings.SyntheticComment{
			Leading:         leading,
			SingleLine:      !strings.Contains(c.Text, "\n"),
			Text:            normalizeCommentText(c.Text),
			TrailingNewLine: true,
		})
	}
	return cmts
}

func normalizeCommentText(text string) string {
	// TODO: Is there a better way to get just the text of the comment?
	text = strings.TrimPrefix(text, "//")
	text = strings.TrimPrefix(text, "/*")
	text = strings.TrimSuffix(text, "*/")
	return text
}
