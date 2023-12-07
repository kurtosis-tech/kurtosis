package utils

func MapList[T, U any](data []T, function func(T) U) []U {
	res := make([]U, 0, len(data))
	for _, elem := range data {
		res = append(res, function(elem))
	}
	return res
}

func MapListWithRefStopOnError[T, U any](data []T, function func(*T) (U, error)) ([]U, error) {
	res := make([]U, 0, len(data))
	for _, elem := range data {
		value, err := function(&elem)
		if err != nil {
			return res, err
		}
		res = append(res, value)
	}
	return res, nil
}

func MapListStopOnError[T, U any](data []T, function func(T) (U, error)) ([]U, error) {
	res := make([]U, 0, len(data))
	for _, elem := range data {
		value, err := function(elem)
		if err != nil {
			return res, err
		}
		res = append(res, value)
	}
	return res, nil
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
