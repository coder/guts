package bindings

import (
	"fmt"

	"github.com/dop251/goja"
	"golang.org/x/xerrors"
)

func (b *Bindings) SerializeToTypescript(node *goja.Object) (string, error) {
	toTypeScriptF, err := b.f("toTypescript")
	if err != nil {
		return "", err
	}

	res, err := toTypeScriptF(goja.Undefined(), node)
	if err != nil {
		return "", xerrors.Errorf("call printNode: %w", err)
	}

	return res.String(), nil
}

func (b *Bindings) ToTypescriptNode(ety Node) (*goja.Object, error) {
	var siObj *goja.Object
	var err error

	switch node := ety.(type) {
	case *HeritageClause:
		siObj, err = b.HeritageClause(node)
	case *PropertySignature:
		siObj, err = b.PropertySignature(node)
	case *TypeParameter:
		siObj, err = b.TypeParameter(node)
	case DeclarationType:
		// Defer to the ExpressionType implementation
		siObj, err = b.ToTypescriptDeclarationNode(node)
	case ExpressionType:
		// Defer to the ExpressionType implementation
		siObj, err = b.ToTypescriptExpressionNode(node)
	default:
		return nil, fmt.Errorf("unsupported node type for typescript serialization: %T", node)
	}

	if err != nil {
		return nil, err
	}

	// Append any source comments
	if hasSource, ok := ety.(HasSource); ok {
		cmt, set := hasSource.SourceComment()
		if set {
			siObj, err = b.CommentGojaObject([]SyntheticComment{cmt}, siObj)
			if err != nil {
				return nil, xerrors.Errorf("source comment declaration: %w", err)
			}
		}
	}

	// Append any other comments
	if commented, ok := ety.(Commentable); ok {
		siObj, err = b.CommentGojaObject(commented.Comments(), siObj)
		if err != nil {
			return nil, xerrors.Errorf("comment declaration: %w", err)
		}
	}

	return siObj, nil
}

func (b *Bindings) ToTypescriptDeclarationNode(ety DeclarationType) (*goja.Object, error) {
	var siObj *goja.Object
	var err error

	switch ety := ety.(type) {
	case *Interface:
		siObj, err = b.Interface(ety)
	case *Alias:
		siObj, err = b.Alias(ety)
	case *VariableStatement:
		siObj, err = b.VariableStatement(ety)
	case *Enum:
		siObj, err = b.EnumDeclaration(ety)
	default:
		return nil, xerrors.Errorf("unsupported type for declaration type: %T", ety)
	}

	return siObj, err
}

func (b *Bindings) ToTypescriptExpressionNode(ety ExpressionType) (*goja.Object, error) {
	var siObj *goja.Object
	var err error

	switch ety := ety.(type) {
	case *LiteralKeyword:
		siObj, err = b.LiteralKeyword(ety)
	case *ReferenceType:
		siObj, err = b.Reference(ety)
	case *TupleType:
		siObj, err = b.Tuple(ety.Length, ety.Node)
	case *ArrayType:
		siObj, err = b.Array(ety.Node)
	case *UnionType:
		siObj, err = b.Union(ety)
	case *EnumMember:
		siObj, err = b.EnumMember(ety)
	case *Null:
		siObj, err = b.Null()
	case *VariableDeclarationList:
		siObj, err = b.VariableDeclarationList(ety)
	case *VariableDeclaration:
		siObj, err = b.VariableDeclaration(ety)
	case *LiteralType:
		switch v := ety.Value.(type) {
		case string:
			siObj, err = b.StringLiteral(v)
		case int64:
			siObj, err = b.NumericLiteral(v)
		case float64:
			siObj, err = b.FloatLiteral(v)
		case bool:
			siObj, err = b.BooleanLiteral(0)
		default:
			return nil, xerrors.Errorf("unsupported literal type: %T", ety.Value)
		}
	case *ArrayLiteralType:
		siObj, err = b.ArrayLiteral(ety)
	case *OperatorNodeType:
		siObj, err = b.OperatorNode(ety)
	case *TypeLiteralNode:
		siObj, err = b.TypeLiteralNode(ety)
	case *TypeIntersection:
		siObj, err = b.TypeIntersection(ety)
	default:
		return nil, xerrors.Errorf("unsupported type for field type: %T", ety)
	}

	return siObj, err
}

