package bindings

import (
	"fmt"
	"go/token"
	"go/types"

	"github.com/dop251/goja"
)

type Node interface {
	isNode()
}

// Identifier is a name given to a variable, function, class, etc.
// Identifiers should be unique within a package. Package information is
// included to help with disambiguation in the case of name collisions.
type Identifier struct {
	Name    string
	Package *types.Package
	Prefix  string
}

// GoName should be a unique name for the identifier across all Go packages.
func (i Identifier) GoName() string {
	if i.PkgName() != "" {
		return fmt.Sprintf("%s.%s", i.PkgName(), i.Name)
	}
	return i.Name
}

func (i Identifier) PkgName() string {
	if i.Package == nil {
		return ""
	}
	return i.Package.Path()
}

func (i Identifier) String() string {
	return i.Name
}

// Ref returns the identifier reference to be used in the generated code.
// This is the identifier to be used in typescript, since all generated code
// lands in the same namespace.
func (i Identifier) Ref() string {
	return i.Prefix + i.Name
}

// Source is the golang file that an entity is sourced from.
type Source struct {
	File     string
	Position token.Position
}

type HeritageType string

const (
	HeritageTypeExtends    HeritageType = "extends"
	HeritageTypeImplements HeritageType = "implements"
)

// HeritageClause
// interface Foo extends Bar, Baz {}
type HeritageClause struct {
	Token HeritageType
	Args  []ExpressionType
}

func (h *HeritageClause) isNode() {}

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
}

func (c *Comment) isNode() {}

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
