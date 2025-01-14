package ordered_map

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOrderedMapOperations(t *testing.T) {
	om := New[string, string]()

	// Test Store and Get
	om.Store("key1", "value1")
	om.Store("key2", "value2")
	om.Store("key3", "value3")

	val, err := om.Get("key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", val)

	val, err = om.Get("key2")
	assert.NoError(t, err)
	assert.Equal(t, "value2", val)

	val, err = om.Get("key3")
	assert.NoError(t, err)
	assert.Equal(t, "value3", val)

	// Test Overwrite
	om.Store("key1", "newValue1")
	val, err = om.Get("key1")
	assert.NoError(t, err)
	assert.Equal(t, "newValue1", val)

	// Test Delete
	err = om.Delete("key2")
	assert.NoError(t, err)

	_, err = om.Get("key2")
	assert.Error(t, err)

	// Test GetAll
	all := om.GetAll()
	expected := []struct {
		Key   string
		Value string
	}{
		{"key1", "newValue1"},
		{"key3", "value3"},
	}
	assert.Equal(t, expected, all)

	// Test WithInitialData
	omWithData := New[string, string](WithInitialData(Pair[string, string]{"a", "1"}, Pair[string, string]{"b", "2"}))
	allWithData := omWithData.GetAll()
	expectedWithData := []struct {
		Key   string
		Value string
	}{
		{"a", "1"},
		{"b", "2"},
	}
	assert.Equal(t, expectedWithData, allWithData)

	// Test WithCapacity
	omWithCapacity := New[string, string](WithCapacity[string, string](10))
	assert.Equal(t, 10, cap(omWithCapacity.order))
}

func BenchmarkOrderedMap(b *testing.B) {
	om := New[int, int](WithCapacity[int, int](b.N))

	b.Run("Store", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			om.Store(i, i*2)
		}
	})

	b.Run("Get", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = om.Get(i)
		}
	})

	b.Run("Delete", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = om.Delete(i)
		}
	})
}

func TestPerformanceOrderedMap(t *testing.T) {
	sizes := []int{100, 1000, 10000, 100000}

	for _, size := range sizes {
		t.Run(fmt.Sprintf("Size_%d", size), func(t *testing.T) {
			om := New[int, int](WithCapacity[int, int](size))

			start := time.Now()
			for i := 0; i < size; i++ {
				om.Store(i, i*2)
			}
			storeDuration := time.Since(start)

			start = time.Now()
			for i := 0; i < size; i++ {
				_, _ = om.Get(i)
			}
			getDuration := time.Since(start)

			start = time.Now()
			for i := 0; i < size; i++ {
				_ = om.Delete(i)
			}
			deleteDuration := time.Since(start)

			t.Logf("Size: %d, Store Duration: %v, Get Duration: %v, Delete Duration: %v", size, storeDuration, getDuration, deleteDuration)
		})
	}
}
