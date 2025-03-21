package perfecthashing

import (
	"errors"
	"math/rand"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// computeStats вычисляет среднее, q1, медиану и q3 для среза длительностей.
func computeStats(durations []time.Duration) (mean, q1, median, q3 time.Duration) {
	n := len(durations)
	if n == 0 {
		return 0, 0, 0, 0
	}

	// Сортируем срез по возрастанию
	sort.Slice(durations, func(i, j int) bool {
		return durations[i] < durations[j]
	})

	// Считаем сумму всех длительностей
	var total time.Duration
	for _, d := range durations {
		total += d
	}
	mean = total / time.Duration(n)

	// Наивное вычисление квартилей
	q1 = durations[n/4]
	median = durations[n/2]
	q3 = durations[(3*n)/4]

	return mean, q1, median, q3
}

func generateTestData(n int) ([]string, []any) {
	keys := make([]string, n)
	values := make([]any, n)
	for i := 0; i < n; i++ {
		keys[i] = "key" + strconv.Itoa(i)
		values[i] = i
	}
	return keys, values
}

func TestPerfectHashing(t *testing.T) {
	tests := []struct {
		name           string
		keys           []string
		values         []any
		lookupKey      string
		expectedFound  bool
		expectedValue  any
		expectedError  error
		expectedKeys   []string
		expectedValues []any
	}{
		{
			name:           "single key-value pair",
			keys:           []string{"apple"},
			values:         []any{1},
			lookupKey:      "apple",
			expectedFound:  true,
			expectedValue:  1,
			expectedError:  nil,
			expectedKeys:   []string{"apple"},
			expectedValues: []any{1},
		},
		{
			name:           "multiple key-value pairs",
			keys:           []string{"apple", "banana", "cherry"},
			values:         []any{1, 2, 3},
			lookupKey:      "banana",
			expectedFound:  true,
			expectedValue:  2,
			expectedError:  nil,
			expectedKeys:   []string{"apple", "banana", "cherry"},
			expectedValues: []any{1, 2, 3},
		},
		{
			name:           "nonexistent key",
			keys:           []string{"apple", "banana", "cherry"},
			values:         []any{1, 2, 3},
			lookupKey:      "nonexistent",
			expectedFound:  false,
			expectedValue:  nil,
			expectedError:  errors.New("key not found"),
			expectedKeys:   []string{"apple", "banana", "cherry"},
			expectedValues: []any{1, 2, 3},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ph, err := NewPerfectHash(test.keys, test.values)
			require.NoError(t, err)

			// Проверка Lookup
			found, err := ph.Lookup(test.lookupKey)
			assert.Equal(t, test.expectedFound, found)
			if test.expectedError != nil {
				assert.EqualError(t, err, test.expectedError.Error())
			} else {
				require.NoError(t, err)
			}

			// Проверка GetValueByKey
			value, err := ph.GetValueByKey(test.lookupKey)
			if test.expectedError != nil {
				assert.EqualError(t, err, test.expectedError.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expectedValue, value)
			}
		})
	}
}

func BenchmarkNewPerfectHashLarge(b *testing.B) {
	keys, values := generateTestData(10000)
	durations := make([]time.Duration, 0, b.N)

	for i := 0; i < b.N; i++ {
		start := time.Now()
		_, err := NewPerfectHash(keys, values)
		if err != nil {
			b.Fatalf("cannot create PerfectHash: %v", err)
		}
		durations = append(durations, time.Since(start))
	}
	mean, q1, median, q3 := computeStats(durations)
	b.Logf("NewPerfectHash: mean=%v, q1=%v, median=%v, q3=%v", mean, q1, median, q3)

	var totalTime time.Duration
	for _, d := range durations {
		totalTime += d
	}
	b.ReportMetric(float64(totalTime.Nanoseconds())/float64(b.N), "ns/op")
}

func BenchmarkLookupLarge(b *testing.B) {
	keys, values := generateTestData(10000)
	ph, err := NewPerfectHash(keys, values)
	if err != nil {
		b.Fatalf("cannot create PerfectHash: %v", err)
	}
	lookupKey := "key" + strconv.Itoa(rand.Intn(10000))
	durations := make([]time.Duration, 0, b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()
		_, _ = ph.Lookup(lookupKey)
		durations = append(durations, time.Since(start))
	}
	mean, q1, median, q3 := computeStats(durations)
	b.Logf("Lookup: mean=%v, q1=%v, median=%v, q3=%v", mean, q1, median, q3)

	var totalTime time.Duration
	for _, d := range durations {
		totalTime += d
	}
	b.ReportMetric(float64(totalTime.Nanoseconds())/float64(b.N), "ns/op")
}

func BenchmarkGetValueByKeyLarge(b *testing.B) {
	keys, values := generateTestData(10000)
	ph, err := NewPerfectHash(keys, values)
	if err != nil {
		b.Fatalf("cannot create PerfectHash: %v", err)
	}
	lookupKey := "key" + strconv.Itoa(rand.Intn(10000))
	durations := make([]time.Duration, 0, b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()
		_, _ = ph.GetValueByKey(lookupKey)
		durations = append(durations, time.Since(start))
	}
	mean, q1, median, q3 := computeStats(durations)
	b.Logf("GetValueByKey: mean=%v, q1=%v, median=%v, q3=%v", mean, q1, median, q3)

	var totalTime time.Duration
	for _, d := range durations {
		totalTime += d
	}
	b.ReportMetric(float64(totalTime.Nanoseconds())/float64(b.N), "ns/op")
}
