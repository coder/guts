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
	case *bindings.TupleType:
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
	case *bindings.Enum:
		walkList(v, n.Members)
	case *bindings.VariableStatement:
		Walk(v, n.Declarations)
	case *bindings.VariableDeclarationList:
		walkList(v, n.Declarations)
	case *bindings.VariableDeclaration:
		Walk(v, n.Type)
		Walk(v, n.Initializer)
	case *bindings.ReferenceType:
		walkList(v, n.Arguments)
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
	case *bindings.EnumMember:
		Walk(v, n.Value)
	case *bindings.TypeLiteralNode:
		walkList(v, n.Members)
	case *bindings.TypeIntersection:
		walkList(v, n.Types)
	case *bindings.ExpressionWithTypeArguments:
		Walk(v, n.Expression)
		walkList(v, n.Arguments)
	case *bindings.ImportDeclaration:
		walkList(v, n.Named)
	case *bindings.ImportSpecifier:
		// noop
	case *bindings.IdentifierExpression:
		// noop
	case *bindings.PropertyAccessExpression:
		Walk(v, n.Expression)
	case *bindings.CallExpression:
		Walk(v, n.Expression)
		walkList(v, n.Arguments)
	case *bindings.ObjectLiteralExpression:
		walkList(v, n.Properties)
	case *bindings.PropertyAssignment:
		Walk(v, n.Initializer)
	case *bindings.Parameter:
		Walk(v, n.Type)
	case *bindings.ArrowFunction:
		walkList(v, n.Parameters)
		Walk(v, n.ReturnType)
		Walk(v, n.Body)
	case *bindings.TypeQuery:
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
