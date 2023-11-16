package utils

func DerefWith[T any](value *T, defaultValue T) T {
	if value == nil {
		return defaultValue
	}
	return *value
}
