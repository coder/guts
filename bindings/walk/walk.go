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
	if node == nil {
		return
	}
	if v = v.Visit(node); v == nil {
		return
	}

	// Walk all node types
	// If there is a missing node, please add it.
	switch n := node.(type) {
	case *bindings.ArrayLiteralType:
		walkList(v, n.Elements)
	case *bindings.ArrayType:
		Walk(v, n.Node)
	case *bindings.Interface:
		walkList(v, n.Parameters)
		walkList(v, n.Heritage)
		walkList(v, n.Fields)
	case *bindings.PropertySignature:
		Walk(v, n.Type)
	case *bindings.Alias:
		Walk(v, n.Type)
	case *bindings.TypeParameter:
		Walk(v, n.Type)
	case *bindings.UnionType:
		walkList(v, n.Types)
	case *bindings.VariableStatement:
		Walk(v, n.Declarations)
	case *bindings.VariableDeclarationList:
		walkList(v, n.Declarations)
	case *bindings.VariableDeclaration:
		Walk(v, n.Type)
		Walk(v, n.Initializer)
	case *bindings.ReferenceType:
		// noop
	case *bindings.LiteralKeyword:
		// noop
	case *bindings.LiteralType:
		// noop
	case *bindings.Null:
	// noop
	case *bindings.HeritageClause:
		walkList(v, n.Args)
	case *bindings.OperatorNodeType:
		Walk(v, n.Type)
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
