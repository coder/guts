package bindings

import (
	"fmt"
	"strings"
)

// These string functions are purely for debugging

func (a Alias) String() string          { return fmt.Sprintf("Alias:%s", a.Name) }
func (k LiteralKeyword) String() string { return string(k) }
func (a ArrayLiteralType) String() string {
	strs := []string{}
	for _, e := range a.Elements {
		strs = append(strs, fmt.Sprintf("%s", e))
	}
	return fmt.Sprintf("{%s}", strings.Join(strs, ","))
}
func (a ArrayType) String() string           { return fmt.Sprintf("[]%s", a.Node) }
func (i Interface) String() string           { return fmt.Sprintf("Interface:%s", i.Name) }
func (f PropertySignature) String() string   { return fmt.Sprintf("PropertySignature:%s", f.Name) }
func (r ReferenceType) String() string       { return fmt.Sprintf("Reference:%s", r.Name) }
func (r UnionType) String() string           { return "Union" }
func (o OperatorNodeType) String() string    { return fmt.Sprintf("Operator:%s", o.Keyword) }
func (p TypeParameter) String() string       { return fmt.Sprintf("TypeParameter:%s", p.Name) }
func (v VariableDeclaration) String() string { return fmt.Sprintf("VariableDeclaration:%s", v.Name) }
