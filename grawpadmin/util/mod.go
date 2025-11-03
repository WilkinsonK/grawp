package util

import "iter"

func Collect[T, U any](seq iter.Seq[T], f func(T) U) []U {
	var collected []U
	return CollectVar(&collected, seq, f)
}

func CollectVar[T, U any](collected *[]U, seq iter.Seq[T], f func(T) U) []U {
	for c := range Map(seq, f) {
		*collected = append(*collected, c)
	}
	return *collected
}

func Map[T, U any](seq iter.Seq[T], f func(T) U) iter.Seq[U] {
	return func(yield func(U) bool) {
		for a := range seq {
			if !yield(f(a)) {
				return
			}
		}
	}
}

func ForEach[T any](seq iter.Seq[T], f func(T)) {
	for a := range seq {
		f(a)
	}
}
