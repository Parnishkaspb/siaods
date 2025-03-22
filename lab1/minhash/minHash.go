package minhash

import (
	"math"
	"math/rand"
	"os"

	"hash/fnv"
)

type SetValueType interface {
	uint64 | string
}

type Set[v_type SetValueType] struct {
	Values map[v_type]bool
}

func hash(elem string) uint64 {
	h := fnv.New64()
	h.Write([]byte(elem))
	return h.Sum64()
}

func hashBand(band []uint64) uint64 {
	h := fnv.New64()
	for _, val := range band {
		var b [8]byte
		for i := 0; i < 8; i++ {
			b[i] = byte(val >> (i * 8))
		}
		h.Write(b[:])
	}
	return h.Sum64()
}

const (
	mersennePrime = (1 << 61) - 1
)

type hash_func func(uint64) uint64

func create_permutation(a, b uint64) hash_func {
	return func(x uint64) uint64 {
		return (a*x + b%mersennePrime)
	}
}

type MinHash struct {
	Permutations []hash_func
	Signatures   [][]uint64
	Size         int
	Buckets      map[int]map[uint64][]int
	Bands        int
}

func NewMinHash(hash_func_cnt, bands int, sets [][]string) *MinHash {
	sets_len := len(sets)
	if sets_len <= 1 {
		os.Exit(1)
	}

	obj := &MinHash{
		Size:         hash_func_cnt,
		Permutations: make([]hash_func, hash_func_cnt),
		Signatures:   make([][]uint64, len(sets)),
		Buckets:      make(map[int]map[uint64][]int),
		Bands:        bands,
	}

	obj.createPermutations()

	for set_id, set := range sets {
		obj.Signatures[set_id] = obj.generateSignature(set)
	}

	obj.bucketizeSignatures()

	return obj
}

func (mh *MinHash) createPermutations() {
	for i := 0; i < mh.Size; i++ {
		a := rand.Uint64()%mersennePrime + 1
		b := rand.Uint64() % mersennePrime
		mh.Permutations[i] = create_permutation(a, b)
	}
}

func (mh *MinHash) generateSignature(set []string) []uint64 {
	signature := make([]uint64, mh.Size)
	for i := range signature {
		signature[i] = math.MaxUint64
	}

	for _, elem := range set {
		hashVal := hash(elem)
		for i, perm := range mh.Permutations {
			minHash := perm(hashVal)
			if minHash < signature[i] {
				signature[i] = minHash
			}
		}
	}
	return signature
}

func (mh *MinHash) bucketizeSignatures() {
	rows := mh.Size / mh.Bands

	for band := 0; band < mh.Bands; band++ {
		mh.Buckets[band] = make(map[uint64][]int)
		for i, sig := range mh.Signatures {
			start := band * rows
			end := start + rows
			bandSig := hashBand(sig[start:end])
			mh.Buckets[band][bandSig] = append(mh.Buckets[band][bandSig], i)
		}
	}
}

func (mh *MinHash) FindSimilarPairs() [][]int {
	var pairs [][]int
	seen := make(map[[2]int]bool)

	for _, bandBuckets := range mh.Buckets {
		for _, candidates := range bandBuckets {
			cand_len := len(candidates)
			if cand_len < 2 {
				continue
			}

			for i := 0; i < len(candidates); i++ {
				for j := i + 1; j < len(candidates); j++ {
					pair := [2]int{candidates[i], candidates[j]}
					if !seen[pair] {
						seen[pair] = true
						pairs = append(pairs, []int{candidates[i], candidates[j]})
					}
				}
			}
		}
	}
	return pairs
}

func (mh *MinHash) FindSimilarPairsNoReturn() {
	mh.FindSimilarPairs()
}

func (mh *MinHash) Similarity() [][]float64 {
	pairs := mh.FindSimilarPairs()

	var result [][]float64

	for _, pair := range pairs {
		intersection := 0

		for k := 0; k < mh.Size; k++ {
			if mh.Signatures[pair[0]][k] == mh.Signatures[pair[1]][k] {
				intersection++
			}
		}

		result = append(result, []float64{float64(pair[0]), float64(pair[1]), float64(intersection) / float64(mh.Size)})
	}

	return result
}

func (mh *MinHash) SimilarityForBench(pairs [][]int) {
	var result [][]float64

	for _, pair := range pairs {
		intersection := 0

		for k := 0; k < mh.Size; k++ {
			if mh.Signatures[pair[0]][k] == mh.Signatures[pair[1]][k] {
				intersection++
			}
		}

		result = append(result, []float64{float64(pair[0]), float64(pair[1]), float64(intersection) / float64(mh.Size)})
	}
}
