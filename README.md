# Go Unto Ts (guts)

[![Go Reference](https://pkg.go.dev/badge/github.com/coder/guts.svg)](https://pkg.go.dev/github.com/coder/guts)

`guts` is a tool to convert golang types to typescript for enabling a consistent type definition across the frontend and backend. It is intended to be called and customized as a library, rather than as a command line tool.

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

# How it works

`guts` first parses a set of golang packages. The Go AST is traversed to find all the types defined in the packages. 

These types are placed into a simple AST that directly maps to the typescript AST.

Using [goja](https://github.com/dop251/goja), these types are then converted to typescript using the typescript compiler API. 


# Generator Opinions

The generator aims to do the bare minimum type conversion. An example of a common opinion, is to create enum lists.

```typescript
export type Enum = "bar" | "baz" | "foo" | "qux" // <-- Golang type
export const Enums: Enum[] = ["bar", "baz", "foo", "qux"] // <-- Helpful additional generated type
```

These kinds of opinions can be added with:
```golang
ts.ApplyMutations(
	config.EnumLists,
)
output, _ := ts.Serialize()
```

# Helpful notes

An incredible website to visualize the AST of typescript: https://ts-ast-viewer.com/