package convert

import (
	"context"
	"fmt"
	"go/constant"
	"go/token"
	"go/types"
	"log/slog"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/fatih/structtag"
	"golang.org/x/tools/go/packages"
	"golang.org/x/xerrors"

	"github.com/coder/gots/bindings"
)

// GoParser takes in Golang packages, and can convert them to the intermediate
// typescript representation. The intermediate representation is closely
// aligned with the typescript AST.
type GoParser struct {
	Pkgs     map[string]*packages.Package
	Generate map[string]bool

	// customMappings can override any field type with a custom type.
	// Eg: "time.Time" -> "string"
	customMappings map[string]bindings.ExpressionType
	config         *packages.Config
	fileSet        *token.FileSet
}

func NewGolangParser() (*GoParser, error) {
	fileSet := token.NewFileSet()
	config := &packages.Config{
		// Just accept the fact we need these flags for what we want. Feel free to add
		// more, it'll just increase the time it takes to parse.
		Mode: packages.NeedTypes | packages.NeedName | packages.NeedTypesInfo |
			packages.NeedTypesSizes | packages.NeedSyntax | packages.NeedDeps,
		Tests: false,
		Fset:  fileSet,
		//Dir:     "/home/steven/go/src/github.com/coder/gots",
		Context: context.Background(),
	}

	return &GoParser{
		fileSet:        fileSet,
		config:         config,
		Pkgs:           make(map[string]*packages.Package),
		Generate:       map[string]bool{},
		customMappings: map[string]bindings.ExpressionType{},
	}, nil
}

// IncludeCustomDeclaration is an advanced form of IncludeCustom.
func (p *GoParser) IncludeCustomDeclaration(mappings map[string]bindings.ExpressionType) error {
	for k, v := range mappings {
		p.customMappings[k] = v
	}
	return nil
}

// IncludeCustom only works for basic literal types and non-parameterized reference types.
func (p *GoParser) IncludeCustom(mappings map[string]string) error {
	for k, v := range mappings {
		switch v {
		case "string":
			p.customMappings[k] = bindings.KeywordString
		case "number":
			p.customMappings[k] = bindings.KeywordNumber
		case "boolean":
			p.customMappings[k] = bindings.KeywordBoolean
		case "any":
			p.customMappings[k] = bindings.KeywordAny
		case "unknown":
			p.customMappings[k] = bindings.KeywordUnknown
		default:
			// TODO: Verify these at all?
			p.customMappings[k] = bindings.ReferenceType{
				Name: v,
			}
		}
	}
	return nil
}

// Include parses a directory and adds the parsed package to the list of packages.
// If "generate" is set to true, it will also generate the typescript code
// for types in the package.
// Setting it to false just includes it as a reference package.
// TODO: Reference packages are optional, as deps are fetched now
func (p *GoParser) Include(directory string, generate bool) error {
	pkgs, err := packages.Load(p.config, directory)
	if err != nil {
		return fmt.Errorf("failed to parse directory %s: %w", directory, err)
	}

	for _, v := range pkgs {
		if _, ok := p.Pkgs[v.PkgPath]; ok {
			return fmt.Errorf("package %s already exists", v.PkgPath)
		}
		p.Pkgs[v.PkgPath] = v
		p.Generate[v.PkgPath] = generate
		if len(v.Errors) > 0 {
			for _, e := range v.Errors {
				slog.Error(fmt.Sprintf("parsing pkg %s", v.PkgPath), slog.String("error", e.Error()))
			}
		}
	}
	return nil
}

// ToTypescript translates the Go types into the intermediate typescript AST
func (p *GoParser) ToTypescript() (*Typescript, error) {
	typescript := &Typescript{
		typescriptNodes: make(map[string]TypescriptNode),
		parsed:          p,
	}

	err := typescript.parseGolangIdentifiers()
	if err != nil {
		return nil, err
	}

	return typescript, nil
}

type Typescript struct {
	// TypescriptNodes is a map of typescript nodes that are generated from the
	// parsed go code. All names should be unique. If non-unique names exist, that
	// means packages contain the same named types.
	typescriptNodes map[string]TypescriptNode
	parsed          *GoParser
}

