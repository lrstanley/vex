// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package types

import (
	"iter"
	"slices"
	"sync"
	"sync/atomic"
	"time"
)

type OrderedMap[K comparable, V any] struct {
	mu    sync.RWMutex
	store map[K]V
	keys  []K
}

func NewOrderedMap[K comparable, V any]() *OrderedMap[K, V] {
	return &OrderedMap[K, V]{
		store: map[K]V{},
		keys:  []K{},
	}
}

func (o *OrderedMap[K, V]) Get(key K) (V, bool) {
	o.mu.RLock()
	defer o.mu.RUnlock()
	val, exists := o.store[key]
	return val, exists
}

func (o *OrderedMap[K, V]) Exists(key K) bool {
	o.mu.RLock()
	defer o.mu.RUnlock()
	_, exists := o.store[key]
	return exists
}

func (o *OrderedMap[K, V]) Set(k K, v V) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if _, exists := o.store[k]; !exists {
		o.keys = append(o.keys, k)
	}
	o.store[k] = v
}

func (o *OrderedMap[K, V]) Delete(key K) {
	o.mu.Lock()
	defer o.mu.Unlock()
	delete(o.store, key)
	for i, val := range o.keys {
		if val == key {
			o.keys = append(o.keys[:i], o.keys[i+1:]...)
			return
		}
	}
}

func (o *OrderedMap[K, V]) Len() int {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return len(o.keys)
}

func (o *OrderedMap[K, V]) Peek() (K, V) {
	o.mu.RLock()
	defer o.mu.RUnlock()
	if len(o.keys) == 0 {
		var zero K
		var zeroV V
		return zero, zeroV
	}
	return o.keys[len(o.keys)-1], o.store[o.keys[len(o.keys)-1]]
}

func (o *OrderedMap[K, V]) Pop() (K, V) {
	o.mu.RLock()
	if len(o.keys) == 0 {
		var zero K
		var zeroV V
		o.mu.RUnlock()
		return zero, zeroV
	}
	o.mu.RUnlock()

	o.mu.Lock()
	key := o.keys[len(o.keys)-1]
	val := o.store[key]
	o.mu.Unlock()
	o.Delete(key)
	return key, val
}

func (o *OrderedMap[K, V]) Clear() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.store = map[K]V{}
	o.keys = []K{}
}

func (o *OrderedMap[K, V]) ClearAndSet(k K, v V) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.store = map[K]V{k: v}
	o.keys = []K{k}
}

func (o *OrderedMap[K, V]) Keys() []K {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.keys
}

func (o *OrderedMap[K, V]) Values() (values []V) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	for _, v := range o.keys {
		values = append(values, o.store[v])
	}
	return values
}

type atomicExpirable[T any] struct {
	expiresAt time.Time
	value     *T
}

type AtomicExpires[T any] struct {
	v atomic.Pointer[atomicExpirable[T]]
}

func (a *AtomicExpires[T]) Set(v *T, ttl time.Duration) {
	a.v.Store(&atomicExpirable[T]{
		expiresAt: time.Now().Add(ttl),
		value:     v,
	})
}

func (a *AtomicExpires[T]) Get() *T {
	v := a.v.Load()
	if v == nil || v.value == nil {
		return nil
	}
	if time.Now().After(v.expiresAt) {
		a.v.Store(nil)
		return nil
	}
	return v.value
}

type AtomicSlice[T any] struct {
	values atomic.Pointer[[]T]
}

func (a *AtomicSlice[T]) Get() []T {
	v := a.values.Load()
	if v == nil {
		return []T{}
	}
	return *v
}

func (a *AtomicSlice[T]) Iter() iter.Seq2[int, T] {
	return slices.All(a.Get())
}

func (a *AtomicSlice[T]) IterValues() iter.Seq[T] {
	values := a.Get()
	return func(yield func(T) bool) {
		for _, v := range values {
			if !yield(v) {
				return
			}
		}
	}
}

func (a *AtomicSlice[T]) Set(v []T) {
	a.values.Store(&v)
}

func (a *AtomicSlice[T]) Push(v T) {
	a.Set(append(a.Get(), v))
}

func (a *AtomicSlice[T]) Pop() (T, bool) {
	values := a.Get()
	if len(values) == 0 {
		return *new(T), false
	}
	last := values[len(values)-1]
	a.Set(values[:len(values)-1])
	return last, true
}

func (a *AtomicSlice[T]) Len() int {
	return len(a.Get())
}

func (a *AtomicSlice[T]) Clear() {
	a.Set([]T{})
}

func (a *AtomicSlice[T]) Peek() T {
	values := a.Get()
	if len(values) == 0 {
		return *new(T)
	}
	return values[len(values)-1]
}
