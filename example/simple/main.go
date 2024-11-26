package main

import (
	"fmt"
	"time"

	"github.com/coder/gots"
)

// SimpleType is a simple struct with a generic type
type SimpleType[T comparable] struct {
	FieldString     string
	FieldInt        int
	FieldComparable T
	FieldTime       time.Time
}

func main() {
	golang, _ := gots.NewGolangParser()
	// Generate the typescript types for this package
	_ = golang.Include("github.com/coder/gots/example/simple", true)
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
