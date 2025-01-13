package ordered_map

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestOrderedMapOperations(t *testing.T) {
	om := New[string, string]()

	// Test Store and Get
	om.Store("key1", "value1")
	om.Store("key2", "value2")
	om.Store("key3", "value3")

	if val, err := om.Get("key1"); err != nil || val != "value1" {
		t.Errorf("expected 'value1', got '%v' (err: %v)", val, err)
	}

	if val, err := om.Get("key2"); err != nil || val != "value2" {
		t.Errorf("expected 'value2', got '%v' (err: %v)", val, err)
	}

	if val, err := om.Get("key3"); err != nil || val != "value3" {
		t.Errorf("expected 'value3', got '%v' (err: %v)", val, err)
	}

	// Test Overwrite
	om.Store("key1", "newValue1")
	if val, err := om.Get("key1"); err != nil || val != "newValue1" {
		t.Errorf("expected 'newValue1', got '%v' (err: %v)", val, err)
	}

	// Test Delete
	if err := om.Delete("key2"); err != nil {
		t.Errorf("unexpected error when deleting 'key2': %v", err)
	}

	if _, err := om.Get("key2"); err == nil {
		t.Errorf("expected error for 'key2' after deletion, got nil")
	}

	// Test GetAll
	all := om.GetAll()
	expected := []struct {
		Key   string
		Value string
	}{
		{"key1", "newValue1"},
		{"key3", "value3"},
	}

	if !reflect.DeepEqual(all, expected) {
		t.Errorf("expected %v, got %v", expected, all)
	}

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

	if !reflect.DeepEqual(allWithData, expectedWithData) {
		t.Errorf("expected %v, got %v", expectedWithData, allWithData)
	}

	// Test WithCapacity
	omWithCapacity := New[string, string](WithCapacity[string, string](10))
	if cap(omWithCapacity.order) != 10 {
		t.Errorf("expected capacity 10, got %d", cap(omWithCapacity.order))
	}
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
