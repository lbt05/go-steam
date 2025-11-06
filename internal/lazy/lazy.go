package lazy

import "sync"

// Value wraps a factory so it only runs once, safely.
type Value[T any] struct {
	once  sync.Once
	init  func() T
	value T
}

// New creates a new Value.
func New[T any](init func() T) *Value[T] {
	return &Value[T]{init: init}
}

// Get returns the value.
func (v *Value[T]) Get() T {
	v.once.Do(func() { v.value = v.init() })
	return v.value
}
