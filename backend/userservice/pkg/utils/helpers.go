package utils

func VPtr[T any](v T) *T {
	return &v
}
