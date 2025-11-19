package comments

// CommentedStructure is a struct with a comment.
//
// It actually has 2 comments?!
// TODO: Maybe add a third comment!
type CommentedStructure struct {
	Inline string // Field comment

	// Leading comment
	Leading string

	Trailing string
	// Trailing comment

	// Leading comment
	All string // Inline comment
	// Trailing comment

	/*  Another leading comment */
	Block string

	/* Multi
	Line
	Comment
	*/
	MultiLine string
}

type InheritedCommentedStructure struct {
	// CommentedStructure is a commented field, but in typescript it has no place
	// since it is covered by an "extend" clause. Not sure where to put it.
	CommentedStructure
}

/*
  BlockComments are not idiomatic in Go, but can be used
*/
type BlockComment struct {
}

// Constant is just a value
const Constant = "value" // An inline note

// DeprecatedComment is a comment with a deprecation note
//
// Deprecated: this type is no longer used
type DeprecatedComment struct {
}
