// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package formatter

import (
	"fmt"
	"strings"

	"github.com/goccy/go-yaml"
)

// ToYAML will convert the provided data value into a YAML string, optionally
// masking the values if the mask flag is true.
func ToYAML(data any, mask bool, indent int) string {
	if data == nil {
		return "null"
	}

	indent = max(indent, 2)

	var output any
	if mask {
		output = MaskValue(data)
	} else {
		output = data
	}

	b, err := yaml.MarshalWithOptions(output, yaml.Indent(indent), yaml.UseJSONMarshaler())
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}

	return strings.TrimSuffix(string(b), "\n")
}
