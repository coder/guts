// How to use compiler API
// https://github.com/microsoft/TypeScript/wiki/Using-the-Compiler-API#creating-and-printing-a-typescript-ast

// AST Node types:
// https://github.com/typescript-eslint/typescript-eslint/blob/244435126619afb9497ace04cbf4819012e27330/packages/ast-spec/src/ast-node-types.ts#L144
// AST Node docs:
// https://typescript-eslint.io/packages/ast-spec/generated/#arrayexpression

// Helpful resources:
// https://ts-ast-viewer.com/
// https://astexplorer.net/

import * as ts from "typescript";
import { Modifier, ModifierSyntaxKind } from "typescript";

type modifierKeys = FilterKeys<typeof ts.SyntaxKind, ModifierSyntaxKind>;
export function modifier(name: modifierKeys | ts.Modifier): Modifier {
  if (typeof name !== "string") {
    return name;
  }

  const x = ts.SyntaxKind[name];

  return ts.factory.createModifier(x);
}

export function identifier(name: string | ts.Identifier): ts.Identifier {
  if (typeof name !== "string") {
    return name;
  }
  return ts.factory.createIdentifier(name);
}

export function propertySignature(
  modifiers: readonly modifierKeys[] | readonly ts.Modifier[] | undefined,
  name: string,
  question: boolean,
  type: ts.TypeNode
): ts.PropertySignature {
  return ts.factory.createPropertySignature(
    modifiers?.map((m) => modifier(m)),
    identifier(name),
    question ? questionToken() : undefined,
    type
  );
}

export function questionToken(): ts.QuestionToken {
  return ts.factory.createToken(ts.SyntaxKind.QuestionToken);
}

export function reference(
  name: string | ts.Identifier,
  typeParameters: ts.TypeNode[]
): ts.TypeReferenceNode {
  return ts.factory.createTypeReferenceNode(
    identifier(name),
    typeParameters // Type arguments, generics
  );
}

type keywordKeys = FilterKeys<typeof ts.SyntaxKind, ts.KeywordTypeSyntaxKind>;
export function literalKeyword(keyword: keywordKeys): ts.KeywordTypeNode {
  return ts.factory.createKeywordTypeNode(ts.SyntaxKind[keyword]);
}

// ts.factory.createInterfaceDeclaration()

// resultFile is some file context, not really used
const resultFile = ts.createSourceFile(
  "generated.ts",
  "",
  ts.ScriptTarget.Latest,
  /*setParentNodes*/ false,
  ts.ScriptKind.TS
);

const savedNodes: ts.Node[] = [];

// printer is used to convert AST to string
const printer = ts.createPrinter({
  newLine: ts.NewLineKind.LineFeed,
  removeComments: false,
});

export function toTypescript(node: ts.Node): string {
  return printer.printNode(ts.EmitHint.Unspecified, node, resultFile);
}

export function interfaceDecl(
  modifiers: readonly modifierKeys[] | readonly ts.Modifier[] | undefined,
  name: ts.Identifier | string,
  typeParameters: ts.TypeParameterDeclaration[] | undefined,
  heritageClauses: ts.HeritageClause[] | undefined,
  fields: ts.TypeElement[]
): ts.InterfaceDeclaration {
  const node = ts.factory.createInterfaceDeclaration(
    modifiers?.map((m) => modifier(m)),
    identifier(name),
    typeParameters, // Generics
    heritageClauses, // Inheritance
    fields // Fields
  );
  printer.printNode(ts.EmitHint.Unspecified, node, resultFile);
  return node;
}

export function arrayType(node: ts.TypeNode): ts.ArrayTypeNode {
  return ts.factory.createArrayTypeNode(node);
}

export function aliasDecl(
  modifiers: readonly modifierKeys[] | readonly ts.Modifier[] | undefined,
  name: ts.Identifier | string,
  typeParameters: ts.TypeParameterDeclaration[] | undefined,
  type: ts.TypeNode
): ts.TypeAliasDeclaration {
  // (modifiers: readonly ModifierLike[] | undefined, name: string | Identifier, typeParameters: readonly TypeParameterDeclaration[] | undefined, type: TypeNode): TypeAliasDeclaration;
  return ts.factory.createTypeAliasDeclaration(
    modifiers?.map((m) => modifier(m)),
    identifier(name),
    typeParameters, // Generics
    type // Type
  );
}

