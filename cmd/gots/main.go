package main

import (
	"fmt"
	"os"

	"github.com/coder/guts"
	"github.com/coder/guts/config"
)

func main() {
	//ctx := context.Background()
	gen, err := guts.NewGolangParser()
	if err != nil {
		panic(err)
	}

	for _, arg := range os.Args[1:] {
		err = gen.IncludeGenerate(arg)
		if err != nil {
			panic(err)
		}
	}

	gen.IncludeCustomDeclaration(config.StandardMappings())

	ts, err := gen.ToTypescript()
	if err != nil {
		panic(err)
	}

	ts.ApplyMutations(
		config.EnumLists,
		config.ExportTypes,
		config.ReadOnly,
	)

	output, err := ts.Serialize()
	if err != nil {
		panic(err)
	}
	fmt.Println(output)
}
