// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package types

import (
	"sync"

	"github.com/segmentio/ksuid"
)

func init() { //nolint:gochecknoinits
	ksuid.SetRand(ksuid.FastRander)
}

type uuid struct {
	once  sync.Once
	value string
}

func (u *uuid) String() string {
	u.once.Do(func() {
		u.value = ksuid.New().String()
	})
	return u.value
}
