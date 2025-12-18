// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package utils

import "cmp"

// Clamp returns the value clamped between the min and max values.
func Clamp[T cmp.Ordered](v, minv, maxv T) T {
	if minv > maxv {
		minv, maxv = maxv, minv
	}
	return min(max(v, minv), maxv)
}

// Abs returns the absolute number of the given value.
func Abs[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr | ~float32 | ~float64](v T) T {
	if v < 0 {
		return -v
	}
	return v
}
