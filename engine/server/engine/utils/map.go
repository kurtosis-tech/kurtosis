package utils

func MapMapValues[T, U any, K comparable](data map[K]T, function func(T) U) map[K]U {
	mappedMap := make(map[K]U, len(data))
	for key, value := range data {
		mappedMap[key] = function(value)
	}

	return mappedMap
}

func NewMapFromList[U any, K comparable](data []K, function func(K) U) map[K]U {
	mappedMap := make(map[K]U, len(data))
	for _, key := range data {
		mappedMap[key] = function(key)
	}

	return mappedMap
}
