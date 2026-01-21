package codersdk

// User represents a user in the system.
// This comment should be preserved after conversion to a type alias.
type User struct {
	// ID is the unique identifier
	ID string `json:"id"`
	// Name is the user's full name
	Name string `json:"name"`
	// Email is the user's email address
	Email string `json:"email"`
}

// Address contains location information.
type Address struct {
	// Street address
	Street string `json:"street"`
	// City name
	City string `json:"city"`
}

// Profile combines user and address information.
// It extends multiple types.
type Profile struct {
	User
	Address
	// Bio is the user's biography
	Bio string `json:"bio"`
}
