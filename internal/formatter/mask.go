// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package formatter

import (
	"fmt"
	"reflect"
)

const MaskReplacementValue = "***"

// MaskValue recursively masks concrete values in the data structure.
func MaskValue(v any) any {
	if v == nil {
		return nil
	}

	val := reflect.ValueOf(v)
	switch val.Kind() { //nolint:exhaustive
	case reflect.Map:
		result := make(map[string]any)
		for _, key := range val.MapKeys() {
			keyStr := fmt.Sprintf("%v", key.Interface())
			result[keyStr] = MaskValue(val.MapIndex(key).Interface())
		}
		return result
	case reflect.Slice, reflect.Array:
		result := make([]any, val.Len())
		for i := range val.Len() {
			result[i] = MaskValue(val.Index(i).Interface())
		}
		return result
	case reflect.Struct:
		result := make(map[string]any)
		typ := val.Type()
		for i := range val.NumField() {
			field := val.Field(i)
			fieldType := typ.Field(i)
			fieldName := fieldType.Name
			if jsonTag := fieldType.Tag.Get("json"); jsonTag != "" {
				if jsonTag == "-" {
					continue // Skip fields with json:"-" tag
				}
				fieldName = jsonTag
			}
			result[fieldName] = MaskValue(field.Interface())
		}
		return result
	case reflect.Ptr:
		if val.IsNil() {
			return nil
		}
		return MaskValue(val.Elem().Interface())
	default:
		return MaskReplacementValue
	}
}
