// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package cache

import (
	"time"

	gc "github.com/Code-Hex/go-generics-cache"
	"github.com/Code-Hex/go-generics-cache/policy/lfu"
)

// New returns a new cache with the given capacity. If capacity is 0, the cache will
// have no limit on the number of items it can hold. Uses LFU eviction policy.
func New[K comparable, V any](capacity int) *gc.Cache[K, V] {
	return gc.New(
		gc.AsLFU[K, V](lfu.WithCapacity(max(capacity, 10))),
		gc.WithJanitorInterval[K, V](10*time.Second),
	)
}