func (b *Bindings) Identifier(name string) (*goja.Object, error) {
	modifier, err := b.f("identifier")
	if err != nil {
		return nil, err
	}

	res, err := modifier(goja.Undefined(), b.vm.ToValue(name))
	if err != nil {
		panic(err)
	}

	return res.ToObject(b.vm), nil
}

func (b *Bindings) Reference(ref *ReferenceType) (*goja.Object, error) {
	modifier, err := b.f("reference")
	if err != nil {
		return nil, err
	}

	var args []interface{}
	for _, arg := range ref.Arguments {
		v, err := b.ToTypescriptNode(arg)
		if err != nil {
			return nil, fmt.Errorf("reference argument: %w", err)
		}
		args = append(args, v)
	}

	res, err := modifier(goja.Undefined(),
		b.vm.ToValue(ref.Name.Ref()),
		b.vm.NewArray(args...),
	)
	if err != nil {
		panic(err)
	}

	return res.ToObject(b.vm), nil
}

func (b *Bindings) PropertySignature(sig *PropertySignature) (*goja.Object, error) {
	propertySignature, err := b.f("propertySignature")
	if err != nil {
		return nil, err
	}

	siObj, err := b.ToTypescriptNode(sig.Type)
	if err != nil {
		return nil, fmt.Errorf("property field type: %w", err)
	}

	res, err := propertySignature(goja.Undefined(),
		b.vm.ToValue(ToStrings(sig.Modifiers)),
		b.vm.ToValue(sig.Name),
		b.vm.ToValue(sig.QuestionToken),
		siObj,
	)
	if err != nil {
		return nil, xerrors.Errorf("call propertySignature: %w", err)
	}

	return res.ToObject(b.vm), nil
}

func (b *Bindings) LiteralKeyword(word *LiteralKeyword) (*goja.Object, error) {
	literalKeyword, err := b.f("literalKeyword")
	if err != nil {
		return nil, err
	}

	res, err := literalKeyword(goja.Undefined(), b.vm.ToValue(word))
	if err != nil {
		return nil, xerrors.Errorf("call literalKeyword: %w", err)
	}
	return res.ToObject(b.vm), nil
}

func (b *Bindings) Interface(ti *Interface) (*goja.Object, error) {
	interfaceDecl, err := b.f("interfaceDecl")
	if err != nil {
		return nil, err
	}

	var fields []interface{}
	for _, field := range ti.Fields {
		v, err := b.ToTypescriptNode(field)
		if err != nil {
			return nil, err
		}

		fields = append(fields, v)
	}

	var typeParams []interface{}
	for _, tp := range ti.Parameters {
		v, err := b.ToTypescriptNode(tp)
		if err != nil {
			return nil, err
		}
		typeParams = append(typeParams, v)
	}

	var heritage []interface{}
	for _, h := range ti.Heritage {
		v, err := b.ToTypescriptNode(h)
		if err != nil {
			return nil, err
		}
		heritage = append(heritage, v)
	}

	res, err := interfaceDecl(goja.Undefined(),
		b.vm.ToValue(ToStrings(ti.Modifiers)),
		b.vm.ToValue(ti.Name.Ref()),
		b.vm.NewArray(typeParams...),
		b.vm.NewArray(heritage...),
		b.vm.NewArray(fields...),
	)
	if err != nil {
		return nil, xerrors.Errorf("call interfaceDecl: %w", err)
	}

	obj := res.ToObject(b.vm)

	return obj, nil
}

func (b *Bindings) HeritageClause(clause *HeritageClause) (*goja.Object, error) {
	clauseF, err := b.f("heritageClause")
	if err != nil {
		return nil, err
	}

	var args []interface{}
	for _, arg := range clause.Args {
		v, err := b.ToTypescriptExpressionNode(arg)
		if err != nil {
			return nil, fmt.Errorf("heritage clause argument: %w", err)
		}
		args = append(args, v)
	}

	res, err := clauseF(goja.Undefined(),
		b.vm.ToValue(string(clause.Token)),
		b.vm.NewArray(args...),
	)
	if err != nil {
		return nil, xerrors.Errorf("call heritageClause: %w", err)
	}
	return res.ToObject(b.vm), nil
}

func (b *Bindings) Modifier(name Modifier) (*goja.Object, error) {
	modifier, err := b.f("modifier")
	if err != nil {
		return nil, err
	}

	res, err := modifier(goja.Undefined(), b.vm.ToValue(name))
	if err != nil {
		panic(err)
	}

	return res.ToObject(b.vm), nil
}

