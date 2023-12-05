package utils

func DerefWith[T any](value *T, defaultValue T) T {
	if value == nil {
		return defaultValue
	}
	return *value
}

func MapWithRef[T any, U any](value T, function func(*T) U) U {
	return function(&value)
}

func MapPointer[T any, U any](value *T, function func(T) U) *U {
	if value == nil {
		return nil
	}
	mapped := function(*value)
	return &mapped
}

func MapPointerWithError[T any, U any](value *T, function func(T) (U, error)) (*U, error) {
	if value == nil {
		return nil, nil
	}
	mapped, err := function(*value)
	if err != nil {
		return nil, err
	}
	return &mapped, nil
}
