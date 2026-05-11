package bindings_test

import (
	"fmt"
	"go/token"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/coder/guts/bindings"
)

func TestSyntheticComments(t *testing.T) {
	b, err := bindings.New()
	require.NoError(t, err)

	str := bindings.LiteralKeyword(bindings.KeywordString)
	param := &bindings.TypeParameter{
		Name: bindings.Identifier{
			Name:    "testparam",
			Package: nil,
			Prefix:  "",
		},
		Modifiers:   nil,
		Type:        &str,
		DefaultType: nil,
		//SupportComments: bindings.SupportComments{},
	}
	//param.LeadingComment("a type parameter")

	exp := &bindings.Interface{
		Name: bindings.Identifier{
			Name: "TestingInterface",
		},
		Modifiers: nil,
		Fields:    nil,
		Parameters: []*bindings.TypeParameter{
			param,
		},
		Heritage: nil,
		Source: bindings.Source{
			File: "test.go",
			Position: token.Position{
				Filename: "test.go",
				Offset:   0,
				Line:     5,
				Column:   10,
			},
		},
	}

	exp.AppendComment(bindings.SyntheticComment{
		Leading:         true,
		SingleLine:      true,
		Text:            "hello world",
		TrailingNewLine: false,
	})

	exp.AppendComment(bindings.SyntheticComment{
		Leading:         false,
		SingleLine:      true,
		Text:            "goodbye world",
		TrailingNewLine: false,
	})

	node, err := b.ToTypescriptNode(exp)
	require.NoError(t, err)

	ts, err := b.SerializeToTypescript(node)
	require.NoError(t, err)
	fmt.Println(ts)
}
