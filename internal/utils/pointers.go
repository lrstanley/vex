// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package utils

import (
	"iter"
)

// Ptr returns a pointer to the given value.
func Ptr[T any](v T) *T {
	return &v
}

// Ptrs returns a slice of pointers to the given values.
func Ptrs[T any](v ...T) (out []*T) {
	for _, v := range v {
		out = append(out, &v)
	}
	return out
}

// IterPtrs returns an iterator of pointers to the given values.
func IterPtrs[T any](in iter.Seq[T]) iter.Seq[*T] {
	return func(yield func(*T) bool) {
		in(func(v T) bool {
			return yield(Ptr(v))
		})
	}
}
