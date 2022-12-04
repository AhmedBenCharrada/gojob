package gojob

func include(errs []error, err error) bool {
	if len(errs) == 0 {
		return false
	}

	for _, e := range errs {
		if e.Error() == err.Error() {
			return true
		}
	}

	return false
}

func orElse[T any](cond bool, val, other T) T {
	if cond {
		return val
	}

	return other
}
