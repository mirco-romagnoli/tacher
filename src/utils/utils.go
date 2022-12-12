package utils

import "io"

// if 'object' is null then return 'def'
func NonNullOrElse[T comparable](object T, def T) T {
	var zero T
	if object != zero {
		return object
	} else {
		return def
	}
}

// apply the mapping function to each element of the slice and return
// a slice of the results
func Map[T any, R any](list []T, mapFunction func(T) R) (ret []R) {
	ret = make([]R, 0, len(list))
	for _, o := range list {
		ret = append(ret, mapFunction(o))
	}
	return
}

func Find[T any](list []T, compareFunction func(elem T) bool) (int, bool) {
	for i, elem := range list {
		if compareFunction(elem) {
			return i, true
		}
	}
	return -1, false
}

// remove the i-th from the slice, the following elements are moved left
func RemoveIndex[T interface{}](s []T, index int) []T {
	return append(s[:index], s[index+1:]...)
}

// check if the close function returns an error
func CheckClose(reader io.Closer) error {
	if err := reader.Close(); err != nil {
		return err
	}
	return nil
}
