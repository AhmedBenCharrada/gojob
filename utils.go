package gojob

func orElse[T any](cond bool, val, other T) T {
	if cond {
		return val
	}

	return other
}
