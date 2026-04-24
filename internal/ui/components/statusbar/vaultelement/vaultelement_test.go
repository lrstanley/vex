// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package vaultelement

import (
	"fmt"
	"testing"
)

func Test_connectionInsecure(t *testing.T) {
	t.Parallel()
	tests := []struct {
		addr string
		want bool
	}{
		{"http://127.0.0.1:8200", true},
		{"HTTP://localhost:8200", true},
		{"https://127.0.0.1:8200", false},
		{"", false},
		{"%zz", false},
	}
	for _, tt := range tests {
		name := tt.addr
		if name == "" {
			name = "empty"
		}
		t.Run(fmt.Sprintf("%q", name), func(t *testing.T) {
			t.Parallel()
			if got := connectionInsecure(tt.addr); got != tt.want {
				t.Errorf("connectionInsecure(%q) = %v, want %v", tt.addr, got, tt.want)
			}
		})
	}
}
