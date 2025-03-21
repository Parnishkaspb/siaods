package minhash

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strings"
	"testing"
	"time"
)

// computeStats вычисляет мат. ожидание, первый квартиль (q1), медиану и третий квартиль (q3)
// по срезу измерений длительности операций.
func computeStats(durations []time.Duration) (mean, q1, median, q3 time.Duration) {
	n := len(durations)
	if n == 0 {
		return 0, 0, 0, 0
	}

	// Сортируем длительности по возрастанию
	sort.Slice(durations, func(i, j int) bool {
		return durations[i] < durations[j]
	})

	// Суммируем все значения
	var total time.Duration
	for _, d := range durations {
		total += d
	}
	mean = total / time.Duration(n)

	// Наивное вычисление квартилей:
	q1 = durations[n/4]
	median = durations[n/2]
	q3 = durations[(3*n)/4]

	return mean, q1, median, q3
}

// TestBasicFunctionality проверяет базовую функциональность алгоритма MinHash
func TestBasicFunctionality(t *testing.T) {
	file, err := os.Open("data/articles_100.text")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	id_to_num := map[int]string{}
	sets := [][]string{}

	scanner := bufio.NewScanner(file)
	i := 0
	for scanner.Scan() {
		words := strings.Fields(scanner.Text())
		id_to_num[i] = words[0]
		sets = append(sets, words[1:])
		i++
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	// Создаём MinHash с 16 хеш-функциями и 4 полосами
	mh := NewMinHash(16, 4, sets)
	result := mh.Similarity()

	// Ожидаемые пары с максимальной схожестью (1.0)
	expected := map[[2]string]float64{
		{"t1088", "t5015"}: 1,
		{"t1297", "t4638"}: 1,
		{"t1768", "t5248"}: 1,
		{"t1952", "t3495"}: 1,
		{"t980", "t2023"}:  1,
	}

	for _, rec := range result {
		row1_id := id_to_num[int(rec[0])]
		row2_id := id_to_num[int(rec[1])]

		if sim, exists := expected[[2]string{row1_id, row2_id}]; exists {
			if sim-rec[2] > 0.15 {
				t.Fatalf("Row %s - Row %s: similarity: %v; expected 1", row1_id, row2_id, rec[2])
			} else {
				t.Logf("Row %s - Row %s: similarity: %v", row1_id, row2_id, rec[2])
			}
		} else {
			if rec[2] > 0.95 {
				t.Fatalf("Row %s - Row %s: similarity: %v; expected меньше", row1_id, row2_id, rec[2])
			}
		}
	}
}

// TestAllData проверяет работу алгоритма на наборах данных различного размера
func TestAllData(t *testing.T) {
	files := [][3]string{
		{"100", "data/articles_100.text", "data/articles_100.test"},
		{"1000", "data/articles_1000.text", "data/articles_1000.test"},
		{"2500", "data/articles_2500.text", "data/articles_2500.test"},
	}

	for _, data_set := range files {
		t.Run(fmt.Sprintf("Test file size: %v", data_set[0]), func(t *testing.T) {
			sets, id_to_num := create_sets_from_file(data_set[1])
			expected := create_expected_result_from_file(data_set[2])

			mh := NewMinHash(16, 4, sets)
			result := mh.Similarity()

			for _, rec := range result {
				row1_id := id_to_num[int(rec[0])]
				row2_id := id_to_num[int(rec[1])]

				if sim, exists := expected[[2]string{row1_id, row2_id}]; exists {
					if sim-rec[2] > 0.15 {
						t.Fatalf("Row %s - Row %s: similarity: %v; expected 1", row1_id, row2_id, rec[2])
					} else {
						t.Logf("Row %s - Row %s: similarity: %v", row1_id, row2_id, rec[2])
					}
				} else {
					if rec[2] > 0.95 {
						t.Fatalf("Row %s - Row %s: similarity: %v; expected меньше", row1_id, row2_id, rec[2])
					}
				}
			}
		})
	}
}

// create_sets_from_file читает файл, где каждая строка содержит идентификатор и набор слов,
// и возвращает срез наборов и мапу соответствия индекса идентификатору.
func create_sets_from_file(filepath string) ([][]string, map[int]string) {
	file, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	id_to_num := map[int]string{}
	sets := [][]string{}

	scanner := bufio.NewScanner(file)
	i := 0
	for scanner.Scan() {
		words := strings.Fields(scanner.Text())
		id_to_num[i] = words[0]
		sets = append(sets, words[1:])
		i++
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return sets, id_to_num
}

// create_expected_result_from_file читает файл с ожидаемыми результатами схожести и возвращает мапу,
// где ключ – пара идентификаторов, а значение – ожидаемая схожесть (1.0).
func create_expected_result_from_file(filepath string) map[[2]string]float64 {
	file, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	expected := map[[2]string]float64{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		expected[[2]string{fields[0], fields[1]}] = 1
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return expected
}

// measureTime измеряет время выполнения переданной функции, вычисляет мат. ожидание,
// квартиль, медиану и стандартное отклонение и логирует результаты.
func measureTime(b *testing.B, file string, fn func()) {
	times := []time.Duration{}
	var total time.Duration
	run_cnt := 0

	b.Run(fmt.Sprintf("File-%s", file), func(b *testing.B) {
		b.ResetTimer()
		fn()
		t := b.Elapsed()
		times = append(times, t)
		total += t
		run_cnt++
	})

	mean, q1, median, q3 := computeStats(times)

	// Вычисляем стандартное отклонение
	var variance float64
	for _, t := range times {
		diff := float64(t - mean)
		variance += diff * diff
	}
	variance /= float64(run_cnt)
	stdDev := time.Duration(math.Sqrt(variance))

	b.Logf("File %s: mean=%v, q1=%v, median=%v, q3=%v, stdDev=%v", file, mean, q1, median, q3, stdDev)
}

// BenchmarkFull измеряет время полного цикла работы: создание MinHash и вычисление схожести
func BenchmarkFull(b *testing.B) {
	files := []string{"data/articles_100.text", "data/articles_1000.text", "data/articles_2500.text"}
	for _, file := range files {
		sets, _ := create_sets_from_file(file)
		measureTime(b, file, func() {
			mh := NewMinHash(100, 10, sets)
			mh.Similarity()
		})
	}
}

// BenchmarkCreateMinHashOnly измеряет время создания объекта MinHash без вычисления схожести
func BenchmarkCreateMinHashOnly(b *testing.B) {
	files := []string{"data/articles_100.text", "data/articles_1000.text", "data/articles_2500.text"}
	for _, file := range files {
		sets, _ := create_sets_from_file(file)
		measureTime(b, file, func() {
			NewMinHash(100, 10, sets)
		})
	}
}

// BenchmarkCreatePermutations измеряет время генерации перестановок (хеш-функций)
func BenchmarkCreatePermutations(b *testing.B) {
	mh := &MinHash{Size: 100, Permutations: make([]hash_func, 100)}
	measureTime(b, "create permutations", mh.createPermutations)
}

// BenchmarkGenerateSignature измеряет время генерации сигнатур для каждого набора
func BenchmarkGenerateSignature(b *testing.B) {
	files := []string{"data/articles_100.text", "data/articles_1000.text", "data/articles_2500.text"}
	for _, file := range files {
		sets, _ := create_sets_from_file(file)
		obj := &MinHash{
			Size:         100,
			Permutations: make([]hash_func, 100),
			Signatures:   make([][]uint64, len(sets)),
		}
		obj.createPermutations()

		b.Run(fmt.Sprintf("File-%s", file), func(b *testing.B) {
			durations := make([]time.Duration, 0, len(sets))
			for setID, set := range sets {
				startOp := time.Now()
				obj.Signatures[setID] = obj.generateSignature(set)
				opDuration := time.Since(startOp)
				durations = append(durations, opDuration)
			}
			mean, q1, median, q3 := computeStats(durations)
			b.Logf("GenerateSignature для файла %s: mean=%v, q1=%v, median=%v, q3=%v", file, mean, q1, median, q3)
		})
	}
}

// BenchmarkBucketizeSignatures измеряет время этапа bucketization (распределение сигнатур по бакетам)
func BenchmarkBucketizeSignatures(b *testing.B) {
	files := []string{"data/articles_100.text", "data/articles_1000.text", "data/articles_2500.text"}
	for _, file := range files {
		sets, _ := create_sets_from_file(file)
		obj := &MinHash{
			Size:         100,
			Permutations: make([]hash_func, 100),
			Signatures:   make([][]uint64, len(sets)),
			Buckets:      make(map[int]map[uint64][]int),
			Bands:        10,
		}
		obj.createPermutations()
		for setID, set := range sets {
			obj.Signatures[setID] = obj.generateSignature(set)
		}
		measureTime(b, file, obj.bucketizeSignatures)
	}
}

// BenchmarkFindSimilarPairs измеряет время поиска похожих пар
func BenchmarkFindSimilarPairs(b *testing.B) {
	files := []string{"data/articles_100.text", "data/articles_1000.text", "data/articles_2500.text"}
	for _, file := range files {
		sets, _ := create_sets_from_file(file)
		mh := NewMinHash(100, 10, sets)
		measureTime(b, file, mh.FindSimilarPairsNoReturn)
	}
}

// BenchmarkCalculateSimilarity измеряет время вычисления схожести для найденных пар
func BenchmarkCalculateSimilarity(b *testing.B) {
	files := []string{"data/articles_100.text", "data/articles_1000.text", "data/articles_2500.text"}
	for _, file := range files {
		sets, _ := create_sets_from_file(file)
		mh := NewMinHash(100, 10, sets)
		pairs := mh.FindSimilarPairs()
		measureTime(b, file, func() {
			mh.SimilarityForBench(pairs)
		})
	}
}
