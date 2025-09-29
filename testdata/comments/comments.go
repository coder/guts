package codersdk

// User holds information for a user
type User struct {
	Name string `json:"name"` // The name of the user
	Age  int    `json:"age"`  // The age of the user
}

// Product represents a product in the system
// This is a multi-line comment
type Product struct {
	ID    int    `json:"id"`          // Product identifier
	Title string `json:"title"`       // Product title
	Price int    `json:"price,omitempty"` // Price in cents
}