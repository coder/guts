package guts

import (
	"go/types"
	"path"
	"path/filepath"

	"github.com/coder/guts/bindings"
)

func (ts *Typescript) location(obj types.Object) bindings.Source {
	file := ts.parsed.fileSet.File(obj.Pos())
	position := file.Position(obj.Pos())
	return bindings.Source{
		// Do not use filepath, as that changes behavior based on OS
		File:     path.Join(obj.Pkg().Name(), filepath.Base(file.Name())),
		Position: position,
	}
}
