package bindings

// ToStrings works for any type where the base type is a string.
func ToStrings[T ~string](a []T) []string {
	tmp := make([]string, 0, len(a))
	for _, v := range a {
		tmp = append(tmp, string(v))
	}
	return tmp
}

// List is a helper function to reduce boilerplate when converting slices of
// database types to slices of codersdk types.
// Only works if the function takes a single argument.
func List[F any, T any](list []F, convert func(F) T) []T {
	return ListLazy(convert)(list)
}

// ListLazy returns the converter function for a list, but does not eval
// the input. Helpful for combining the Map and the List functions.
func ListLazy[F any, T any](convert func(F) T) func(list []F) []T {
	return func(list []F) []T {
		into := make([]T, 0, len(list))
		for _, item := range list {
			into = append(into, convert(item))
		}
		return into
	}
}

func ToInt[T ~int](a T) int {
	return int(a)
}