func (ts *Typescript) parseGolangIdentifiers() error {
	// Look for comments that indicate to ignore a type for typescript generation.
	// Comment format to skip typescript generation: `@typescript-ignore <ignored_type>`
	ignoreRegex := regexp.MustCompile("@typescript-ignore[:]?(?P<ignored_types>.*)")

	for _, pkg := range ts.parsed.Pkgs {
		if !ts.parsed.Generate[pkg.PkgPath] {
			continue // Skip reference packages
		}

		skippedTypes := make(map[string]struct{})
		for _, file := range pkg.Syntax {
			for _, comment := range file.Comments {
				for _, line := range comment.List {
					text := line.Text
					matches := ignoreRegex.FindStringSubmatch(text)
					ignored := ignoreRegex.SubexpIndex("ignored_types")
					if len(matches) >= ignored && matches[ignored] != "" {
						arr := strings.Split(matches[ignored], ",")
						for _, s := range arr {
							skippedTypes[strings.TrimSpace(s)] = struct{}{}
						}
					}
				}
			}
		}

		allIdents := pkg.Types.Scope().Names()
		for _, ident := range allIdents {
			if _, ok := skippedTypes[ident]; ok {
				continue
			}
			obj := pkg.Types.Scope().Lookup(ident)
			err := ts.parse(obj)
			if err != nil {
				return fmt.Errorf("failed to parse object %q in %q: %w", ident, pkg.PkgPath, err)
			}
		}
	}
	return nil
}

func (ts *Typescript) SetNode(key string, node TypescriptNode) error {
	if _, ok := ts.typescriptNodes[key]; ok {
		return fmt.Errorf("node %q already exists", key)
	}
	ts.typescriptNodes[key] = node
	return nil
}

func (ts *Typescript) UpdateNode(key string, update func(n TypescriptNode) TypescriptNode) {
	v, ok := ts.typescriptNodes[key]
	if !ok {
		ts.typescriptNodes[key] = TypescriptNode{}
	}
	ts.typescriptNodes[key] = update(v)
}

func (ts *Typescript) Serialize(vm *bindings.Bindings) (string, error) {
	nodeList := make([]string, 0, len(ts.typescriptNodes))
	for k := range ts.typescriptNodes {
		nodeList = append(nodeList, k)
	}

	// Sort the nodes so that the output is deterministic.
	sort.Strings(nodeList)

	var str strings.Builder
	str.WriteString("// Code generated by 'gots'. DO NOT EDIT.\n\n")

	for k, v := range ts.typescriptNodes {
		text, err := v.Typescript(vm)
		if err != nil {
			return "", fmt.Errorf("node %q: %w", k, err)
		}
		str.WriteString(text + "\n" + "\n")
	}
	return str.String(), nil

}

