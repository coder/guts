package guts

import (
	"fmt"
	"go/types"
)

type referencedTypes struct {
	// ReferencedTypes is a map of package paths to a map of type strings to a boolean
	// The bool is true if it was generated, false if it was only referenced
	ReferencedTypes map[string]map[string]*referencedState
}

type referencedState struct {
	Generated bool
	Object    types.Object
}

func newReferencedTypes() *referencedTypes {
	return &referencedTypes{
		ReferencedTypes: make(map[string]map[string]*referencedState),
	}
}

func (r *referencedTypes) Remaining(next func(object types.Object) error) error {
	// Keep looping over the referenced types until we don't generate anything new
	// TODO: This could be optimized with a queue vs a full loop every time.
	tried := make(map[string]map[string]struct{})
	for {
		generatedSomething := false
		for pkg, types := range r.ReferencedTypes {
			if _, ok := tried[pkg]; !ok {
				tried[pkg] = make(map[string]struct{})
			}

			for ty, generated := range types {
				if !generated.Generated {
					if _, ok := tried[pkg][ty]; ok {
						return fmt.Errorf("circular generation detected for %s.%s, infinite loop will not end", pkg, ty)
					}
					tried[pkg][ty] = struct{}{}
					err := next(generated.Object)
					if err != nil {
						return err
					}
					generatedSomething = true
				}
			}
		}
		if !generatedSomething {
			break
		}
	}
	return nil
}

func (r *referencedTypes) MarkReferenced(ty types.Object) {
	r.ensurePkg(ty.Pkg().Path())
	if _, ok := r.ReferencedTypes[ty.Pkg().Path()][ty.Type().String()]; ok {
		return // Already added
	}
	r.ReferencedTypes[ty.Pkg().Path()][ty.Type().String()] = &referencedState{
		Generated: false,
		Object:    ty,
	}
}

func (r *referencedTypes) MarkGenerated(ty types.Object) {
	r.ensurePkg(ty.Pkg().Path())
	if _, ok := r.ReferencedTypes[ty.Pkg().Path()][ty.Type().String()]; !ok {
		r.MarkReferenced(ty)
	}
	r.ReferencedTypes[ty.Pkg().Path()][ty.Type().String()].Generated = true
}

func (r *referencedTypes) IsReferenced(ty types.Object) bool {
	r.ensurePkg(ty.Pkg().Path())
	_, ok := r.ReferencedTypes[ty.Pkg().Path()][ty.Type().String()]
	return ok
}

func (r *referencedTypes) IsGenerated(ty types.Object) bool {
	r.ensurePkg(ty.Pkg().Path())
	if _, ok := r.ReferencedTypes[ty.Pkg().Path()][ty.Type().String()]; !ok {
		return false
	}
	return r.ReferencedTypes[ty.Pkg().Path()][ty.Type().String()].Generated
}

func (r *referencedTypes) ensurePkg(pkg string) {
	if _, ok := r.ReferencedTypes[pkg]; !ok {
		r.ReferencedTypes[pkg] = make(map[string]*referencedState)
	}
}
