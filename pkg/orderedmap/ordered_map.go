package orderedmap

import (
	"errors"
	"sync"
)

var (
	// ErrKeyNotFound is returned when the key is not found
	ErrKeyNotFound = errors.New("key not found")
	// ErrInvalidOption is returned when an invalid option is passed to New
	ErrInvalidOption = errors.New("invalid option passed to New")
)

// Pair is a key-value pair
type Pair[K comparable, V any] struct {
	Key   K
	Value V
}

type initConfig[K comparable, V any] struct {
	capacity    int
	initialData []Pair[K, V]
}

// InitOption is a function type for configuring the OrderedMap during initialization
type InitOption[K comparable, V any] func(config *initConfig[K, V])

// WithCapacity sets the initial capacity of the OrderedMap
func WithCapacity[K comparable, V any](capacity int) InitOption[K, V] {
	return func(c *initConfig[K, V]) {
		c.capacity = capacity
	}
}

// WithInitialData sets the initial data of the OrderedMap
func WithInitialData[K comparable, V any](initialData ...Pair[K, V]) InitOption[K, V] {
	return func(c *initConfig[K, V]) {
		c.initialData = initialData
		if c.capacity < len(initialData) {
			c.capacity = len(initialData)
		}
	}
}

// OrderedMap is a map that maintains the order of keys
type OrderedMap[K comparable, V any] struct {
	mu       sync.RWMutex
	items    map[K]V
	order    []K
	indexMap map[K]int // Tracks the position of each key in the 'order' slice
}

// New creates a new OrderedMap instance
func New[K comparable, V any](options ...any) (*OrderedMap[K, V], error) {
	var config initConfig[K, V]

	for _, untypedOption := range options {
		switch option := untypedOption.(type) {
		case int:
			// If there's more than 1 option, returning an error
			// because we can't parse anything else after an int.
			if len(options) != 1 {
				return nil, ErrInvalidOption
			}
			config.capacity = option

		case InitOption[K, V]:
			option(&config)

		default:
			return nil, ErrInvalidOption
		}
	}

	if config.capacity < 0 {
		config.capacity = 0
	}

	om := &OrderedMap[K, V]{}
	om.initialize(config.capacity)
	om.StorePairs(config.initialData...)

	return om, nil
}

func (om *OrderedMap[K, V]) initialize(capacity int) {
	om.items = make(map[K]V, capacity)
	om.order = make([]K, 0, capacity)
	om.indexMap = make(map[K]int, capacity)
}

// Store stores a key-value pair in the map
func (om *OrderedMap[K, V]) Store(key K, value V) {
	om.mu.Lock()
	defer om.mu.Unlock()

	if _, exists := om.items[key]; !exists {
		// New key: insert it into the end of the order slice
		om.order = append(om.order, key)
		om.indexMap[key] = len(om.order) - 1
	}
	om.items[key] = value
}

// StorePairs stores multiple key-value pairs in the map
func (om *OrderedMap[K, V]) StorePairs(pairs ...Pair[K, V]) {
	om.mu.Lock()
	defer om.mu.Unlock()

	for _, pair := range pairs {
		if _, exists := om.items[pair.Key]; !exists {
			om.order = append(om.order, pair.Key)
			om.indexMap[pair.Key] = len(om.order) - 1
		}
		om.items[pair.Key] = pair.Value
	}
}

// Delete deletes a key from the map
func (om *OrderedMap[K, V]) Delete(key K) error {
	om.mu.Lock()
	defer om.mu.Unlock()

	if _, exists := om.items[key]; !exists {
		return ErrKeyNotFound
	}
	delete(om.items, key)

	// Swap-and-pop removal from the order slice
	index := om.indexMap[key]
	lastIndex := len(om.order) - 1
	if index != lastIndex {
		// Move the last key into the deleted key's position
		lastKey := om.order[lastIndex]
		om.order[index] = lastKey
		om.indexMap[lastKey] = index
	}
	om.order = om.order[:lastIndex]
	delete(om.indexMap, key)

	return nil
}

// Get retrieves a value from the map
func (om *OrderedMap[K, V]) Get(key K) (V, error) {
	om.mu.RLock()
	defer om.mu.RUnlock()

	val, exists := om.items[key]
	if !exists {
		var zero V
		return zero, ErrKeyNotFound
	}
	return val, nil
}

// GetAll retrieves all key-value pairs from the map
func (om *OrderedMap[K, V]) GetAll() []Pair[K, V] {
	om.mu.RLock()
	defer om.mu.RUnlock()

	result := make([]Pair[K, V], 0, len(om.order))
	for _, key := range om.order {
		result = append(result, Pair[K, V]{
			Key:   key,
			Value: om.items[key],
		})
	}
	return result
}
