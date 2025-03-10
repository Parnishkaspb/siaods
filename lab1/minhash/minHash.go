package minhash

import (
	"math/rand"
	"time"
)

type MinHash struct {
	numHashes    int
	permutations [][]int
}

func NewMinHash(numHashes, vectorSize int) *MinHash {
	rand.Seed(time.Now().UnixNano())

	permutations := make([][]int, numHashes)
	for i := range permutations {
		permutations[i] = rand.Perm(vectorSize)
	}

	return &MinHash{
		numHashes:    numHashes,
		permutations: permutations,
	}
}

// Compute вычисляет MinHash для вектора
func (m *MinHash) Compute(vector []int) []int {
	hashes := make([]int, m.numHashes)
	for i := 0; i < m.numHashes; i++ {
		hashes[i] = m.computeHash(vector, m.permutations[i])
	}
	return hashes
}

// computeHash вычисляет хэш для одного хэш-функции
func (m *MinHash) computeHash(vector, permutation []int) int {
	for _, p := range permutation {
		if vector[p] == 1 {
			return p
		}
	}
	return -1
}

// Similarity вычисляет схожесть MinHash-векторов
func Similarity(hash1, hash2 []int) float64 {
	matches := 0
	for i := range hash1 {
		if hash1[i] == hash2[i] {
			matches++
		}
	}
	return float64(matches) / float64(len(hash1))
}
