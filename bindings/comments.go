package bindings

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

// Commentable indicates if the AST node supports adding comments.
// Any number of comments are supported and can be attached to a typescript AST node.
type Commentable interface {
	Comment(comment SyntheticComment)
	Comments() []SyntheticComment
}

// SyntheticComment is the state of a comment for a given AST node.
// See the compiler for how these are serialized.
type SyntheticComment struct {
	Leading         bool
	SingleLine      bool
	Text            string
	TrailingNewLine bool
}

type SupportComments struct {
	comments []SyntheticComment
}

func (s *SupportComments) ASTCommentGroup(grp *ast.CommentGroup) {
	if grp == nil {
		return
	}
	for _, cmt := range grp.List {
		s.ASTComment(cmt)
	}
}

func (s *SupportComments) ASTComment(cmt *ast.Comment) {
	// TODO: Is there a better way to get just the text of the comment?
	text := cmt.Text
	text = strings.TrimPrefix(text, "//")
	text = strings.TrimPrefix(text, "/*")
	text = strings.TrimSuffix(text, "*/")

	s.Comment(SyntheticComment{
		Leading:         true,
		SingleLine:      !strings.Contains(cmt.Text, "\n"),
		Text:            text,
		TrailingNewLine: true,
	})
}

// LeadingComment is a helper function for the most common type of comment.
func (s *SupportComments) LeadingComment(text string) {
	s.Comment(SyntheticComment{
		Leading:    true,
		SingleLine: true,
		// All go comments are `// ` prefixed, so add a space.
		Text:            " " + text,
		TrailingNewLine: false,
	})
}

func (s *SupportComments) Comment(comment SyntheticComment) {
	s.comments = append(s.comments, comment)
}

func (s *SupportComments) Comments() []SyntheticComment {
	return s.comments
}

type HasSource interface {
	SourceComment() (SyntheticComment, bool)
}

// Source is the golang file that an entity is sourced from.
type Source struct {
	File     string
	Position token.Position
}

// SourceComment returns a synthetic comment indicating the source file.
// If the source file is empty, the second return is false.
func (s Source) SourceComment() (SyntheticComment, bool) {
	return SyntheticComment{
		Leading:         true,
		SingleLine:      true,
		Text:            fmt.Sprintf(" From %s", s.File),
		TrailingNewLine: false,
	}, s.File != ""
}
