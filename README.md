# GoTs

`GoTs` is a go tool to convert golang types to typescript for enabling a consistent type definition across the frontend and backend. 

# How it works

`GoTs` first parses a set of golang packages. The Go AST is traversed to find all the types defined in the packages. 

These types are placed into a simple AST that directly maps to the typescript AST.

Using [goja](https://github.com/dop251/goja), these types are then converted to typescript using the typescript compiler API. 

