package bindings

import (
	"fmt"
	"go/token"
)

// Commentable indicates if the AST node supports adding comments.
// Any number of comments are supported and can be attached to a typescript AST node.
type Commentable interface {
	AppendComment(comment SyntheticComment)
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

// LeadingComment is a helper function for the most common type of comment.
func (s *SupportComments) LeadingComment(text string) {
	s.AppendComment(SyntheticComment{
		Leading:    true,
		SingleLine: true,
		// All go comments are `// ` prefixed, so add a space.
		Text:            " " + text,
		TrailingNewLine: false,
	})
}

func (s *SupportComments) AppendComments(comments []SyntheticComment) {
	s.comments = append(s.comments, comments...)
}

func (s *SupportComments) AppendComment(comment SyntheticComment) {
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
