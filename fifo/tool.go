package fifo

import "slices"

func appendUnique[T comparable](a []T, x T) []T {
	if slices.Index(a, x) < 0 {
		return append(a, x)
	}
	return a
}