export function typeParameterDeclaration(
  modifiers: readonly modifierKeys[] | readonly ts.Modifier[] | undefined,
  name: ts.Identifier | string,
  type?: ts.TypeNode,
  defaultType?: ts.TypeNode
): ts.TypeParameterDeclaration {
  return ts.factory.createTypeParameterDeclaration(
    modifiers?.map((m) => modifier(m)),
    identifier(name),
    type,
    defaultType
  );
}

export function unionType(nodes: ts.TypeNode[]): ts.UnionTypeNode {
  return ts.factory.createUnionTypeNode(nodes);
}

export function createNull(): ts.NullLiteral {
  return ts.factory.createNull();
}

export function addSyntheticComment(
  node: ts.Node,
  leading: boolean,
  singleLine: boolean,
  text: string,
  hasTrailingNewLine: boolean
): ts.Node {
  const kind = singleLine
    ? ts.SyntaxKind.SingleLineCommentTrivia
    : ts.SyntaxKind.MultiLineCommentTrivia;
  if (leading) {
    return ts.addSyntheticLeadingComment(node, kind, text, hasTrailingNewLine);
  } else {
    return ts.addSyntheticTrailingComment(node, kind, text, hasTrailingNewLine);
  }
}

export function heritageClause(
  tokenString: "extends" | "implements",
  types: ts.ExpressionWithTypeArguments[]
): ts.HeritageClause {
  const token =
    tokenString === "extends"
      ? ts.SyntaxKind.ExtendsKeyword
      : ts.SyntaxKind.ImplementsKeyword;
  return ts.factory.createHeritageClause(token, types);
}

export function stringLiteral(value: string): ts.StringLiteral {
  return ts.factory.createStringLiteral(value);
}

export function numericLiteral(value: number): ts.NumericLiteral {
  return ts.factory.createNumericLiteral(value.toString());
}

export function booleanLiteral(value: boolean): ts.BooleanLiteral {
  if (value) {
    return ts.factory.createTrue();
  }
  return ts.factory.createFalse();
}

export function literalTypeNode(
  node: ts.LiteralTypeNode["literal"]
): ts.LiteralTypeNode {
  return ts.factory.createLiteralTypeNode(node);
}

export function variableStatement(
  modifiers: readonly modifierKeys[] | readonly ts.Modifier[] | undefined,
  list: ts.VariableDeclarationList
) {
  return ts.factory.createVariableStatement(
    modifiers?.map((m) => modifier(m)),
    list
  );
}

export function variableDeclarationList(
  declarations: ts.VariableDeclaration[],
  flags: ts.NodeFlags
): ts.VariableDeclarationList {
  return ts.factory.createVariableDeclarationList(declarations, flags);
}

export function variableDeclaration(
  name: string,
  exclamation: boolean,
  type?: ts.TypeNode,
  initializer?: ts.Expression
): ts.VariableDeclaration {
  return ts.factory.createVariableDeclaration(
    name,
    exclamation
      ? ts.factory.createToken(ts.SyntaxKind.ExclamationToken)
      : undefined,
    type,
    initializer
  );
}

export function arrayLiteral(
  elements: ts.Expression[]
): ts.ArrayLiteralExpression {
  return ts.factory.createArrayLiteralExpression(elements);
}

export function typeOperatorNode(
  operator: "KeyOfKeyword" | "UniqueKeyword" | "ReadonlyKeyword",
  node: ts.TypeNode
): ts.TypeOperatorNode {
  return ts.factory.createTypeOperatorNode(ts.SyntaxKind[operator], node);
}

module.exports = {
  modifier: modifier,
  identifier: identifier,
  propertySignature: propertySignature,
  questionToken: questionToken,
  reference: reference,
  toTypescript: toTypescript,
  interfaceDecl: interfaceDecl,
  literalKeyword: literalKeyword,
  arrayType: arrayType,
  aliasDecl: aliasDecl,
  typeParameterDeclaration: typeParameterDeclaration,
  unionType: unionType,
  createNull: createNull,
  addSyntheticComment: addSyntheticComment,
  heritageClause: heritageClause,
  stringLiteral: stringLiteral,
  numericLiteral: numericLiteral,
  booleanLiteral: booleanLiteral,
  literalTypeNode: literalTypeNode,
  variableStatement: variableStatement,
  variableDeclaration: variableDeclaration,
  variableDeclarationList: variableDeclarationList,
  arrayLiteral: arrayLiteral,
  typeOperatorNode: typeOperatorNode,
};
