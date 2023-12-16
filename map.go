package gojob

import (
	"sync"
	"sync/atomic"
)

// Map sync map wrapper.
type Map[K comparable, V any] struct {
	count *atomic.Int32
	sm    *sync.Map
}

// Len returns the size of the map.
func (m *Map[K, V]) Len() int {
	if m.count == nil {
		return 0
	}

	return int(m.count.Load())
}

// Put add item to the map.
func (m *Map[K, V]) Put(k K, v V) {
	m.init()
	m.sm.Store(k, v)
	m.count.Add(1)
}

// Get return an item by key.
func (m Map[K, V]) Get(k K) (V, bool) {
	var v V
	if m.sm == nil || m.count == nil {
		return v, false
	}

	if m.Len() == 0 {
		return v, false
	}

	r, ok := m.sm.Load(k)
	if !ok {
		return v, ok
	}

	v, ok = r.(V)
	return v, ok
}

// Remove removes an item from the map.
func (m *Map[K, V]) Remove(k K) {
	if m.sm == nil || m.count == nil {
		return
	}

	if m.Len() == 0 {
		return
	}

	if _, ok := m.sm.LoadAndDelete(k); ok {
		m.count.Add(-1)
	}
}

// Range iterates the the map items and applies the passed callback.
func (m *Map[K, V]) Range(fn func(K, V) error) {
	if m.sm == nil || m.count == nil {
		return
	}
	if m.Len() == 0 {
		return
	}

	m.sm.Range(func(key, value any) bool {
		if err := fn(key.(K), value.(V)); err != nil {
			return false
		}
		return true
	})
}

func (m *Map[K, V]) init() {
	if m.sm == nil {
		m.sm = new(sync.Map)
	}

	if m.count == nil {
		m.count = new(atomic.Int32)
	}
}
