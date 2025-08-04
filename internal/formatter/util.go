// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package formatter

// IsFlatValue returns true when the map values provided is a flat value. Some examples:
//   - string
//   - int/float/etc
//   - bool
//   - nil
func IsFlatValue(data map[string]any) bool {
	if len(data) == 0 {
		return false
	}
	return false // TODO
}
