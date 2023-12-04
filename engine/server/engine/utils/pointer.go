package utils

func DerefWith[T any](value *T, defaultValue T) T {
	if value == nil {
		return defaultValue
	}
	return *value
}

func MapPointer[T any, U any](value *T, function func(T) U) *U {
	if value == nil {
		return nil
	}
	mapped := function(*value)
	return &mapped
}
