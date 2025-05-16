# Go Unto Ts (guts)

[![Go Reference](https://pkg.go.dev/badge/github.com/coder/guts.svg)](https://pkg.go.dev/github.com/coder/guts)

`guts` is a tool to convert golang types to typescript for enabling a consistent type definition across the frontend and backend. It is intended to be called and customized as a library, rather than as a command line executable.

See the [simple example](./example/simple) for a basic usage of the library.
```go
type SimpleType[T comparable] struct {
	FieldString     string
	FieldInt        int
	FieldComparable T
	FieldTime       time.Time
}
```

Gets converted into

```typescript
type Comparable = string | number | boolean;

// From main/main.go
interface SimpleType<T extends Comparable> {
    FieldString: string;
    FieldInt: number;
    FieldComparable: T;
    FieldTime: string;
}
```

# How to use it

`guts` is a library, not a command line utility. This is to allow configuration with code, and also helps with package resolution.

See the [simple example](./example/simple) for a basic usage of the library. A larger example can be found in the [Coder repository](https://github.com/coder/coder/blob/main/scripts/apitypings/main.go).

```go
// Step 1: Create a new Golang parser
golang, _ := guts.NewGolangParser()
// Step 2: Configure the parser
_ = golang.IncludeGenerate("github.com/coder/guts/example/simple")
// Step 3: Convert the Golang to the typescript AST
ts, _ := golang.ToTypescript()
// Step 4: Mutate the typescript AST
ts.ApplyMutations(
    config.ExportTypes, // add 'export' to all top level declarations
)
// Step 5: Serialize the typescript AST to a string
output, _ := ts.Serialize()
fmt.Println(output)
```


# How it works

`guts` first parses a set of golang packages. The Go AST is traversed to find all the types defined in the packages. 

These types are placed into a simple AST that directly maps to the typescript AST.

Using [goja](https://github.com/dop251/goja), these types are then serialized to typescript using the typescript compiler API. 


# Generator Opinions

The generator aims to do the bare minimum type conversion. An example of a common opinion, is to use types to represent enums. Without the mutation, the following is generated:

```typescript
export enum EnumString {
    EnumBar = "bar",
    EnumBaz = "baz",
    EnumFoo = "foo",
    EnumQux = "qux"
}
```

Add the mutation:
```golang
ts.ApplyMutations(
	config.EnumAsTypes,
)
output, _ := ts.Serialize()
```

And the output is:

```typescript
export type EnumString = "bar" | "baz" | "foo" | "qux";
```

# Alternative solutions

The guts package was created to offer a more flexible, programmatic alternative to existing Go-to-TypeScript code generation tools out there.

The other solutions out there function as command-line utilities with yaml configurability. `guts` is a library, giving it a much more flexible and dynamic configuration that static generators can’t easily support.

Unlike many of its counterparts, guts leverages the official TypeScript compiler under the hood, ensuring that the generated TypeScript definitions are semantically correct, syntactically valid, and aligned with the latest language features.


# Helpful notes

An incredible website to visualize the AST of typescript: https://ts-ast-viewer.com/