func (b *Bindings) Array(arrType ExpressionType) (*goja.Object, error) {
	array, err := b.f("arrayType")
	if err != nil {
		return nil, err
	}

	siObj, err := b.ToTypescriptNode(arrType)
	if err != nil {
		return nil, fmt.Errorf("array type: %w", err)
	}

	res, err := array(goja.Undefined(), siObj)
	if err != nil {
		return nil, xerrors.Errorf("call arrayType: %w", err)
	}
	return res.ToObject(b.vm), nil
}

func (b *Bindings) Tuple(length int, tupleType ExpressionType) (*goja.Object, error) {
	tuple, err := b.f("homogeneousTupleType")
	if err != nil {
		return nil, err
	}

	siObj, err := b.ToTypescriptNode(tupleType)
	if err != nil {
		return nil, fmt.Errorf("array type: %w", err)
	}

	res, err := tuple(goja.Undefined(), b.vm.ToValue(length), siObj)
	if err != nil {
		return nil, xerrors.Errorf("call arrayType: %w", err)
	}
	return res.ToObject(b.vm), nil
}

func (b *Bindings) Alias(alias *Alias) (*goja.Object, error) {
	aliasFunc, err := b.f("aliasDecl")
	if err != nil {
		return nil, err
	}

	siObj, err := b.ToTypescriptNode(alias.Type)
	if err != nil {
		return nil, fmt.Errorf("alias type: %w", err)
	}

	var typeParams []interface{}
	for _, tp := range alias.Parameters {
		v, err := b.ToTypescriptNode(tp)
		if err != nil {
			return nil, err
		}
		typeParams = append(typeParams, v)
	}

	res, err := aliasFunc(goja.Undefined(),
		b.vm.ToValue(ToStrings(alias.Modifiers)),
		b.vm.ToValue(alias.Name.Ref()),
		b.vm.NewArray(typeParams...),
		siObj,
	)
	if err != nil {
		return nil, xerrors.Errorf("call aliasDecl: %w", err)
	}

	obj := res.ToObject(b.vm)

	return obj, nil
}

func (b *Bindings) TypeParameter(ty *TypeParameter) (*goja.Object, error) {
	typeParamF, err := b.f("typeParameterDeclaration")
	if err != nil {
		return nil, err
	}

	paramType := goja.Undefined()
	if ty.Type != nil {
		paramType, err = b.ToTypescriptExpressionNode(ty.Type)
		if err != nil {
			return nil, fmt.Errorf("type parameter type: %w", err)
		}
	}

	defaultType := goja.Undefined()
	if ty.DefaultType != nil {
		defaultType, err = b.ToTypescriptExpressionNode(ty.DefaultType)
		if err != nil {
			return nil, fmt.Errorf("type parameter default type: %w", err)
		}
	}

	res, err := typeParamF(goja.Undefined(),
		b.vm.ToValue(ToStrings(ty.Modifiers)),
		b.vm.ToValue(ty.Name.Ref()),
		paramType,
		defaultType,
	)
	if err != nil {
		return nil, xerrors.Errorf("call typeParameter: %w", err)
	}

	return res.ToObject(b.vm), nil
}

func (b *Bindings) Union(ty *UnionType) (*goja.Object, error) {
	unionF, err := b.f("unionType")
	if err != nil {
		return nil, err
	}

	var types []any
	for _, t := range ty.Types {
		v, err := b.ToTypescriptNode(t)
		if err != nil {
			return nil, fmt.Errorf("union type: %w", err)
		}
		types = append(types, v)
	}

	res, err := unionF(goja.Undefined(), b.vm.NewArray(types...))
	if err != nil {
		return nil, xerrors.Errorf("call unionType: %w", err)
	}

	return res.ToObject(b.vm), nil
}

func (b *Bindings) Null() (*goja.Object, error) {
	nullF, err := b.f("createNull")
	if err != nil {
		return nil, err
	}

	res, err := nullF(goja.Undefined())
	if err != nil {
		return nil, xerrors.Errorf("call nullType: %w", err)
	}

	return res.ToObject(b.vm), nil
}

