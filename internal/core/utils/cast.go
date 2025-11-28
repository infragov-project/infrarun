package utils

func SliceCast[T any](slice []any) ([]T, bool) {
	res := make([]T, len(slice))

	for i, elem := range slice {
		casted, ok := elem.(T)

		if !ok {
			return nil, false
		}

		res[i] = casted
	}

	return res, true
}
