package bindings

import (
	"fmt"

	"github.com/dop251/goja"
)

type Node interface {
	isNode()
}
type isTypescriptNode struct{}

func (isTypescriptNode) isNode() {}

type Identifier string

type Source struct {
	File string
}

type HeritageType string

const (
	HeritageTypeExtends    HeritageType = "extends"
	HeritageTypeImplements HeritageType = "implements"
)

type HeritageClause struct {
	Token HeritageType
	Args  []ExpressionType
	isTypescriptNode
}

func HeritageClauseExtends(args ...ExpressionType) *HeritageClause {
	return &HeritageClause{
		Token: HeritageTypeExtends,
		Args:  args,
	}
}

func (s Source) Comment(n *goja.Object) Comment {
	return Comment{
		SingleLine:      true,
		Text:            fmt.Sprintf("From %s", s.File),
		TrailingNewLine: false,
		Node:            n,
	}
}

type Comment struct {
	// Single or multi-line comment
	SingleLine      bool
	Text            string
	TrailingNewLine bool

	Node *goja.Object
	isTypescriptNode
}

type Modifier string

const (
	ModifierAbstract  = "AbstractKeyword"
	ModifierAccessor  = "AccessorKeyword"
	ModifierAsync     = "AsyncKeyword"
	ModifierConst     = "ConstKeyword"
	ModifierDeclare   = "DeclareKeyword"
	ModifierDefault   = "DefaultKeyword"
	ModifierExport    = "ExportKeyword"
	ModifierIn        = "InKeyword"
	ModifierPrivate   = "PrivateKeyword"
	ModifierProtected = "ProtectedKeyword"
	ModifierPublic    = "PublicKeyword"
	ModifierReadonly  = "ReadonlyKeyword"
	ModifierOut       = "OutKeyword"
	ModifierOverride  = "OverrideKeyword"
	ModifierStatic    = "StaticKeyword"
)

type NodeFlags int

const (
	NodeFlagsConstant = 2
)
