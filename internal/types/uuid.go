// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package types

import (
	"sync"

	uuidv5 "github.com/gofrs/uuid/v5"
)

type uuid struct {
	once  sync.Once
	value string
}

func (u *uuid) String() string {
	u.once.Do(func() {
		u.value = uuidv5.Must(uuidv5.NewV4()).String()
	})
	return u.value
}
