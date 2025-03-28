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
	var descIndex int = -1

	// ищем индекс колонки "Description"
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

// --- Unit Tests ---
func TestInsertAndSearch(t *testing.T) {
	root := &TreeNode[int]{Val: 10}

	values := []int{5, 15, 3, 7, 13, 20}
	for _, v := range values {
		if err := root.Insert(v); err != nil {
			t.Errorf("Insert(%d) failed: %v", v, err)
		}
	}

	for _, v := range values {
		node, err := root.Search(v)
		if err != nil || node == nil || node.Val != v {
			t.Errorf("Search(%d) failed, got %v", v, node)
		}
	}
}

func TestInsertDuplicate(t *testing.T) {
	root := &TreeNode[int]{Val: 10}
	if err := root.Insert(10); err == nil {
		t.Error("Expected error when inserting duplicate")
	}
}

func TestDeleteLeafNode(t *testing.T) {
	root := &TreeNode[int]{Val: 10}
	root.Insert(5)
	root.Insert(15)
	root.Insert(3) // leaf

	root.Delete(3)

	if _, err := root.Search(3); err == nil {
		t.Error("Expected error after deleting leaf node")
	}
}

func TestDeleteNodeWithOneChild(t *testing.T) {
	root := &TreeNode[int]{Val: 10}
	root.Insert(5)
	root.Insert(3) // left of 5

	root.Delete(5)

	if _, err := root.Search(5); err == nil {
		t.Error("Expected error after deleting node with one child")
	}
}

func TestDeleteNodeWithTwoChildren(t *testing.T) {
	root := &TreeNode[int]{Val: 10}
	root.Insert(5)
	root.Insert(15)
	root.Insert(3)
	root.Insert(7)

	root.Delete(5)

	if _, err := root.Search(5); err == nil {
		t.Error("Expected error after deleting node with two children")
	}
}

func TestFindMinMax(t *testing.T) {
	root := &TreeNode[int]{Val: 10}
	values := []int{5, 15, 3, 7, 13, 20}
	for _, v := range values {
		root.Insert(v)
	}

	if min := root.FindMin(); min != 3 {
		t.Errorf("Expected min = 3, got %v", min)
	}
	if max := root.FindMax(); max != 20 {
		t.Errorf("Expected max = 20, got %v", max)
	}
}

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

func BenchmarkInsertFromCSV(b *testing.B) {
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
		b.Run(fmt.Sprintf("Insert-%d", size), func(b *testing.B) {
			durations := make([]time.Duration, 0, size)
			root := &TreeNode[string]{Val: values[0]}

			startAll := time.Now()
			for _, val := range values[1:size] {
				startOp := time.Now()
				_ = root.Insert(val)
				durations = append(durations, time.Since(startOp))
			}
			total := time.Since(startAll)
			mean, q1, median, q3 := computeStats(durations)
			b.Logf("Insert %d items: total=%v mean=%v q1=%v median=%v q3=%v", size, total, mean, q1, median, q3)
		})
	}
}

func BenchmarkSearchFromCSV(b *testing.B) {
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
			root := &TreeNode[string]{Val: values[0]}
			for _, val := range values[1:size] {
				_ = root.Insert(val)
			}

			durations := make([]time.Duration, 0, 1000)
			startAll := time.Now()
			for i := 0; i < 1000; i++ {
				v := values[i%size]
				startOp := time.Now()
				_, _ = root.Search(v)
				durations = append(durations, time.Since(startOp))
			}
			total := time.Since(startAll)
			mean, q1, median, q3 := computeStats(durations)
			b.Logf("Search %d items: total=%v mean=%v q1=%v median=%v q3=%v", size, total, mean, q1, median, q3)
		})
	}
}

func BenchmarkDeleteFromCSV(b *testing.B) {
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
			root := &TreeNode[string]{Val: values[0]}
			for _, val := range values[1:size] {
				_ = root.Insert(val)
			}

			durations := make([]time.Duration, 0, size)
			startAll := time.Now()
			for _, val := range values[:size] {
				startOp := time.Now()
				root.Delete(val)
				durations = append(durations, time.Since(startOp))
			}
			total := time.Since(startAll)
			mean, q1, median, q3 := computeStats(durations)
			b.Logf("Delete %d items: total=%v mean=%v q1=%v median=%v q3=%v", size, total, mean, q1, median, q3)
		})
	}
}
