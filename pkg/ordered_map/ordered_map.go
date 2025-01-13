package ordered_map

import (
	"errors"
)

var (
	ErrKeyNotFound   = errors.New("key not found")
	ErrInvalidOption = errors.New("invalid option passed to New")
)

type Pair[K comparable, V any] struct {
	Key   K
	Value V
}

type initConfig[K comparable, V any] struct {
	capacity    int
	initialData []Pair[K, V]
}

type InitOption[K comparable, V any] func(config *initConfig[K, V])

func WithCapacity[K comparable, V any](capacity int) InitOption[K, V] {
	return func(c *initConfig[K, V]) {
		c.capacity = capacity
	}
}

func WithInitialData[K comparable, V any](initialData ...Pair[K, V]) InitOption[K, V] {
	return func(c *initConfig[K, V]) {
		c.initialData = initialData
		if c.capacity < len(initialData) {
			c.capacity = len(initialData)
		}
	}
}

// TODO: current implementation is very naive & slow, change to trie-based after server-client hook up
type OrderedMap[K comparable, V any] struct {
	items map[K]V
	order []K
}

func New[K comparable, V any](options ...any) *OrderedMap[K, V] {
	orderedMap := &OrderedMap[K, V]{}

	var config initConfig[K, V]
	for _, untypedOption := range options {
		switch option := untypedOption.(type) {
		case int:
			if len(options) != 1 {
				panic(ErrInvalidOption)
			}
			config.capacity = option

		case InitOption[K, V]:
			option(&config)

		default:
			panic(ErrInvalidOption)
		}
	}

	orderedMap.initialize(config.capacity)
	orderedMap.StorePairs(config.initialData...)

	return orderedMap
}

func (om *OrderedMap[K, V]) initialize(capacity int) {
	om.items = make(map[K]V, capacity)
	om.order = make([]K, 0, capacity)
}

func (om *OrderedMap[K, V]) Store(key K, value V) {
	if _, exists := om.items[key]; !exists {
		om.order = append(om.order, key)
	}
	om.items[key] = value
}

func (om *OrderedMap[K, V]) StorePairs(pairs ...Pair[K, V]) {
	for _, pair := range pairs {
		om.Store(pair.Key, pair.Value)
	}
}

func (om *OrderedMap[K, V]) Delete(key K) error {
	if _, exists := om.items[key]; !exists {
		return ErrKeyNotFound
	}
	delete(om.items, key)

	for i, k := range om.order {
		if k == key {
			om.order = append(om.order[:i], om.order[i+1:]...)
			break
		}
	}
	return nil
}

func (om *OrderedMap[K, V]) Get(key K) (V, error) {
	if value, exists := om.items[key]; exists {
		return value, nil
	}
	var zero V
	return zero, ErrKeyNotFound
}

func (om *OrderedMap[K, V]) GetAll() []struct {
	Key   K
	Value V
} {
	result := make([]struct {
		Key   K
		Value V
	}, 0, len(om.order))
	for _, key := range om.order {
		result = append(result, struct {
			Key   K
			Value V
		}{Key: key, Value: om.items[key]})
	}
	return result
}
