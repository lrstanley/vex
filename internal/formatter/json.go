// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package formatter

import (
	"encoding/json"
	"fmt"
	"strings"
)

// JSONMask will convert the provided data value into JSON, with all concrete values
// masked with asterisks.
func JSONMask(data any, indent int) string {
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