func (ts *Typescript) parse(obj types.Object) error {
	if obj == nil || obj.Type() == nil {
		slog.Info("nil object")
		return nil
	}

	objectName := obj.Name() // Package names can collide!

	switch obj := obj.(type) {
	// All named types are type declarations
	case *types.TypeName:
		var rhs types.Type
		switch typedObj := obj.Type().(type) {
		case *types.Named:
			rhs = typedObj.Underlying()
		case *types.Alias:
			rhs = typedObj.Rhs().Underlying()
		default:
			return fmt.Errorf("not supported top level type %T for %q", obj.Type(), objectName)
		}

		switch underNamed := rhs.(type) {
		case *types.Struct:
			// type <Name> struct
			// Structs are obvious.
			node, err := ts.buildStruct(obj, underNamed)
			if err != nil {
				return xerrors.Errorf("generate %q: %w", objectName, err)
			}
			return ts.SetNode(objectName, TypescriptNode{
				Node: node,
			})
		case *types.Basic:
			// type <Name> string
			// These are enums. Store to expand later.
			rhs, err := ts.typescriptType(underNamed)
			if err != nil {
				return xerrors.Errorf("generate basic %q: %w", objectName, err)
			}

			// If this has 'consts', then it is an enum. The enum code will
			// patch this value to be more specific.
			ts.UpdateNode(objectName, func(n TypescriptNode) TypescriptNode {
				n.Node = bindings.Alias{
					Name:       bindings.Identifier(objectName),
					Modifiers:  []bindings.Modifier{},
					Type:       rhs.Value,
					Parameters: rhs.TypeParameters,
					Source:     ts.location(obj),
				}
				return n
			})
			return nil
		case *types.Map, *types.Array, *types.Slice:
			// Declared maps that are not structs are still valid codersdk objects.
			// Handle them custom by calling 'typescriptType' directly instead of
			// iterating through each struct field.bindings.Union()
			// These types support no json/typescript tags.
			// These are **NOT** enums, as a map in Go would never be used for an enum.
			ty, err := ts.typescriptType(obj.Type().Underlying())
			if err != nil {
				return xerrors.Errorf("(map) generate %q: %w", objectName, err)
			}

			return ts.SetNode(objectName, TypescriptNode{
				Node: bindings.Alias{
					Name:       bindings.Identifier(objectName),
					Modifiers:  []bindings.Modifier{},
					Type:       ty.Value,
					Parameters: ty.TypeParameters,
					Source:     ts.location(obj),
				},
			})
		case *types.Interface:
			// Interfaces are used as generics. Non-generic interfaces are
			// not supported.
			if underNamed.NumEmbeddeds() == 1 {
				union, ok := underNamed.EmbeddedType(0).(*types.Union)
				if !ok {
					// If the underlying is not a union, but has 1 type. It's
					// just that one type.
					union = types.NewUnion([]*types.Term{
						// Set the tilde to true to support underlying.
						// Doesn't actually affect our generation.
						types.NewTerm(true, underNamed.EmbeddedType(0)),
					})
				}

				block, err := ts.buildUnion(obj, union)
				if err != nil {
					return xerrors.Errorf("generate union %q: %w", objectName, err)
				}
				return ts.SetNode(objectName, TypescriptNode{
					Node: block,
				})
			}
		case *types.Signature:
			// Ignore named functions.
			return nil
		default:
			// If you hit this error, you added a new unsupported named type.
			// The easiest way to solve this is add a new case above with
			// your type and a TODO to implement it.
			return xerrors.Errorf("unsupported named type %q", underNamed.String())
		}
	case *types.Var:
		// TODO: Are any enums var declarations? This is also codersdk.Me.
	case *types.Const:
		// Names are very likely enums
		named, ok := obj.Type().(*types.Named)
		if !ok {
			// It could be a raw const value to generate.
			if _, ok := obj.Type().(*types.Basic); ok {
				cnst, err := ts.constantDeclaration(obj)
				if err != nil {
					return xerrors.Errorf("basic const %q: %w", objectName, err)
				}
				return ts.SetNode(objectName, TypescriptNode{
					Node: cnst,
				})
			}
			return xerrors.Errorf("const %q is not a named type", objectName)
		}

		// Treat it as an enum.
		enumObjName := named.Obj().Name()

		switch named.Underlying().(type) {
		case *types.Basic:
		default:
			return xerrors.Errorf("const %q is not a basic type, enums only support basic", objectName)
		}

		// Grab the value of the constant. This is the enum value.
		constValue, err := ts.constantValue(obj)
		if err != nil {
			return xerrors.Errorf("const %q: %w", objectName, err)
		}

		// This is a little hacky, but we need to add the enum to the Alias
		// type. However, the order types are parsed is not guaranteed, so we
		// add the enum to the Alias as a post-processing step.
		ts.UpdateNode(enumObjName, func(n TypescriptNode) TypescriptNode {
			return n.AddEnum(constValue)
		})
		return nil
	case *types.Func:
		// Noop
		return nil
	default:
		return xerrors.Errorf("unsupported object type %T", obj)
	}

	return xerrors.Errorf("should never hit this")
}

func (ts *Typescript) constantDeclaration(obj *types.Const) (bindings.VariableStatement, error) {
	val, err := ts.constantValue(obj)
	if err != nil {
		return bindings.VariableStatement{}, err
	}

	return bindings.VariableStatement{
		Modifiers: []bindings.Modifier{},
		Declarations: bindings.VariableDeclarationList{
			Declarations: []bindings.VariableDeclaration{
				{
					Name:            obj.Name(),
					ExclamationMark: false,
					Initializer:     val,
				},
			},
			Flags: bindings.NodeFlagsConstant,
		},
		Source: ts.location(obj),
	}, nil
}

func (ts *Typescript) constantValue(obj *types.Const) (bindings.LiteralType, error) {
	var constValue bindings.LiteralType
	switch obj.Val().Kind() {
	case constant.String:
		constValue.Value = constant.StringVal(obj.Val())
	case constant.Int:
		// TODO: might want to check this
		constValue.Value, _ = constant.Int64Val(obj.Val())
	case constant.Float:
		constValue.Value, _ = constant.Float64Val(obj.Val())
	case constant.Bool:
		constValue.Value = constant.BoolVal(obj.Val())
	default:
		return bindings.LiteralType{}, xerrors.Errorf("const %q is not a supported basic type, enums only support basic", obj.Name())
	}
	return constValue, nil
}