func (b *Bindings) StringLiteral(value string) (*goja.Object, error) {
	literalF, err := b.f("stringLiteral")
	if err != nil {
		return nil, err
	}

	res, err := literalF(goja.Undefined(), b.vm.ToValue(value))
	if err != nil {
		return nil, xerrors.Errorf("call stringLiteral: %w", err)
	}
	return res.ToObject(b.vm), nil
}

func (b *Bindings) FloatLiteral(value float64) (*goja.Object, error) {
	literalF, err := b.f("numericLiteral")
	if err != nil {
		return nil, err
	}

	res, err := literalF(goja.Undefined(), b.vm.ToValue(value))
	if err != nil {
		return nil, xerrors.Errorf("call numericLiteral: %w", err)
	}
	return res.ToObject(b.vm), nil
}

func (b *Bindings) NumericLiteral(value int64) (*goja.Object, error) {
	literalF, err := b.f("numericLiteral")
	if err != nil {
		return nil, err
	}

	res, err := literalF(goja.Undefined(), b.vm.ToValue(value))
	if err != nil {
		return nil, xerrors.Errorf("call numericLiteral: %w", err)
	}
	return res.ToObject(b.vm), nil
}

func (b *Bindings) BooleanLiteral(value int) (*goja.Object, error) {
	literalF, err := b.f("numericLiteral")
	if err != nil {
		return nil, err
	}

	res, err := literalF(goja.Undefined(), b.vm.ToValue(value))
	if err != nil {
		return nil, xerrors.Errorf("call numericLiteral: %w", err)
	}
	return res.ToObject(b.vm), nil
}

func (b *Bindings) ArrayLiteral(value *ArrayLiteralType) (*goja.Object, error) {
	literalF, err := b.f("arrayLiteral")
	if err != nil {
		return nil, err
	}

	var elements []interface{}
	for _, elem := range value.Elements {
		v, err := b.ToTypescriptNode(elem)
		if err != nil {
			return nil, fmt.Errorf("array literal element: %w", err)
		}
		elements = append(elements, v)
	}

	res, err := literalF(goja.Undefined(), b.vm.NewArray(elements...))
	if err != nil {
		return nil, xerrors.Errorf("call numericLiteral: %w", err)
	}
	return res.ToObject(b.vm), nil
}

func (b *Bindings) VariableStatement(stmt *VariableStatement) (*goja.Object, error) {
	aliasFunc, err := b.f("variableStatement")
	if err != nil {
		return nil, err
	}

	siObj, err := b.ToTypescriptNode(stmt.Declarations)
	if err != nil {
		return nil, fmt.Errorf("alias type: %w", err)
	}

	res, err := aliasFunc(goja.Undefined(),
		b.vm.ToValue(ToStrings(stmt.Modifiers)),
		siObj,
	)
	if err != nil {
		return nil, xerrors.Errorf("call aliasDecl: %w", err)
	}

	obj := res.ToObject(b.vm)

	return obj, nil
}

func (b *Bindings) VariableDeclarationList(list *VariableDeclarationList) (*goja.Object, error) {
	aliasFunc, err := b.f("variableDeclarationList")
	if err != nil {
		return nil, err
	}

	var decls []interface{}
	for _, decl := range list.Declarations {
		v, err := b.ToTypescriptNode(decl)
		if err != nil {
			return nil, err
		}
		decls = append(decls, v)
	}

	res, err := aliasFunc(
		goja.Undefined(),
		b.vm.NewArray(decls...),
		b.vm.ToValue(list.Flags),
	)
	if err != nil {
		return nil, xerrors.Errorf("call variableDeclarationList: %w", err)
	}

	return res.ToObject(b.vm), nil
}

func (b *Bindings) VariableDeclaration(decl *VariableDeclaration) (*goja.Object, error) {
	aliasFunc, err := b.f("variableDeclaration")
	if err != nil {
		return nil, err
	}

	var declType goja.Value = goja.Undefined()
	if decl.Type != nil {
		declType, err = b.ToTypescriptNode(decl.Type)
		if err != nil {
			return nil, fmt.Errorf("alias type: %w", err)
		}
	}

	var declInit goja.Value = goja.Undefined()
	if decl.Initializer != nil {
		declInit, err = b.ToTypescriptNode(decl.Initializer)
		if err != nil {
			return nil, fmt.Errorf("alias type: %w", err)
		}
	}

	res, err := aliasFunc(
		goja.Undefined(),
		b.vm.ToValue(decl.Name.Ref()),
		b.vm.ToValue(decl.ExclamationMark),
		declType,
		declInit,
	)
	if err != nil {
		return nil, xerrors.Errorf("call variableDeclaration: %w", err)
	}

	return res.ToObject(b.vm), nil
}

