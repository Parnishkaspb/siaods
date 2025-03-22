package extendablehash

import (
	"fmt"
	"sort"
	"testing"
	"time"
)

func computeStats(durations []time.Duration) (mean, q1, median, q3 time.Duration) {
	n := len(durations)
	if n == 0 {
		return 0, 0, 0, 0
	}

	// Сортируем по возрастанию
	sort.Slice(durations, func(i, j int) bool {
		return durations[i] < durations[j]
	})

	// Сумма всех измерений
	var total time.Duration
	for _, d := range durations {
		total += d
	}

	// Среднее (мат. ожидание)
	mean = total / time.Duration(n)

	// Квартили (наивный вариант для примера)
	q1 = durations[n/4]
	median = durations[n/2]
	q3 = durations[(3*n)/4]

	return mean, q1, median, q3
}

func TestExtendableHash(t *testing.T) {
	t.Run("smoke test", func(t *testing.T) {
		hash := NewExtendableHashTable()

		hash.Insert("1", "value1")
		hash.Insert("2", "value2")
		hash.Insert("3", "value3")
		hash.Insert("4", "value4")
		hash.Insert("5", "value5")

		hash.Insert("6", "value6")
		hash.Insert("7", "value7")

		tests := []struct {
			key      string
			expected string
		}{
			{"1", "value1"},
			{"2", "value2"},
			{"3", "value3"},
			{"4", "value4"},
			{"5", "value5"},
			{"6", "value6"},
			{"7", "value7"},
		}

		for _, test := range tests {
			value, exists := hash.Get(test.key)
			if !exists || value != test.expected {
				t.Errorf("Expected %v for key %v, got %v", test.expected, test.key, value)
			}
		}

		if _, exists := hash.Get("3123"); exists {
			t.Error("Expected not to find a nonexistent key")
		}
	})

	t.Run("correct values", func(t *testing.T) {
		const size = 800

		data := make(map[string]string, size)
		keys := make([]string, size)

		p_hash := NewExtendableHashTable()

		for i := 0; i < size; i++ {
			key := fmt.Sprintf("key%09d", i)
			value := fmt.Sprintf("value%d", i)
			data[key] = value
			keys[i] = key
			p_hash.Insert(key, value)
		}

		for idx := range p_hash.Buckets {
			bucket := p_hash.loadBucketFromFile(idx)
			if len(bucket.Items) > BUCKET_SIZE {
				t.Errorf("Bucket #%b len: %v\n", idx, len(bucket.Items))
			}
		}

		for key, value := range data {
			if got, exists := p_hash.Get(key); !exists || got != value {
				t.Errorf("Key %v: expected %v, got %v", key, value, got)
			}
		}
	})
}

func BenchmarkInsert(b *testing.B) {
	//sizes := []int{200, 400, 800}
	sizes := []int{1000, 10000, 100000}

	for _, size := range sizes {
		data := make(map[string]string, size)
		for i := 0; i < size; i++ {
			data[fmt.Sprintf("key%09d", i)] = fmt.Sprintf("value%d", i)
		}

		b.Run(fmt.Sprintf("Insert-%d", size), func(b *testing.B) {
			eh := NewExtendableHashTable()
			durations := make([]time.Duration, 0, size)

			// Измеряем время каждой операции
			startAll := time.Now()
			for key, value := range data {
				startOp := time.Now()
				eh.Insert(key, value)
				durations = append(durations, time.Since(startOp))
			}
			totalElapsed := time.Since(startAll)

			mean, q1, median, q3 := computeStats(durations)

			b.Logf("Insert %d items: total=%v mean=%v q1=%v median=%v q3=%v",
				size, totalElapsed, mean, q1, median, q3)
		})
	}
}

func BenchmarkGet(b *testing.B) {
	//sizes := []int{200, 400, 800}
	sizes := []int{1000, 10000, 100000}

	for _, size := range sizes {
		// Готовим данные
		eh := NewExtendableHashTable()
		keys := make([]string, size)

		for i := 0; i < size; i++ {
			key := fmt.Sprintf("key%09d", i)
			keys[i] = key
			eh.Insert(key, fmt.Sprintf("value%d", i))
		}

		b.Run(fmt.Sprintf("Get-%d", size), func(b *testing.B) {
			durations := make([]time.Duration, 0, 10000)

			startAll := time.Now()
			for i := 0; i < 10000; i++ {
				key := keys[i%size]
				startOp := time.Now()
				_, _ = eh.Get(key)
				durations = append(durations, time.Since(startOp))
			}
			totalElapsed := time.Since(startAll)

			mean, q1, median, q3 := computeStats(durations)

			b.Logf("Get %d items (10k gets): total=%v mean=%v q1=%v median=%v q3=%v",
				size, totalElapsed, mean, q1, median, q3)
		})
	}
}
