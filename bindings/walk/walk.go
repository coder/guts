package walk

import (
	"fmt"
	"strings"

	"github.com/coder/guts/bindings"
)

// Visitor mimics the golang ast visitor interface.
type Visitor interface {
	Visit(node bindings.Node) (w Visitor)
}

// Walk walks the Typescript tree in depth-first order.
// The node can be anything, would be nice to have some types.
func Walk(v Visitor, node bindings.Node) {
	if v = v.Visit(node); v == nil {
		return
	}

	// Walk all node types
	switch n := node.(type) {
	case *bindings.Interface:
		walkList(v, n.Fields)
	case *bindings.PropertySignature:
		Walk(v, n.Type)
	case *bindings.Alias:
		Walk(v, n.Type)
	case *bindings.TypeParameter:
		Walk(v, n.Type)
	case *bindings.UnionType:
		walkList(v, n.Types)
	case *bindings.ReferenceType:
		// noop
	case *bindings.LiteralKeyword:
		// noop
	default:
		panic(fmt.Sprintf("convert.Walk: unexpected node type %T", n))
	}
}

func walkList[N bindings.Node](v Visitor, list []N) {
	for _, node := range list {
		Walk(v, node)
	}
}

// PrintingVisitor prints the tree to stdout.
type PrintingVisitor int

func (p PrintingVisitor) Visit(node bindings.Node) (w Visitor) {
	spaces := 2 * int(p)
	fmt.Printf("%s%s\n", strings.Repeat(" ", spaces), node)
	return p + 1
}