// buildStruct just prints the typescript def for a type.
// Generic type parameters are inferred from the type and inferred.
func (ts *Typescript) buildStruct(obj types.Object, st *types.Struct) (bindings.Interface, error) {
	tsi := bindings.Interface{
		Name:       bindings.Identifier(obj.Name()),
		Modifiers:  []bindings.Modifier{},
		Fields:     []bindings.PropertySignature{},
		Parameters: []bindings.TypeParameter{},  // Generics
		Heritage:   []bindings.HeritageClause{}, // Extends
		Source:     ts.location(obj),
	}

	// Handle named embedded structs in the codersdk package via extension.
	// This is inheritance.
	// TODO: Maybe this could be done inline in the main for loop?
	var extends []ParsedType
	for i := 0; i < st.NumFields(); i++ {
		field := st.Field(i)
		tag := reflect.StructTag(st.Tag(i))
		// Adding a json struct tag causes the json package to consider
		// the field unembedded.
		if field.Embedded() && tag.Get("json") == "" {
			// TODO: Generic args
			heritage, err := ts.typescriptType(field.Type())
			if err != nil {
				return tsi, xerrors.Errorf("heritage type: %w", err)
			}
			extends = append(extends, heritage)
		}
	}

	if len(extends) > 0 {
		var heritages []bindings.ExpressionType
		for _, heritage := range extends {
			heritages = append(heritages, heritage.Value)
		}
		tsi.Heritage = append(tsi.Heritage, bindings.HeritageClauseExtends(heritages...))
	}

	if _, ok := obj.(*types.TypeName); ok {
		var typeParamed interface{ TypeParams() *types.TypeParamList }
		switch typedObj := obj.Type().(type) {
		case *types.Named:
			typeParamed = typedObj
		case *types.Alias:
			typeParamed = typedObj
		default:
			return tsi, xerrors.Errorf("not supported top level type %T for %q", obj.Type(), obj.Name())
		}

		// This code is usually redundant, as we infer generics from the
		// child usage. However, if the field is unused, then this comes in
		// handy.
		// Note: Maybe we can remove all generic values bubbling up in favor
		// of this?
		// Note: Maybe do not even need this, as it includes unused generics.
		typeParameters, err := ts.typeParametersParameters(typeParamed)
		if err != nil {
			return tsi, xerrors.Errorf("type parameters: %w", err)
		}
		tsi.Parameters = typeParameters
	}

	// Iterate through the fields of the struct.
	for i := 0; i < st.NumFields(); i++ {
		field := st.Field(i)
		tag := reflect.StructTag(st.Tag(i))
		tags, err := structtag.Parse(string(tag))
		if err != nil {
			panic("invalid struct tags on type " + obj.String())
		}

		if field.Embedded() {
			// Heritage was done above
			continue
		}

		if !field.Exported() {
			// Skip unexported fields
			continue
		}

		// Create a new field in the intermediate typescript representation.
		tsField := bindings.PropertySignature{
			Name:          field.Name(),
			Modifiers:     []bindings.Modifier{},
			QuestionToken: false,
			Type:          nil,
		}

		// Use the json name if present
		jsonTag, err := tags.Get("json")
		if err == nil {
			if jsonTag.Name == "-" {
				// Completely ignore this field.
				continue
			}
			tsField.Name = jsonTag.Name
			if len(jsonTag.Options) > 0 && jsonTag.Options[0] == "omitempty" {
				tsField.QuestionToken = true
			}
		}

		// Infer the type.
		tsType, err := ts.typescriptType(field.Type())
		if err != nil {
			return tsi, xerrors.Errorf("typescript type: %w", err)
		}
		tsField.Type = tsType.Value
		tsi.Parameters = append(tsi.Parameters, tsType.TypeParameters...)
		tsField.FieldComments = tsType.RaisedComments

		// Some tag support
		// TODO: Add more tag support?
		typescriptTag, err := tags.Get("typescript")
		if err == nil {
			if typescriptTag.Name == "-" {
				// Completely ignore this field.
				continue
			}
		}

		tsi.Fields = append(tsi.Fields, tsField)
	}

	simple, err := bindings.Simplify(tsi.Parameters)
	if err != nil {
		return tsi, xerrors.Errorf("simplify generics: %w", err)
	}
	tsi.Parameters = simple
	return tsi, nil
}

