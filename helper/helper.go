package helper

// Ptr Returns a pointer from a literal value in order to avoid creating a new variable
func Ptr[T any](v T) *T {
	return &v
}
