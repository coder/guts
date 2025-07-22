package main

import (
	"fmt"
	"time"

	"github.com/coder/guts"
	"github.com/coder/guts/config"
)

// SimpleType is a simple struct with a generic type
type SimpleType[T comparable] struct {
	FieldString     string
	FieldInt        int
	FieldComparable T
	FieldTime       time.Time
}

type SecondaryType struct {
	FieldString string
}

func main() {
	golang, _ := guts.NewGolangParser()
	// Generate the typescript types for this package
	_ = golang.IncludeGenerate("github.com/coder/guts/example/simple")
	// Map time.Time to string
	_ = golang.IncludeCustom(map[string]string{
		"time.Time": "string",
	})
	// Exclude SecondaryType from output
	_ = golang.ExcludeCustom("github.com/coder/guts/example/simple.SecondaryType")

	// Convert the golang types to typescript AST
	ts, _ := golang.ToTypescript()

	// ApplyMutations allows adding in generation opinions to the typescript output.
	// The basic generator has no opinions, so mutations are required to make the output
	// more usable and idiomatic.
	ts.ApplyMutations(
		// Export all top level types
		config.ExportTypes,
		// Readonly changes all fields and types to being immutable.
		// Useful if the types are only used for api responses, which should
		// not be modified.
		//config.ReadOnly,
	)

	// to see the AST tree
	//ts.ForEach(func(key string, node bindings.Node) {
	//	walk.Walk(walk.PrintingVisitor(0), node.Node)
	//})

	output, _ := ts.Serialize()
	fmt.Println(output)
}