type ParsedType struct {
	// Value is the typescript type of the passed in go type.
	Value bindings.ExpressionType
	// TypeParameters are any generic types that are used in the Value.
	TypeParameters []bindings.TypeParameter
	// RaisedComments exists to add comments to the first parent that is willing
	// to accept them. It is for formatting purposes.
	RaisedComments []string
}

func SimpleParsedType(et bindings.ExpressionType) ParsedType {
	return ParsedType{
		Value: et,
	}
}

func (p ParsedType) WithComments(comments ...string) ParsedType {
	p.RaisedComments = append(p.RaisedComments, comments...)
	return p
}

// TODO: Return comments?
func (ts *Typescript) typescriptType(ty types.Type) (ParsedType, error) {
	switch ty := ty.(type) {
	case *types.Basic:
		bs := ty
		// All basic literals (string, bool, int, etc).
		switch {
		case bs.Info()&types.IsNumeric > 0:
			return SimpleParsedType(bindings.KeywordNumber), nil
		case bs.Info()&types.IsBoolean > 0:
			return SimpleParsedType(bindings.KeywordBoolean), nil
		case bs.Kind() == types.Byte:
			// TODO: @emyrk What is a byte for typescript? A string? A uint8?
			// TODO: Comment
			//return bindings.PrependComment("This is a byte in golang", bindings.Literal(bindings.KeywordNumber)), nil
			return SimpleParsedType(bindings.KeywordNumber), nil
		case bs.Kind() == types.String, bs.Kind() == types.Rune:
			return SimpleParsedType(bindings.KeywordString), nil
		default:
			return ParsedType{}, xerrors.Errorf("unsupported basic type %q", bs.String())
		}
	case *types.Struct:
		// This handles anonymous structs. This should never happen really.
		// If you require this, either change your datastructures, or implement
		// anonymous structs here.
		// Such as:
		//  type Name struct {
		//	  Embedded struct {
		//		  Field string `json:"field"`
		//	  }
		//  }
		// TODO: Comment: indentedComment("Embedded anonymous struct, please fix by naming it"),
		parsed := SimpleParsedType(bindings.KeywordUnknown)
		parsed.RaisedComments = append(parsed.RaisedComments, "embedded anonymous struct, please fix by naming it")
		return parsed, nil
	case *types.Map:
		// Record is reference type with 2 type parameters.
		// map[string][string] -> Record<string, string>

		m := ty
		keyType, err := ts.typescriptType(m.Key())
		if err != nil {
			return ParsedType{}, xerrors.Errorf("map key: %w", err)
		}
		valueType, err := ts.typescriptType(m.Elem())
		if err != nil {
			return ParsedType{}, xerrors.Errorf("map key: %w", err)
		}

		tp, err := bindings.Simplify(append(keyType.TypeParameters, valueType.TypeParameters...))
		if err != nil {
			return ParsedType{}, xerrors.Errorf("simplify generics in map: %w", err)
		}
		parsed := ParsedType{
			Value:          bindings.Reference("Record", keyType.Value, valueType.Value),
			TypeParameters: tp,
			RaisedComments: append(keyType.RaisedComments, valueType.RaisedComments...),
		}
		return parsed, nil
	case *types.Slice, *types.Array:
		// Slice/Arrays are pretty much the same.
		type hasElem interface {
			Elem() types.Type
		}

		arr, _ := ty.(hasElem)
		switch {
		// When type checking here, just use the string. You can cast it
		// to a types.Basic and get the kind if you want too :shrug:
		case arr.Elem().String() == "byte":
			// All byte arrays are strings on the typescript.
			// Is this ok?
			return SimpleParsedType(bindings.Array(bindings.KeywordString)), nil
		default:
			// By default, just do an array of the underlying type.
			underlying, err := ts.typescriptType(arr.Elem())
			if err != nil {
				return ParsedType{}, xerrors.Errorf("array: %w", err)
			}
			//genValue := ""
			//
			//if underlying.GenericValue != "" {
			//	genValue = "Readonly<Array<" + underlying.GenericValue + ">>"
			//}
			return ParsedType{
				Value:          bindings.Array(underlying.Value),
				TypeParameters: underlying.TypeParameters,
				RaisedComments: underlying.RaisedComments,
			}, nil
			//return TypescriptType{
			//	ValueType:     "Readonly<Array<" + underlying.ValueType + ">>",
			//	GenericValue:  genValue,
			//	AboveTypeLine: underlying.AboveTypeLine,
			//	GenericTypes:  underlying.GenericTypes,
			//}, nil
		}
	case *types.Named:
		n := ty

		// These are external named types that we handle uniquely.
		// This is unfortunate, but our current code assumes all defined
		// types are enums, but these are really just basic primitives.
		// We would need to add more logic to determine this, but for now
		// just hard code them.
		// TODO: Allow comments here
		custom, ok := ts.parsed.customMappings[n.String()]
		if ok {
			return ParsedType{
				Value: custom,
			}, nil
		}

		// If it is not a custom mapping, we should assume the type is
		// defined elsewhere. We want to know where and what that definition
		// is, such that we can raise up any type parameters.
		ref, ok := ts.parsed.lookupNamedReference(n)
		if ok {
			if ref.Pkg().Path() != n.Obj().Pkg().Path() {
				slog.Info("found external type", slog.String("name", ref.Name()), slog.String("ext_pkg", ref.Pkg().Path()))
			}

			args, err := ts.typeParametersArgs(n)
			if err != nil {
				return ParsedType{}, xerrors.Errorf("type parameter arguments: %w", err)
			}

			parsed := ParsedType{}
			exprArgs := make([]bindings.ExpressionType, 0, len(args))
			for _, arg := range args {
				exprArgs = append(exprArgs, arg.Value)
				parsed.TypeParameters = append(parsed.TypeParameters, arg.TypeParameters...)
				parsed.RaisedComments = append(parsed.RaisedComments, arg.RaisedComments...)
			}
			parsed.Value = bindings.Reference(ref.Name(), exprArgs...)

			return parsed, nil
		}

		// If it's a struct, just use the name of the struct type
		if _, ok := n.Underlying().(*types.Struct); ok {
			// This struct comes from an external package that we did not parse.
			// We can introspect it, but then it acts as an anonymous struct
			// embed. Structs should be flat in their fields, so just return a
			// reference with a comment.
			return SimpleParsedType(bindings.KeywordUnknown).WithComments(
				// '.Include(<pkg_path>, false)' to include this type
				fmt.Sprintf("external type %q, to include this type the package must be explictly included in the parsing", n.String())), nil
		}

		// Defer to the underlying type.
		ts, err := ts.typescriptType(ty.Underlying())
		if err != nil {
			return ParsedType{}, xerrors.Errorf("named underlying: %w", err)
		}

		return ts.WithComments(fmt.Sprintf("this is likely an enum in an external package %q", n.String())), nil
	case *types.Pointer:
		// Dereference pointers.
		pt := ty
		resp, err := ts.typescriptType(pt.Elem())
		if err != nil {
			return ParsedType{}, xerrors.Errorf("pointer: %w", err)
		}

		// Golang pointers can json marshal to 'null' if they are nil
		resp.Value = bindings.Union(resp.Value, bindings.Null{})
		return resp, nil
	case *types.Interface:
		// only handle the empty interface (interface{}) for now
		intf := ty
		if intf.Empty() {
			// This field is 'interface{}'. We can't infer any type from 'interface{}'
			// so just use "unknown" as the type.
			parsed := SimpleParsedType(bindings.KeywordUnknown)
			parsed.RaisedComments = append(parsed.RaisedComments, "empty interface{} type, falling back to unknown")
			return parsed, nil
			//return TypescriptType{
			//	AboveTypeLine: indentedComment("empty interface{} type, falling back to unknown"),
			//	ValueType:     "unknown",
			//}, nil
		}

		if intf.NumEmbeddeds() == 1 {
			parsedI, err := ts.typescriptType(intf.EmbeddedType(0))
			if err != nil {
				return ParsedType{}, xerrors.Errorf("parse interface: %w", err)
			}
			return parsedI, nil
		}

		// Interfaces are difficult to determine the JSON type, so just return
		// an 'unknown'.
		parsed := SimpleParsedType(bindings.KeywordUnknown)
		parsed.RaisedComments = append(parsed.RaisedComments, "interface type, falling back to unknown")
		return parsed, nil
	case *types.TypeParam:
		_, ok := ty.Underlying().(*types.Interface)
		if !ok {
			// If it's not an interface, it is likely a usage of generics that
			// we have not hit yet. Feel free to add support for it.
			return ParsedType{}, xerrors.New("type param must be an interface")
		}

		// type Foo[T any] struct {
		name := ty.Obj().Name()    // T
		generic := ty.Constraint() // generic

		// We don't mess with multiple packages, so just trim the package path
		// from the name.
		pkgPath := ty.Obj().Pkg().Path()
		constraintName := strings.TrimPrefix(generic.String(), pkgPath+".")

		// Any is the default
		var constraintNode bindings.ExpressionType
		switch constraintName {
		case "comparable":
			// TODO: Generate this on demand.
			constraintNode = bindings.Reference(BuiltInComparable)
			ts.includeComparable()
		case "any":
			constraintNode = bindings.KeywordAny
		default:
			parsedGeneric, err := ts.typescriptType(generic)
			if err != nil {
				return ParsedType{}, xerrors.Errorf("type param %q: %w", generic.String(), err)
			}

			// TODO: Raise comments and generics?
			constraintNode = parsedGeneric.Value
		}

		return ParsedType{
			Value: bindings.Reference(name),
			TypeParameters: []bindings.TypeParameter{
				{
					Name:      name,
					Modifiers: []bindings.Modifier{},
					// All generics in Golang have some type of constraint (even if it's 'any').
					// TODO: if the constraint is 'any', we should probably not bother with the type
					// It is redundant.
					Type:        constraintNode,
					DefaultType: nil,
				},
			},
		}, nil
	case *types.Alias:
		// TODO: Verify this is correct.
		return ts.typescriptType(ty.Underlying())
	}

	// These are all the other types we need to support.
	return ParsedType{}, xerrors.Errorf("unknown type: %s", ty.String())
}

