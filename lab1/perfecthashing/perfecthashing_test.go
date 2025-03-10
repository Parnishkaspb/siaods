package perfecthashing

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

			// Проверка GetAllKeys
			assert.ElementsMatch(t, test.expectedKeys, ph.GetAllKeys())
			// Проверка GetAllValues
			assert.ElementsMatch(t, test.expectedValues, ph.GetAllValues())
		})
	}
}

func TestPerfectHashingErrors(t *testing.T) {
	t.Run("keys and values length mismatch", func(t *testing.T) {
		keys := []string{"apple", "banana"}
		values := []any{1}
		_, err := NewPerfectHash(keys, values)
		assert.EqualError(t, err, "keys and values must have the same length")
	})
}

// --- Бенчмарки ---

func BenchmarkNewPerfectHash(b *testing.B) {
	keys := []string{"apple", "banana", "cherry"}
	values := []any{1, 2, 3}

	for i := 0; i < b.N; i++ {
		_, err := NewPerfectHash(keys, values)
		if err != nil {
			b.Fatalf("cannot create PerfectHash: %v", err)
		}
	}
}

func BenchmarkLookup(b *testing.B) {
	keys := []string{"apple", "banana", "cherry"}
	values := []any{1, 2, 3}

	ph, err := NewPerfectHash(keys, values)
	if err != nil {
		b.Fatalf("cannot create PerfectHash: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ph.Lookup("banana")
	}
}

func BenchmarkGetValueByKey(b *testing.B) {
	keys := []string{"apple", "banana", "cherry"}
	values := []any{1, 2, 3}

	ph, err := NewPerfectHash(keys, values)
	if err != nil {
		b.Fatalf("cannot create PerfectHash: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ph.GetValueByKey("banana")
	}
}
