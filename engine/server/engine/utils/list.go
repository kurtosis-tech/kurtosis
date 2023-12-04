package utils

func MapList[T, U any](data []T, function func(T) U) []U {
	res := make([]U, 0, len(data))
	for _, e := range data {
		res = append(res, function(e))
	}
	return res
}

func FilterListNils[T any](data []*T) []T {
	filterList := make([]T, 0)
	for _, elem := range data {
		if elem != nil {
			filterList = append(filterList, *elem)
		}
	}
	return filterList
}