// buildStruct just prints the typescript def for a type.
func (ts *Typescript) buildUnion(obj types.Object, st *types.Union) (bindings.Alias, error) {
	alias := bindings.Alias{
		Name:       bindings.Identifier(obj.Name()),
		Modifiers:  []bindings.Modifier{},
		Type:       nil,
		Parameters: nil,
		Source:     ts.location(obj),
	}

	allTypes := make([]bindings.ExpressionType, 0, st.Len())
	for i := 0; i < st.Len(); i++ {
		term := st.Term(i)
		scriptType, err := ts.typescriptType(term.Type())
		if err != nil {
			return alias, xerrors.Errorf("union %q for %q failed to get type: %w", st.String(), obj.Name(), err)
		}
		// TODO: Generics
		// scriptType.TypeParameters
		allTypes = append(allTypes, scriptType.Value)
	}

	alias.Type = bindings.Union(allTypes...)
	return alias, nil
}

// typeParametersParameters extracts the generic parameters from a named type.
func (ts *Typescript) typeParametersParameters(obj interface{ TypeParams() *types.TypeParamList }) ([]bindings.TypeParameter, error) {
	args := obj.TypeParams()
	if args == nil || args.Len() == 0 {
		return []bindings.TypeParameter{}, nil
	}

	params := make([]bindings.TypeParameter, 0, args.Len())
	for i := 0; i < args.Len(); i++ {
		arg := args.At(i)
		argType, err := ts.typescriptType(arg)
		if err != nil {
			return nil, xerrors.Errorf("type parameter %q: %w", arg.String(), err)
		}

		params = append(params, argType.TypeParameters...)
	}
	return params, nil
}

func (ts *Typescript) typeParametersArgs(obj *types.Named) ([]ParsedType, error) {
	args := obj.TypeArgs()
	if args == nil || args.Len() == 0 {
		return []ParsedType{}, nil
	}

	params := make([]ParsedType, 0, args.Len())
	for i := 0; i < args.Len(); i++ {
		arg := args.At(i)
		argType, err := ts.typescriptType(arg)
		if err != nil {
			return nil, xerrors.Errorf("type parameter %q: %w", arg.String(), err)
		}
		params = append(params, argType)
	}
	return params, nil
}

func (p *GoParser) lookupNamedReference(n *types.Named) (types.Object, bool) {
	lookupPkg := n.Obj().Pkg().Path()
	pkg, ok := p.Pkgs[lookupPkg]
	if !ok {
		return nil, false
	}

	lookupName := n.Obj().Name()
	obj := pkg.Types.Scope().Lookup(lookupName)
	if obj == nil {
		return nil, false
	}
	return obj, true
}
