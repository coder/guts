package bindings

import "fmt"

// These string functions are purely for debugging

func (i Identifier) String() string        { return string(i) }
func (a Alias) String() string             { return fmt.Sprintf("Alias:%s", a.Name) }
func (k LiteralKeyword) String() string    { return string(k) }
func (a ArrayType) String() string         { return fmt.Sprintf("[]%s", a.Node) }
func (i Interface) String() string         { return fmt.Sprintf("Interface:%s", i.Name) }
func (f PropertySignature) String() string { return fmt.Sprintf("PropertySignature:%s", f.Name) }
func (r ReferenceType) String() string     { return fmt.Sprintf("Reference:%s", r.Name) }
func (r UnionType) String() string         { return "Union" }
