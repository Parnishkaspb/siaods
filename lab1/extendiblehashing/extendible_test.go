package extendiblehashing

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

//func TestExtendibleHashing(t *testing.T) {
//	tests := []struct {
//		name           string
//		keysToInsert   []string
//		valuesToInsert []any
//		keysToFind     []string
//		expectedFound  map[string]bool
//		initialGD      int
//		maxSize        int
//	}{
//		{
//			name:           "InitialState",
//			keysToInsert:   []string{},
//			valuesToInsert: []any{},
//			keysToFind:     []string{},
//			expectedFound:  map[string]bool{},
//			initialGD:      2,
//			maxSize:        2,
//		},
//		{
//			name:           "InsertAndFind",
//			keysToInsert:   []string{"apple", "banana", "cherry"},
//			valuesToInsert: []any{1, 2, 3},
//			keysToFind:     []string{"apple", "banana", "cherry", "grape"},
//			expectedFound:  map[string]bool{"apple": true, "banana": true, "cherry": true, "grape": false},
//			initialGD:      2,
//			maxSize:        2,
//		},
//		{
//			name:           "CollisionHandling",
//			keysToInsert:   []string{"key1", "key2", "key3", "key4", "key5"},
//			valuesToInsert: []any{10, 20, 30, 40, 50},
//			keysToFind:     []string{"key1", "key2", "key3", "key4", "key5"},
//			expectedFound:  map[string]bool{"key1": true, "key2": true, "key3": true, "key4": true, "key5": true},
//			initialGD:      2,
//			maxSize:        2,
//		},
//	}
//
//	for _, test := range tests {
//		t.Run(test.name, func(t *testing.T) {
//			eh := NewExtendibleHash(test.initialGD, test.maxSize)
//
//			for i, key := range test.keysToInsert {
//				err := eh.InsertKey(key, test.valuesToInsert[i])
//				assert.NoError(t, err, "Ошибка при вставке ключа %s", key)
//			}
//
//			for key, expected := range test.expectedFound {
//				found, _ := eh.Lookup(key)
//				assert.Equal(t, expected, found, "Ключ %s должен находиться: %v", key, expected)
//			}
//		})
//	}
//}

func generateLargeTestData(n int) ([]string, []any, map[string]bool) {
	keys := make([]string, n)
	values := make([]any, n)
	expectedFound := make(map[string]bool)

	for i := 0; i < n; i++ {
		keys[i] = fmt.Sprintf("key%d", i)
		values[i] = i
		expectedFound[keys[i]] = true
	}

	// Добавляем один отсутствующий ключ для проверки поиска
	expectedFound["missing_key"] = false

	return keys, values, expectedFound
}

func TestExtendibleHashing(t *testing.T) {
	largeKeys, largeValues, expectedLargeFound := generateLargeTestData(10000)

	tests := []struct {
		name           string
		keysToInsert   []string
		valuesToInsert []any
		keysToFind     []string
		expectedFound  map[string]bool
		initialGD      int
		maxSize        int
	}{
		{
			name:           "InitialState",
			keysToInsert:   []string{},
			valuesToInsert: []any{},
			keysToFind:     []string{},
			expectedFound:  map[string]bool{},
			initialGD:      2,
			maxSize:        2,
		},
		{
			name:           "InsertAndFind_Large",
			keysToInsert:   largeKeys,
			valuesToInsert: largeValues,
			keysToFind:     append(largeKeys[:100], "missing_key"),
			expectedFound:  expectedLargeFound,
			initialGD:      2,
			maxSize:        2,
		},
		{
			name:           "CollisionHandling_Large",
			keysToInsert:   largeKeys,
			valuesToInsert: largeValues,
			keysToFind:     largeKeys[:100],
			expectedFound:  expectedLargeFound,
			initialGD:      2,
			maxSize:        4,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			eh := NewExtendibleHash(test.initialGD, test.maxSize)

			for i, key := range test.keysToInsert {
				err := eh.InsertKey(key, test.valuesToInsert[i])
				assert.NoError(t, err, "Ошибка при вставке ключа %s", key)
			}

			for _, key := range test.keysToFind {
				found, _ := eh.Lookup(key)
				assert.Equal(t, test.expectedFound[key], found, "Ключ %s должен находиться: %v", key, test.expectedFound[key])
			}
		})
	}
}

// Бенчмарки
func BenchmarkExtendibleHashInsert(b *testing.B) {
	benchmarks := []struct {
		name           string
		keysToInsert   []string
		valuesToInsert []any
		initialGD      int
		maxSize        int
	}{
		{
			name:           "InitialState",
			keysToInsert:   []string{},
			valuesToInsert: []any{},
			initialGD:      2,
			maxSize:        2,
		},
		{
			name:           "InsertAndFind",
			keysToInsert:   []string{"apple", "banana", "cherry"},
			valuesToInsert: []any{1, 2, 3},
			initialGD:      2,
			maxSize:        2,
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			eh := NewExtendibleHash(bm.initialGD, bm.maxSize)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for j, key := range bm.keysToInsert {
					eh.InsertKey(key, bm.valuesToInsert[j])
				}
			}
		})
	}
}

func BenchmarkExtendibleHashFind(b *testing.B) {
	keys := []string{"apple", "banana", "cherry", "grape"}
	values := []any{1, 2, 3}
	eh := NewExtendibleHash(2, 2)
	for i := range values {
		eh.InsertKey(keys[i], values[i])
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = eh.Lookup(keys[i%len(keys)])
	}
}

func BenchmarkExtendibleHashDelete(b *testing.B) {
	keys := []string{"apple", "banana", "cherry"}
	values := []any{1, 2, 3}
	eh := NewExtendibleHash(2, 2)
	for i := range values {
		eh.InsertKey(keys[i], values[i])
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		eh.DeleteKey(keys[i%len(keys)])
	}
}
