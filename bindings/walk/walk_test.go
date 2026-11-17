package walk_test

import (
	"fmt"
	"testing"

	"github.com/coder/guts/bindings"
	"github.com/coder/guts/bindings/walk"
)

// recordingVisitor stores every node it visits so tests can assert that
// Walk descends into expected child slots without panicking.
type recordingVisitor struct {
	visited []bindings.Node
}

func (r *recordingVisitor) Visit(node bindings.Node) walk.Visitor {
	r.visited = append(r.visited, node)
	return r
}

// keyword returns a fresh pointer to a LiteralKeyword. LiteralKeyword is a
// string-typed Node, so callers need a pointer-to-string-typed-value, not
// a composite literal.
func keyword(k bindings.LiteralKeyword) *bindings.LiteralKeyword {
	return &k
}

// TestWalkCoversAllNodeTypes builds a synthetic tree that contains every
// concrete bindings.Node implementation and walks it. The default branch
// in Walk panics on unhandled types, so this test fails loudly if a new
// node is added to the bindings package without a matching case here.
func TestWalkCoversAllNodeTypes(t *testing.T) {
	t.Parallel()

	// Construct one of each Node. Some nodes (PropertyAssignment, Parameter,
	// HeritageClause, ImportSpecifier, EnumMember) only appear as children
	// of a parent, so the parent's slot is the way Walk reaches them.
	tree := &bindings.Interface{
		Name: bindings.Identifier{Name: "Root"},
		Parameters: []*bindings.TypeParameter{
			{Name: bindings.Identifier{Name: "T"}, Type: keyword(bindings.KeywordString)},
		},
		Heritage: []*bindings.HeritageClause{
			{
				Args: []bindings.ExpressionType{
					&bindings.ExpressionWithTypeArguments{
						Expression: &bindings.IdentifierExpression{
							Name: bindings.Identifier{Name: "Base"},
						},
						Arguments: []bindings.ExpressionType{
							&bindings.ReferenceType{Name: bindings.Identifier{Name: "X"}},
						},
					},
				},
			},
		},
		Fields: []*bindings.PropertySignature{
			{
				Name: "everything",
				Type: &bindings.TypeIntersection{
					Types: []bindings.ExpressionType{
						&bindings.UnionType{
							Types: []bindings.ExpressionType{
								&bindings.LiteralType{Value: "a"},
								&bindings.Null{},
								keyword(bindings.KeywordString),
							},
						},
						&bindings.TypeLiteralNode{
							Members: []*bindings.PropertySignature{
								{Name: "nested", Type: keyword(bindings.KeywordString)},
							},
						},
						&bindings.ArrayType{Node: keyword(bindings.KeywordString)},
						&bindings.TupleType{Node: keyword(bindings.KeywordString)},
						&bindings.ArrayLiteralType{
							Elements: []bindings.ExpressionType{&bindings.LiteralType{Value: "elt"}},
						},
						bindings.OperatorNode(bindings.KeywordReadonly, keyword(bindings.KeywordString)),
						&bindings.TypeQuery{Name: bindings.Identifier{Name: "Other"}},
					},
				},
			},
		},
	}

	// Independent node group exercising the value-expression and import
	// nodes added in PR 83 and PR 84. Walking each top-level node here
	// touches the remaining bindings.Node implementations.
	nodes := []bindings.Node{
		tree,
		&bindings.Alias{
			Name: bindings.Identifier{Name: "MyAlias"},
			Type: keyword(bindings.KeywordString),
		},
		&bindings.Enum{
			Name: bindings.Identifier{Name: "Color"},
			Members: []*bindings.EnumMember{
				{Name: "Red", Value: &bindings.LiteralType{Value: "red"}},
			},
		},
		&bindings.VariableStatement{
			Declarations: &bindings.VariableDeclarationList{
				Flags: bindings.NodeFlagsConstant,
				Declarations: []*bindings.VariableDeclaration{
					{
						Name: bindings.Identifier{Name: "Schema"},
						Type: &bindings.ReferenceType{Name: bindings.Identifier{Name: "Z"}},
						Initializer: &bindings.CallExpression{
							Expression: &bindings.PropertyAccessExpression{
								Expression: &bindings.IdentifierExpression{
									Name: bindings.Identifier{Name: "z"},
								},
								Name: "object",
							},
							Arguments: []bindings.ExpressionType{
								&bindings.ObjectLiteralExpression{
									Properties: []*bindings.PropertyAssignment{
										{
											Name: "id",
											Initializer: &bindings.CallExpression{
												Expression: &bindings.PropertyAccessExpression{
													Expression: &bindings.IdentifierExpression{
														Name: bindings.Identifier{Name: "z"},
													},
													Name: "string",
												},
											},
										},
									},
								},
								&bindings.ArrowFunction{
									Parameters: []*bindings.Parameter{
										{Name: "x", Type: keyword(bindings.KeywordString)},
									},
									ReturnType: keyword(bindings.KeywordString),
									Body: &bindings.IdentifierExpression{
										Name: bindings.Identifier{Name: "x"},
									},
								},
							},
						},
					},
				},
			},
		},
		&bindings.ImportDeclaration{
			Module: "zod",
			Named: []*bindings.ImportSpecifier{
				{Name: "z"},
			},
		},
	}

	v := &recordingVisitor{}
	for _, node := range nodes {
		walk.Walk(v, node)
	}

	// Verify that Walk reached every new node type. The set of expected
	// node types here is the union of nodes constructed above. If any are
	// missing, Walk lost a child slot somewhere along the way.
	seen := map[string]bool{}
	for _, n := range v.visited {
		seen[fmt.Sprintf("%T", n)] = true
	}

	want := []string{
		"*bindings.Interface",
		"*bindings.TypeParameter",
		"*bindings.HeritageClause",
		"*bindings.ExpressionWithTypeArguments",
		"*bindings.IdentifierExpression",
		"*bindings.ReferenceType",
		"*bindings.PropertySignature",
		"*bindings.TypeIntersection",
		"*bindings.UnionType",
		"*bindings.LiteralType",
		"*bindings.Null",
		"*bindings.LiteralKeyword",
		"*bindings.TypeLiteralNode",
		"*bindings.ArrayType",
		"*bindings.TupleType",
		"*bindings.ArrayLiteralType",
		"*bindings.OperatorNodeType",
		"*bindings.TypeQuery",
		"*bindings.Alias",
		"*bindings.Enum",
		"*bindings.EnumMember",
		"*bindings.VariableStatement",
		"*bindings.VariableDeclarationList",
		"*bindings.VariableDeclaration",
		"*bindings.CallExpression",
		"*bindings.PropertyAccessExpression",
		"*bindings.ObjectLiteralExpression",
		"*bindings.PropertyAssignment",
		"*bindings.ArrowFunction",
		"*bindings.Parameter",
		"*bindings.ImportDeclaration",
		"*bindings.ImportSpecifier",
	}
	for _, name := range want {
		if !seen[name] {
			t.Errorf("Walk did not visit %s", name)
		}
	}
}

// TestWalkStopsWhenVisitReturnsNil verifies that Walk honours the Visitor
// contract: returning nil from Visit halts descent into that subtree.
func TestWalkStopsWhenVisitReturnsNil(t *testing.T) {
	t.Parallel()

	leaf := keyword(bindings.KeywordString)
	parent := &bindings.ArrayType{Node: leaf}

	v := &stoppingVisitor{stopAt: parent}
	walk.Walk(v, parent)

	if len(v.visited) != 1 {
		t.Fatalf("expected 1 visit, got %d", len(v.visited))
	}
	if v.visited[0] != bindings.Node(parent) {
		t.Fatalf("expected to visit parent, got %T", v.visited[0])
	}
}

type stoppingVisitor struct {
	stopAt  bindings.Node
	visited []bindings.Node
}

func (s *stoppingVisitor) Visit(node bindings.Node) walk.Visitor {
	s.visited = append(s.visited, node)
	if node == s.stopAt {
		return nil
	}
	return s
}
