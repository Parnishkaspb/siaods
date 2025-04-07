package B_Tree

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"
	"time"
)

// ---------- CSV Загрузка ----------

func LoadDescriptionsFromCSV(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var descriptions []string
	descIndex := -1
	for i, header := range records[0] {
		if strings.ToLower(header) == "description" {
			descIndex = i
			break
		}
	}
	if descIndex == -1 {
		return nil, os.ErrInvalid
	}

	for _, row := range records[1:] {
		if len(row) > descIndex {
			descriptions = append(descriptions, row[descIndex])
		}
	}
	return descriptions, nil
}

// ---------- Unit Tests ----------

func TestBTree_InsertAndSearch(t *testing.T) {
	tree := NewBTree[int](2)

	values := []int{10, 20, 5, 6, 12, 30, 7, 17}
	for _, v := range values {
		tree.Insert(v)
	}

	for _, v := range values {
		node, i := tree.Root.Search(v)
		if node == nil || node.Keys[i] != v {
			t.Errorf("Search(%d) failed", v)
		}
	}
}

func TestBTree_DeleteAndSearch(t *testing.T) {
	tree := NewBTree[int](2)

	values := []int{10, 20, 5, 6, 12, 30, 7, 17}
	for _, v := range values {
		tree.Insert(v)
	}

	toDelete := []int{6, 10, 30}
	for _, v := range toDelete {
		tree.Delete(v)
		node, i := tree.Root.Search(v)
		if node != nil && i < len(node.Keys) && node.Keys[i] == v {
			t.Errorf("Delete(%d) failed: still found", v)
		}
	}
}

func TestBTree_DuplicateInsert(t *testing.T) {
	tree := NewBTree[int](2)
	tree.Insert(10)
	tree.Insert(10)

	node, idx := tree.Root.Search(10)
	if node == nil || node.Keys[idx] != 10 {
		t.Error("Expected to find 10 after duplicate insert")
	}
}

func TestBTree_PrintAfterInsert(t *testing.T) {
	tree := NewBTree[int](2)

	values := []int{10, 20, 5, 6, 12, 30, 7, 17}
	for _, v := range values {
		tree.Insert(v)
	}

	fmt.Println("Структура дерева после Insert:")
	tree.Root.Print(0)
}

// ---------- Вспомогательная функция ----------

func computeStats(durations []time.Duration) (mean, q1, median, q3 time.Duration) {
	n := len(durations)
	if n == 0 {
		return 0, 0, 0, 0
	}

	sort.Slice(durations, func(i, j int) bool {
		return durations[i] < durations[j]
	})

	var total time.Duration
	for _, d := range durations {
		total += d
	}

	mean = total / time.Duration(n)
	q1 = durations[n/4]
	median = durations[n/2]
	q3 = durations[(3*n)/4]

	return mean, q1, median, q3
}

// ---------- Benchmarks ----------

func BenchmarkBTree_InsertFromCSV(b *testing.B) {
	sizes := []int{1000, 10000, 100000, 1000000}
	values, err := LoadDescriptionsFromCSV("EDAresult.csv")
	if err != nil {
		b.Fatalf("failed to load CSV: %v", err)
	}

	for _, size := range sizes {
		if len(values) < size {
			b.Skipf("CSV does not contain %d entries", size)
			continue
		}

		b.Run(fmt.Sprintf("Insert-%d", size), func(b *testing.B) {
			durations := make([]time.Duration, 0, size)
			tree := NewBTree[string](2)

			startAll := time.Now()
			for _, val := range values[:size] {
				startOp := time.Now()
				tree.Insert(val)
				durations = append(durations, time.Since(startOp))
			}
			total := time.Since(startAll)
			mean, q1, median, q3 := computeStats(durations)
			b.Logf("Insert %d items: total=%v mean=%v q1=%v median=%v q3=%v", size, total, mean, q1, median, q3)
		})
	}
}

func BenchmarkBTree_SearchFromCSV(b *testing.B) {
	sizes := []int{1000, 10000, 100000}
	values, err := LoadDescriptionsFromCSV("EDAresult.csv")
	if err != nil {
		b.Fatalf("failed to load CSV: %v", err)
	}

	for _, size := range sizes {
		if len(values) < size {
			b.Skipf("CSV does not contain %d entries", size)
			continue
		}

		b.Run(fmt.Sprintf("Search-%d", size), func(b *testing.B) {
			tree := NewBTree[string](20)

			for _, val := range values[:size] {
				tree.Insert(val)
			}

			durations := make([]time.Duration, 0, 1000)
			startAll := time.Now()
			for i := 0; i < 1000; i++ {
				v := values[i%size]
				startOp := time.Now()
				_, _ = tree.Root.Search(v)
				durations = append(durations, time.Since(startOp))
			}
			total := time.Since(startAll)
			mean, q1, median, q3 := computeStats(durations)
			b.Logf("Search %d items: total=%v mean=%v q1=%v median=%v q3=%v", size, total, mean, q1, median, q3)
		})
	}
}

func BenchmarkBTree_DeleteFromCSV(b *testing.B) {
	sizes := []int{1000, 10000, 100000}
	values, err := LoadDescriptionsFromCSV("EDAresult.csv")
	if err != nil {
		b.Fatalf("failed to load CSV: %v", err)
	}

	for _, size := range sizes {
		if len(values) < size {
			b.Skipf("CSV does not contain %d entries", size)
			continue
		}

		b.Run(fmt.Sprintf("Delete-%d", size), func(b *testing.B) {
			tree := NewBTree[string](2)

			for _, val := range values[:size] {
				tree.Insert(val)
			}

			durations := make([]time.Duration, 0, size)
			startAll := time.Now()
			for _, val := range values[:size] {
				startOp := time.Now()
				tree.Delete(val)
				durations = append(durations, time.Since(startOp))
			}
			total := time.Since(startAll)
			mean, q1, median, q3 := computeStats(durations)
			b.Logf("Delete %d items: total=%v mean=%v q1=%v median=%v q3=%v", size, total, mean, q1, median, q3)
		})
	}
}
