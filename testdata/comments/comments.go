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
}

/*
  BlockComments are not idiomatic in Go, but can be used
*/
type BlockComment struct {
}

// Constant is just a value
const Constant = "value" // Constant comment
