package store

import (
	"sync"
)

// Store is a generic store that can handle all types all data.
type Store[K comparable, V any] struct {
	data map[K]V
	mux  sync.RWMutex
}

// New returns new Store.
func New[K comparable, V any](data map[K]V) *Store[K, V] {
	if data == nil {
		data = make(map[K]V)
	}

	return &Store[K, V]{
		data: data,
	}
}

// Get ...
func (s *Store[K, V]) Get(key K) (value V, ok bool) {
	s.mux.RLock()
	defer s.mux.RUnlock()

	value, ok = s.data[key]
	return value, ok
}

// Set ...
func (s *Store[K, V]) Set(key K, value V) {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.data[key] = value
}

// Data returns all data that stores.
func (s *Store[K, V]) Data() map[K]V {
	s.mux.RLock()
	defer s.mux.RUnlock()

	return s.data
}
