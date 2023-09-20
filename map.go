package gojob

import (
	"fmt"
	"sync"
)

// Map concurrent generic map.
type Map[T Object] struct {
	m     map[string]T
	mutex sync.RWMutex
}

// Object define the base interface for the map values.
type Object interface {
	ID() string
}

func (sm *Map[T]) Get(id string) (T, bool) {
	sm.mutex.RLock()
	obj, ok := sm.m[id]
	sm.mutex.RUnlock()
	return obj, ok
}

func (sm *Map[T]) Push(obj T) error {
	if len(obj.ID()) == 0 {
		return fmt.Errorf("missing ID")
	}

	sm.mutex.Lock()
	sm.m[obj.ID()] = obj
	sm.mutex.Unlock()
	return nil
}
