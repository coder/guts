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
}

func (*EnumMember) isNode()           {}
func (*EnumMember) isExpressionType() {}
