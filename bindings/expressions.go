package bindings

import (
	"fmt"

	"golang.org/x/xerrors"
)

// ExpressionType
type ExpressionType interface {
	isExpressionType()
	Node
}

type LiteralKeyword string

// LiteralKeyword is a pointer to be consistent with the others
func (*LiteralKeyword) isExpressionType() {}
func (*LiteralKeyword) isNode()           {}

const (
	KeywordVoid      LiteralKeyword = "VoidKeyword"
	KeywordAny       LiteralKeyword = "AnyKeyword"
	KeywordBoolean   LiteralKeyword = "BooleanKeyword"
	KeywordIntrinsic LiteralKeyword = "IntrinsicKeyword"
	KeywordNever     LiteralKeyword = "NeverKeyword"
	KeywordNumber    LiteralKeyword = "NumberKeyword"
	KeywordObject    LiteralKeyword = "ObjectKeyword"
	KeywordString    LiteralKeyword = "StringKeyword"
	KeywordSymbol    LiteralKeyword = "SymbolKeyword"
	KeywordUndefined LiteralKeyword = "UndefinedKeyword"
	KeywordUnknown   LiteralKeyword = "UnknownKeyword"
	KeywordBigInt    LiteralKeyword = "BigIntKeyword"
	KeywordReadonly  LiteralKeyword = "ReadonlyKeyword"
	KeywordUnique    LiteralKeyword = "UniqueKeyword"
	KeywordKeyOf     LiteralKeyword = "KeyOfKeyword"
)

func ToTypescriptLiteralKeyword(word string) (LiteralKeyword, error) {
	switch word {
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "float32", "float64":
		return KeywordNumber, nil
	case "string":
		return KeywordString, nil
	case "bool":
		return KeywordBoolean, nil
	default:
		return KeywordAny, xerrors.Errorf("unsupported literal keyword: %s", word)
	}
}

type LiteralType struct {
	Value any // should be some constant value
}

func (*LiteralType) isNode()           {}
func (*LiteralType) isExpressionType() {}

// ReferenceType can be used to reference another type by name
type ReferenceType struct {
	Name Identifier `json:"name"`
	// TODO: Generics
	Arguments []ExpressionType `json:"arguments"`
}

func Reference(name Identifier, args ...ExpressionType) *ReferenceType {
	return &ReferenceType{Name: name, Arguments: args}
}

func (*ReferenceType) isNode()           {}
func (*ReferenceType) isExpressionType() {}

type TupleType struct {
	// TODO: Technically tuples can be heterogeneous, but golang does not really
	// support that. So just assume that all elements are the same type.
	Node   ExpressionType
	Length int
}

func (*TupleType) isNode()           {}
func (*TupleType) isExpressionType() {}

func HomogeneousTuple(length int, node ExpressionType) *TupleType {
	return &TupleType{
		Node:   node,
		Length: length,
	}
}

type ArrayType struct {
	Node ExpressionType
}

func (*ArrayType) isNode()           {}
func (*ArrayType) isExpressionType() {}

func Array(node ExpressionType) *ArrayType {
	return &ArrayType{
		Node: node,
	}
}

type ArrayLiteralType struct {
	Elements []ExpressionType
}

func (*ArrayLiteralType) isNode()           {}
func (*ArrayLiteralType) isExpressionType() {}

type UnionType struct {
	Types []ExpressionType
}

func (*UnionType) isNode()           {}
func (*UnionType) isExpressionType() {}

func Union(types ...ExpressionType) *UnionType {
	return &UnionType{Types: types}
}

type Null struct {
}

func (*Null) isNode()           {}
func (*Null) isExpressionType() {}

type ExpressionWithTypeArguments struct {
	Expression ExpressionType
	Arguments  []ExpressionType
}

func (*ExpressionWithTypeArguments) isNode()           {}
func (*ExpressionWithTypeArguments) isExpressionType() {}

type VariableDeclarationList struct {
	Declarations []*VariableDeclaration
	Flags        NodeFlags
}

func (*VariableDeclarationList) isNode()           {}
func (*VariableDeclarationList) isExpressionType() {}

type VariableDeclaration struct {
	Name            Identifier
	ExclamationMark bool
	Type            ExpressionType
	Initializer     ExpressionType
}

func (*VariableDeclaration) isNode()           {}
func (*VariableDeclaration) isExpressionType() {}

type OperatorNodeType struct {
	Keyword LiteralKeyword
	Type    ExpressionType
}

