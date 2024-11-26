import * as ts from "typescript";
import * as index from "./index";

// const decl = index.interfaceDecl(["ExportKeyword"], "MyInterface");
// console.log(index.toTypescript(decl));

const decl = ts.factory.createVariableStatement(
  undefined,
  ts.factory.createVariableDeclarationList(
    [
      ts.factory.createVariableDeclaration(
        ts.factory.createIdentifier("ToThe"),
        undefined,
        undefined,
        ts.factory.createStringLiteral("side")
      ),
    ],
    ts.NodeFlags.Const
  )
);

console.log(index.toTypescript(decl));
