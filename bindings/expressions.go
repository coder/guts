package bindings

import "golang.org/x/xerrors"

// ExpressionType
type ExpressionType interface {
	isExpressionType()
}

type LiteralKeyword string

func (LiteralKeyword) isExpressionType() {}

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

func (LiteralType) isExpressionType() {}

// ReferenceType can be used to reference another type by name
type ReferenceType struct {
	Name string `json:"name"`
	// TODO: Generics
	Arguments []ExpressionType `json:"arguments"`
}

func (ReferenceType) isExpressionType() {}

func Reference(name string, args ...ExpressionType) ReferenceType {
	return ReferenceType{Name: name, Arguments: args}
}

type ArrayType struct {
	Node ExpressionType
}

func Array(node ExpressionType) ArrayType {
	return ArrayType{
		Node: node,
	}
}

func (ArrayType) isExpressionType() {}

type UnionType struct {
	Types []ExpressionType
}

func Union(types ...ExpressionType) UnionType {
	return UnionType{Types: types}
}

func (UnionType) isExpressionType() {}

type Null struct{}

func (Null) isExpressionType() {}

type ExpressionWithTypeArguments struct {
	Expression ExpressionType
	Arguments  []ExpressionType
}

func (ExpressionWithTypeArguments) isExpressionType() {}

type VariableDeclarationList struct {
	Declarations []VariableDeclaration
	Flags        NodeFlags
}

func (VariableDeclarationList) isExpressionType() {}

type VariableDeclaration struct {
	Name            string
	ExclamationMark bool
	Type            ExpressionType
	Initializer     ExpressionType
}

func (VariableDeclaration) isExpressionType() {}
