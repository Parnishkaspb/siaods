package R_Tree

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"
	"time"
)

func rectsIntersect(a, b Rect) bool {
	return a.MinX < b.MaxX && a.MaxX > b.MinX &&
		a.MinY < b.MaxY && a.MaxY > b.MinY
}

func TestInsertSearchDelete(t *testing.T) {
	tree := New()

	items := []Item{
		{Rect: Rect{0, 0, 1, 1}, Data: "A"},
		{Rect: Rect{2, 2, 3, 3}, Data: "B"},
		{Rect: Rect{1, 1, 2, 2}, Data: "C"},
		{Rect: Rect{3, 3, 4, 4}, Data: "D"},
		{Rect: Rect{4, 4, 5, 5}, Data: "E"},
	}

	for _, item := range items {
		tree.Insert(item)
	}

	results := tree.Search(Rect{1.5, 1.5, 4.5, 4.5})
	if len(results) != 4 {
		t.Errorf("expected 4 results, got %d", len(results))
	}

	ok := tree.Delete("C")
	if !ok {
		t.Error("failed to delete item C")
	}

	// узкая область, чтобы остался только C
	results = tree.Search(Rect{1.01, 1.01, 1.99, 1.99})
	if len(results) != 0 {
		t.Errorf("expected 0 results after deletion, got %d", len(results))
	}
}

func TestKnn(t *testing.T) {
	tree := New()
	items := []Item{
		{Rect: Rect{0, 0, 1, 1}, Data: "A"},
		{Rect: Rect{5, 5, 6, 6}, Data: "B"},
		{Rect: Rect{2, 2, 3, 3}, Data: "C"},
		{Rect: Rect{7, 7, 8, 8}, Data: "D"},
		{Rect: Rect{4, 4, 5, 5}, Data: "E"},
	}
	for _, item := range items {
		tree.Insert(item)
	}
	knn := tree.Knn(3, 3, 3)
	if len(knn) != 3 {
		t.Errorf("expected 3 knn results, got %d", len(knn))
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

func BenchmarkInsertStats(b *testing.B) {
	sizes := []int{1000, 10000, 100000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Insert-%d", size), func(b *testing.B) {
			tree := New()
			durations := make([]time.Duration, 0, size)
			data := make([]Item, size)

			for i := 0; i < size; i++ {
				x := rand.Float64() * 10000
				y := rand.Float64() * 10000
				data[i] = Item{Rect: Rect{x, y, x + 1, y + 1}, Data: i}
			}

			startAll := time.Now()
			for _, item := range data {
				startOp := time.Now()
				tree.Insert(item)
				durations = append(durations, time.Since(startOp))
			}
			total := time.Since(startAll)

			mean, q1, median, q3 := computeStats(durations)
			b.Logf("Insert %d items: total=%v mean=%v q1=%v median=%v q3=%v", size, total, mean, q1, median, q3)
		})
	}
}

func BenchmarkSearchStats(b *testing.B) {
	sizes := []int{1000, 10000, 100000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Search-%d", size), func(b *testing.B) {
			tree := New()
			for i := 0; i < size; i++ {
				x := rand.Float64() * 10000
				y := rand.Float64() * 10000
				tree.Insert(Item{Rect: Rect{x, y, x + 1, y + 1}, Data: i})
			}

			durations := make([]time.Duration, 0, 1000)
			startAll := time.Now()
			for i := 0; i < 1000; i++ {
				startOp := time.Now()
				tree.Search(Rect{5000, 5000, 5010, 5010})
				durations = append(durations, time.Since(startOp))
			}
			total := time.Since(startAll)

			mean, q1, median, q3 := computeStats(durations)
			b.Logf("Search %d items: total=%v mean=%v q1=%v median=%v q3=%v", size, total, mean, q1, median, q3)
		})
	}
}

func BenchmarkLinearSearchStats(b *testing.B) {
	sizes := []int{1000, 10000, 100000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("LinearSearch-%d", size), func(b *testing.B) {
			items := make([]Item, 0, size)
			for i := 0; i < size; i++ {
				x := rand.Float64() * 10000
				y := rand.Float64() * 10000
				items = append(items, Item{Rect: Rect{x, y, x + 1, y + 1}, Data: i})
			}

			durations := make([]time.Duration, 0, 1000)
			query := Rect{5000, 5000, 5010, 5010}
			startAll := time.Now()
			for i := 0; i < 1000; i++ {
				startOp := time.Now()
				var _results []Item
				for _, item := range items {
					if rectsIntersect(item.Rect, query) {
						_results = append(_results, item)
					}
				}
				durations = append(durations, time.Since(startOp))
			}
			total := time.Since(startAll)

			mean, q1, median, q3 := computeStats(durations)
			b.Logf("Linear Search %d items: total=%v mean=%v q1=%v median=%v q3=%v", size, total, mean, q1, median, q3)
		})
	}
}

func BenchmarkDeleteStats(b *testing.B) {
	sizes := []int{1000, 10000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Delete-%d", size), func(b *testing.B) {
			tree := New()
			data := make([]Item, size)

			for i := 0; i < size; i++ {
				x := rand.Float64() * 10000
				y := rand.Float64() * 10000
				data[i] = Item{Rect: Rect{x, y, x + 1, y + 1}, Data: i}
				tree.Insert(data[i])
			}

			durations := make([]time.Duration, 0, size)
			startAll := time.Now()
			for _, item := range data {
				startOp := time.Now()
				tree.Delete(item)
				durations = append(durations, time.Since(startOp))
			}
			total := time.Since(startAll)

			mean, q1, median, q3 := computeStats(durations)
			b.Logf("Delete %d items: total=%v mean=%v q1=%v median=%v q3=%v", size, total, mean, q1, median, q3)
		})
	}
}

func BenchmarkKnnStats(b *testing.B) {
	sizes := []int{1000, 10000, 100000}
	for _, size := range sizes {
		b.Run(fmt.Sprintf("Knn-%d", size), func(b *testing.B) {
			tree := New()
			for i := 0; i < size; i++ {
				x := rand.Float64() * 10000
				y := rand.Float64() * 10000
				tree.Insert(Item{Rect: Rect{x, y, x + 1, y + 1}, Data: i})
			}

			durations := make([]time.Duration, 0, 1000)
			startAll := time.Now()
			for i := 0; i < 1000; i++ {
				startOp := time.Now()
				tree.Knn(5000, 5000, 10)
				durations = append(durations, time.Since(startOp))
			}
			total := time.Since(startAll)

			mean, q1, median, q3 := computeStats(durations)
			b.Logf("Knn %d items: total=%v mean=%v q1=%v median=%v q3=%v", size, total, mean, q1, median, q3)
		})
	}
}