// OperatorNode allows adding a keyword to a type
// Keyword must be "KeyOfKeyword" | "UniqueKeyword" | "ReadonlyKeyword"
func OperatorNode(keyword LiteralKeyword, node ExpressionType) *OperatorNodeType {
	switch keyword {
	case KeywordReadonly, KeywordUnique, KeywordKeyOf:
	default:
		// TODO: Would be better to raise some error here.
		panic(fmt.Sprint("unsupported operator keyword: ", keyword))
	}
	return &OperatorNodeType{
		Keyword: keyword,
		Type:    node,
	}
}

func (*OperatorNodeType) isNode()           {}
func (*OperatorNodeType) isExpressionType() {}

type EnumMember struct {
	Name string
	// Value is allowed to be nil, which results in `undefined`.
	Value ExpressionType
	SupportComments
}

func (*EnumMember) isNode()           {}
func (*EnumMember) isExpressionType() {}

// TypeLiteralNode represents an object type literal like { name: string }
type TypeLiteralNode struct {
	Members []*PropertySignature
}

func (*TypeLiteralNode) isNode()           {}
func (*TypeLiteralNode) isExpressionType() {}

type TypeIntersection struct {
	Types []ExpressionType
}

func (*TypeIntersection) isNode()           {}
func (*TypeIntersection) isExpressionType() {}

// IdentifierExpression is a value-position TypeScript identifier such as
// the `z` in `z.string()` or `BaseSchema` in `BaseSchema.extend({...})`.
//
// It is distinct from bindings.Identifier despite the similar name.
// Identifier is parser-layer plumbing: a qualified-name handle
// (Name + Package + Prefix) used to resolve and disambiguate references
// across Go packages, and it is not itself a Node. IdentifierExpression is
// tree-layer plumbing: a Node that implements ExpressionType so callers
// can place a value-position identifier inside expression slots like
// CallExpression.Expression.
//
// The Name field is itself an Identifier so cross-package prefixing flows
// through .Ref() the same way it does for ReferenceType.Name and
// VariableDeclaration.Name.
type IdentifierExpression struct {
	Name Identifier
}

func (*IdentifierExpression) isNode()           {}
func (*IdentifierExpression) isExpressionType() {}

// PropertyAccessExpression is `<expression>.<name>`, used to chain method
// names or member references such as `z.string` or `BaseSchema.extend`.
type PropertyAccessExpression struct {
	Expression ExpressionType
	Name       string
}

func (*PropertyAccessExpression) isNode()           {}
func (*PropertyAccessExpression) isExpressionType() {}

// CallExpression is `<expression>(args...)`. It composes with
// PropertyAccessExpression to build chained calls like
// `z.string().optional()`.
type CallExpression struct {
	Expression ExpressionType
	Arguments  []ExpressionType
}

func (*CallExpression) isNode()           {}
func (*CallExpression) isExpressionType() {}

// ObjectLiteralExpression is `{ k: v, ... }` in expression position. It is
// distinct from TypeLiteralNode, which emits a TypeScript object type.
//
// Only the PropertyAssignment form is modeled. ShorthandPropertyAssignment
// (`{ x }`), SpreadAssignment (`{ ...rest }`), MethodDeclaration, and
// accessor properties are not yet supported; add them if you need them.
type ObjectLiteralExpression struct {
	Properties []*PropertyAssignment
}

func (*ObjectLiteralExpression) isNode()           {}
func (*ObjectLiteralExpression) isExpressionType() {}

// PropertyAssignment is `<name>: <initializer>` inside an
// ObjectLiteralExpression. It is a node but not an ExpressionType or a
// DeclarationType because it only appears as a child of
// ObjectLiteralExpression.
type PropertyAssignment struct {
	Name        string
	Initializer ExpressionType
}

func (*PropertyAssignment) isNode() {}

// Parameter is a single parameter in an ArrowFunction signature. Name is
// required; Type may be nil to omit the annotation.
type Parameter struct {
	Name string
	Type ExpressionType
}

func (*Parameter) isNode() {}

// ArrowFunction is `(parameters): returnType => body`. ReturnType may be
// nil to omit the annotation. Body is currently required to be a single
// expression; statement bodies are not yet modeled.
type ArrowFunction struct {
	Parameters []*Parameter
	ReturnType ExpressionType
	Body       ExpressionType
}

func (*ArrowFunction) isNode()           {}
func (*ArrowFunction) isExpressionType() {}

// TypeQuery is `typeof <name>`. It appears in type position, typically as
// a generic argument such as the `typeof FooSchema` inside
// `z.infer<typeof FooSchema>`.
//
// Name is an Identifier so cross-package prefixing flows through .Ref(),
// matching the rest of the AST. Without this, a TypeQuery for a prefixed
// declaration would emit `typeof Foo` while the matching value-position
// IdentifierExpression emits `ExternalFoo`, and the two would not line up.
type TypeQuery struct {
	Name Identifier
}

func (*TypeQuery) isNode()           {}
func (*TypeQuery) isExpressionType() {}
