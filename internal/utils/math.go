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