func (b *Bindings) OperatorNode(value *OperatorNodeType) (*goja.Object, error) {
	literalF, err := b.f("typeOperatorNode")
	if err != nil {
		return nil, err
	}

	obj, err := b.ToTypescriptNode(value.Type)
	if err != nil {
		return nil, fmt.Errorf("operator type: %w", err)
	}

	res, err := literalF(goja.Undefined(), b.vm.ToValue(value.Keyword), obj)
	if err != nil {
		return nil, xerrors.Errorf("call numericLiteral: %w", err)
	}
	return res.ToObject(b.vm), nil
}

func (b *Bindings) EnumMember(value *EnumMember) (*goja.Object, error) {
	literalF, err := b.f("enumMember")
	if err != nil {
		return nil, err
	}

	obj := goja.Undefined()
	if value.Value != nil {
		obj, err = b.ToTypescriptNode(value.Value)
		if err != nil {
			return nil, fmt.Errorf("enum member type: %w", err)
		}
	}

	res, err := literalF(goja.Undefined(), b.vm.ToValue(value.Name), obj)
	if err != nil {
		return nil, xerrors.Errorf("call enumMember: %w", err)
	}
	return res.ToObject(b.vm), nil
}

func (b *Bindings) EnumDeclaration(e *Enum) (*goja.Object, error) {
	aliasFunc, err := b.f("enumDeclaration")
	if err != nil {
		return nil, err
	}

	var members []any
	for _, m := range e.Members {
		v, err := b.ToTypescriptNode(m)
		if err != nil {
			return nil, fmt.Errorf("enum type: %w", err)
		}
		members = append(members, v)
	}

	res, err := aliasFunc(goja.Undefined(),
		b.vm.ToValue(ToStrings(e.Modifiers)),
		b.vm.ToValue(e.Name.Ref()),
		b.vm.NewArray(members...),
	)
	if err != nil {
		return nil, xerrors.Errorf("call enumDeclaration: %w", err)
	}

	obj := res.ToObject(b.vm)

	return obj, nil
}

func (b *Bindings) TypeLiteralNode(node *TypeLiteralNode) (*goja.Object, error) {
	typeLiteralF, err := b.f("typeLiteralNode")
	if err != nil {
		return nil, err
	}

	var members []interface{}
	for _, member := range node.Members {
		v, err := b.ToTypescriptNode(member)
		if err != nil {
			return nil, err
		}

		members = append(members, v)
	}

	res, err := typeLiteralF(goja.Undefined(), b.vm.NewArray(members...))
	if err != nil {
		return nil, xerrors.Errorf("call typeLiteralNode: %w", err)
	}

	return res.ToObject(b.vm), nil
}

func (b *Bindings) TypeIntersection(node *TypeIntersection) (*goja.Object, error) {
	intersectionF, err := b.f("intersectionType")
	if err != nil {
		return nil, err
	}

	var types []interface{}
	for _, t := range node.Types {
		v, err := b.ToTypescriptNode(t)
		if err != nil {
			return nil, fmt.Errorf("intersection type: %w", err)
		}
		types = append(types, v)
	}

	res, err := intersectionF(goja.Undefined(), b.vm.NewArray(types...))
	if err != nil {
		return nil, xerrors.Errorf("call intersectionType: %w", err)
	}
	return res.ToObject(b.vm), nil
}

func (b *Bindings) CommentGojaObject(comments []SyntheticComment, object *goja.Object) (*goja.Object, error) {
	if len(comments) == 0 {
		return object, nil
	}

	commentF, err := b.f("addSyntheticComment")
	if err != nil {
		return nil, err
	}

	node := object
	for _, c := range comments {
		res, err := commentF(goja.Undefined(),
			node,
			b.vm.ToValue(c.Leading),
			b.vm.ToValue(c.SingleLine),
			b.vm.ToValue(" "+c.Text),
			b.vm.ToValue(c.TrailingNewLine),
		)
		if err != nil {
			return nil, xerrors.Errorf("call addSyntheticComment: %w", err)
		}
		node = res.ToObject(b.vm)
	}

	return node, nil
}
