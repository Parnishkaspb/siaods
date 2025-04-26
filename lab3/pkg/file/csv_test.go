package file_test

import (
	"lab3/pkg/file"
	"lab3/pkg/util"
	"os"
	"testing"
	"time"
)

func TestRead(t *testing.T) {
	tests := []struct {
		scenarioName  string
		csvContent    string
		expectError   bool
		expectedCount int
		expectedYear  int
	}{
		{
			scenarioName: "valid csv with year and unknown",
			csvContent: `ID;Country;Description;Designation;Price;Province;Variety;Winery;Year
1;Italy;Good wine;Bianco;25.0;Sicily;White Blend;Nicosia;2013
2;US;Rich flavor;Reserve;69.0;California;Pinot Noir;Castle;unknown`,
			expectError:   false,
			expectedCount: 2,
			expectedYear:  0,
		},
		{
			scenarioName: "invalid price",
			csvContent: `ID;Country;Description;Designation;Price;Province;Variety;Winery;Year
1;Italy;Bad wine;BrokenPrice;abc;Tuscany;Chianti;Barone;2012`,
			expectError: true,
		},
		{
			scenarioName: "missing fields",
			csvContent: `ID;Country;Description;Designation;Price;Province;Variety;Winery;Year
1;Italy;Not enough columns`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.scenarioName, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "testdata_*.csv")
			if err != nil {
				t.Fatalf("не удалось создать временный файл: %v", err)
			}
			defer os.Remove(tmpFile.Name())

			if _, err := tmpFile.WriteString(tt.csvContent); err != nil {
				t.Fatalf("ошибка записи в файл: %v", err)
			}
			tmpFile.Close()

			reader := file.NewDataImpl()
			err, _ = reader.Read(tmpFile.Name())

			if tt.expectError {
				if err == nil {
					t.Errorf("ожидалась ошибка, но её не было")
				}
				return
			}

			if err != nil {
				t.Errorf("неожиданная ошибка: %v", err)
			}

			if len(reader.Data) != tt.expectedCount {
				t.Errorf("ожидалось %d записей, получено %d", tt.expectedCount, len(reader.Data))
			}

			if tt.expectedYear != 0 {
				if val, ok := reader.Data[2]; ok && val.Year != tt.expectedYear {
					t.Errorf("ожидался год %d, получено %d", tt.expectedYear, val.Year)
				}
			}
		})
	}
}

// Тестовые данные для бенчмарков
var benchmarkCases = []struct {
	name    string
	content string
	numRuns int
}{
	{
		name: "SmallFile",
		content: `ID;Country;Description;Designation;Price;Province;Variety;Winery;Year
1;Italy;Good wine;Bianco;25.0;Sicily;White Blend;Nicosia;2013
2;US;Rich flavor;Reserve;69.0;California;Pinot Noir;Castle;unknown`,
		numRuns: 1000,
	},
	{
		name: "MediumFile",
		content: `ID;Country;Description;Designation;Price;Province;Variety;Winery;Year` + "\n" +
			generateCSVLines(1000),
		numRuns: 100,
	},
	{
		name: "LargeFile",
		content: `ID;Country;Description;Designation;Price;Province;Variety;Winery;Year` + "\n" +
			generateCSVLines(10000),
		numRuns: 10,
	},
}

// generateCSVLines создает n строк CSV данных
func generateCSVLines(n int) string {
	var lines string
	for i := 0; i < n; i++ {
		lines += "1;Italy;Good wine;Bianco;25.0;Sicily;White Blend;Nicosia;2013\n"
	}
	return lines
}

func BenchmarkRead(b *testing.B) {
	for _, bc := range benchmarkCases {
		b.Run(bc.name, func(b *testing.B) {
			// Создаем временный файл для теста
			tmpFile, err := os.CreateTemp("", "benchdata_*.csv")
			if err != nil {
				b.Fatalf("не удалось создать временный файл: %v", err)
			}
			defer os.Remove(tmpFile.Name())

			if _, err := tmpFile.WriteString(bc.content); err != nil {
				b.Fatalf("ошибка записи в файл: %v", err)
			}
			tmpFile.Close()

			// Запускаем бенчмарк
			durations := make([]time.Duration, 0, bc.numRuns)
			for i := 0; i < bc.numRuns; i++ {
				reader := file.NewDataImpl()

				start := time.Now()
				err, _ = reader.Read(tmpFile.Name())
				elapsed := time.Since(start)

				if err != nil {
					b.Fatalf("неожиданная ошибка: %v", err)
				}

				durations = append(durations, elapsed)
			}

			// Вычисляем и выводим статистику
			mean, q1, median, q3 := util.ComputeStats(durations)
			b.ReportMetric(float64(mean), "mean")
			b.ReportMetric(float64(q1), "q1")
			b.ReportMetric(float64(median), "median")
			b.ReportMetric(float64(q3), "q3")
		})
	}
}
