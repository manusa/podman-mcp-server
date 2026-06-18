package api

// Ptr returns a pointer to the given value. Useful for setting optional
// pointer fields such as the *bool tool annotation hints.
func Ptr[T any](v T) *T {
	return &v
}
