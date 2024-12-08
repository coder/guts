package main

import (
	"fmt"
	"time"

	"github.com/coder/guts"
)

// SimpleType is a simple struct with a generic type
type SimpleType[T comparable] struct {
	FieldString     string
	FieldInt        int
	FieldComparable T
	FieldTime       time.Time
}

func main() {
	golang, _ := guts.NewGolangParser()
	// Generate the typescript types for this package
	_ = golang.IncludeGenerate("github.com/coder/guts/example/simple")
	// Map time.Time to string
	_ = golang.IncludeCustom(map[string]string{
		"time.Time": "string",
	})

	ts, _ := golang.ToTypescript()

	// to see the AST tree
	//ts.ForEach(func(key string, node *convert.typescriptNode) {
	//	walk.Walk(walk.PrintingVisitor(0), node.Node)
	//})

	output, _ := ts.Serialize()
	fmt.Println(output)
}
