// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package types

import (
	"github.com/gofrs/uuid/v5"
)

// UUID generates a random UUID.
func UUID() string {
	return uuid.Must(uuid.NewV4()).String()
}
