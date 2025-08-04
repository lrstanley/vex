// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package formatter

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ToJSON will convert the provided data value into JSON. If mask is true, all
// concrete values will be masked with asterisks.
func ToJSON(data any, mask bool, indent int) string {
	if !mask {
		b, err := json.MarshalIndent(data, "", strings.Repeat(" ", indent))
		if err != nil {
			return fmt.Sprintf("error: %v", err)
		}
		return string(b)
	}

	if data == nil {
		return "null"
	}

	masked := MaskValue(data)
	b, err := json.MarshalIndent(masked, "", strings.Repeat(" ", indent))
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	return string(b)
}
