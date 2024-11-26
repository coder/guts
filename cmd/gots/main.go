package main

import (
	"fmt"
	"os"

	"github.com/coder/gots/convert"
)

func main() {
	//ctx := context.Background()
	gen, err := convert.NewGolangParser()
	if err != nil {
		panic(err)
	}

	for _, arg := range os.Args[1:] {
		err = gen.Include(arg, true)
		if err != nil {
			panic(err)
		}
	}

	err = gen.IncludeCustom(map[string]string{
		"github.com/google/uuid.UUID": "string",
	})
	if err != nil {
		panic(err)
	}

	ts, err := gen.ToTypescript()
	if err != nil {
		panic(err)
	}

	output, err := ts.Serialize()
	if err != nil {
		panic(err)
	}
	fmt.Println(output)
}
