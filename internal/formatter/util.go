// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package formatter

import (
	"reflect"
	"slices"
)

var flatTypes = []reflect.Kind{
	reflect.Bool,
	reflect.Int,
	reflect.Int8,
	reflect.Int16,
	reflect.Int32,
	reflect.Int64,
	reflect.Uint,
	reflect.Uint8,
	reflect.Uint16,
	reflect.Uint32,
	reflect.Uint64,
	reflect.Uintptr,
	reflect.Float32,
	reflect.Float64,
	reflect.Complex64,
	reflect.Complex128,
	reflect.String,
}

// IsFlatValue returns true when the map values provided is a flat value (can be
// rendered into a 1 dimensional list).
func IsFlatValue[T comparable](data map[T]any) bool {
	if len(data) == 0 {
		return false
	}
	for _, value := range data {
		if value == nil {
			continue
		}
		val := reflect.Indirect(reflect.ValueOf(value))
		if !slices.Contains(flatTypes, val.Kind()) {
			return false
		}
	}
	return true
}
