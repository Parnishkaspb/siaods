package minhash

import (
	"testing"
)

func TestMinHash(t *testing.T) {
	vectorSize := 10
	numHashes := 200

	minHash := NewMinHash(numHashes, vectorSize)

	tests := []struct {
		name      string
		vector1   []int
		vector2   []int
		expectMin float64
		expectMax float64
	}{
		{
			name:      "identical_vectors",
			vector1:   []int{0, 1, 0, 1, 0, 1, 0, 1, 0, 1},
			vector2:   []int{0, 1, 0, 1, 0, 1, 0, 1, 0, 1},
			expectMin: 0.9,
			expectMax: 1.0,
		},
		{
			name:      "slightly_different_vectors",
			vector1:   []int{0, 1, 0, 1, 0, 1, 0, 1, 0, 1},
			vector2:   []int{0, 1, 0, 1, 0, 1, 0, 1, 0, 0},
			expectMin: 0.5,
			expectMax: 0.9,
		},
		{
			name:      "completely_different_vectors",
			vector1:   []int{1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			vector2:   []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			expectMin: 0.0,
			expectMax: 0.2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash1 := minHash.Compute(tt.vector1)
			hash2 := minHash.Compute(tt.vector2)
			similarity := Similarity(hash1, hash2)

			if similarity < tt.expectMin || similarity > tt.expectMax {
				t.Errorf("Test %s failed: expected similarity between %.2f and %.2f, got %.2f",
					tt.name, tt.expectMin, tt.expectMax, similarity)
			}
		})
	}
}

// Бенчмарк для Compute
func BenchmarkMinHashCompute(b *testing.B) {
	vectorSize := 1000
	numHashes := 200
	minHash := NewMinHash(numHashes, vectorSize)

	vector := make([]int, vectorSize)
	for i := 0; i < vectorSize; i += 10 {
		vector[i] = 1
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		minHash.Compute(vector)
	}
}

// Бенчмарк для Similarity (вычисление схожести)
func BenchmarkSimilarity(b *testing.B) {
	vectorSize := 1000
	numHashes := 200
	minHash := NewMinHash(numHashes, vectorSize)

	vector1 := make([]int, vectorSize)
	vector2 := make([]int, vectorSize)

	for i := 0; i < vectorSize; i += 10 {
		vector1[i] = 1
		vector2[i] = 1
	}

	hash1 := minHash.Compute(vector1)
	hash2 := minHash.Compute(vector2)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Similarity(hash1, hash2)
	}
}
