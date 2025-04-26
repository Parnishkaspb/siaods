package index_test

import (
	"lab3/pkg/index"
	"lab3/pkg/util"
	"os"
	"testing"
	"time"

	"lab3/pkg/file"
)

func TestRead(t *testing.T) {
	tests := []struct {
		name        string
		csvContent  string
		filter      file.Data
		startYear   int
		endYear     int
		expectCount int
	}{
		{
			name: "найдён 1 по description и country",
			csvContent: `ID;Country;Description;Designation;Price;Province;Variety;Winery;Year
1;US;Earthy berry flavor;Reserve;45.0;California;Pinot Noir;Castle;2018`,
			filter:      file.Data{Description: "berry", Country: "US"},
			startYear:   2015,
			endYear:     2020,
			expectCount: 1,
		},
		{
			name: "поиск по variety и году (точное совпадение)",
			csvContent: `ID;Country;Description;Designation;Price;Province;Variety;Winery;Year
2;France;Deep aroma;Classic;55.0;Bordeaux;Merlot;Chateau;2014`,
			filter:      file.Data{Variety: "Merlot"},
			startYear:   0,
			endYear:     0,
			expectCount: 1,
		},
		{
			name: "ничего не найдено (год вне диапазона)",
			csvContent: `ID;Country;Description;Designation;Price;Province;Variety;Winery;Year
3;Italy;Smooth finish;Bianco;35.0;Sicily;White Blend;Nicosia;2010`,
			filter:      file.Data{Country: "Italy"},
			startYear:   2020,
			endYear:     2022,
			expectCount: 0,
		},
		{
			name: "два результата попадают в диапазон годов",
			csvContent: `ID;Country;Description;Designation;Price;Province;Variety;Winery;Year
4;US;Red and bold;Reserve;50.0;Oregon;Cabernet;Mount;2017
5;US;Fruity and spicy;Estate;42.0;California;Zinfandel;Valley;2019`,
			filter:      file.Data{Country: "US"},
			startYear:   2016,
			endYear:     2020,
			expectCount: 2,
		},
		{
			name: "один найден с unknown в year",
			csvContent: `ID;Country;Description;Designation;Price;Province;Variety;Winery;Year
6;Germany;Floral and sweet;Kabinett;25.0;Mosel;Riesling;Dr. Loosen;unknown`,
			filter:      file.Data{Country: "Germany"},
			startYear:   0,
			endYear:     0,
			expectCount: 1,
		},
		{
			name: "точный год совпадает с filter.Year",
			csvContent: `ID;Country;Description;Designation;Price;Province;Variety;Winery;Year
7;Spain;Dry and crisp;Gran Reserva;40.0;Rioja;Tempranillo;Torres;2016`,
			filter:      file.Data{Country: "Spain", Year: 2016},
			startYear:   0,
			endYear:     0,
			expectCount: 1,
		},
		{
			name: "поиск по winery и description",
			csvContent: `ID;Country;Description;Designation;Price;Province;Variety;Winery;Year
8;Argentina;Rich with oak;Signature;33.0;Mendoza;Malbec;Catena;2018`,
			filter:      file.Data{Description: "oak", Winery: "Catena"},
			startYear:   2015,
			endYear:     2020,
			expectCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "index_test_*.csv")
			if err != nil {
				t.Fatalf("ошибка создания файла: %v", err)
			}
			defer os.Remove(tmpFile.Name())

			if _, err := tmpFile.WriteString(tt.csvContent); err != nil {
				t.Fatalf("ошибка записи: %v", err)
			}
			tmpFile.Close()

			reader := file.NewDataImpl()
			err, filecsv := reader.Read(tmpFile.Name())
			if err != nil {
				t.Fatalf("ошибка чтения CSV: %v", err)
			}

			_ = os.Remove("file.bleve")

			if err := index.Create(filecsv); err != nil {
				t.Fatalf("ошибка создания индекса: %v", err)
			}

			searchResult, err := index.Search(tt.filter, tt.startYear, tt.endYear)
			if err != nil {
				t.Fatalf("ошибка поиска: %v", err)
			}

			if int(searchResult.Total) != tt.expectCount {
				t.Errorf("ожидалось %d результатов, получено %d", tt.expectCount, searchResult.Total)
			}

			//var sb strings.Builder
			//old := os.Stdout
			//r, w, _ := os.Pipe()
			//os.Stdout = w
			//
			//Print(searchResult)
			//
			//w.Close()
			//os.Stdout = old
			//output := sb.String()
			//
			//if tt.expectCount > 0 && output == "" {
			//	t.Errorf("ожидался вывод результатов, но он пустой")
			//}
		})
	}
}

// Тестовые данные для бенчмарков
var benchmarkCases = []struct {
	name        string
	csvContent  string
	filter      file.Data
	startYear   int
	endYear     int
	expectCount int
	numRuns     int
}{
	{
		name: "SmallDataset_SingleMatch",
		csvContent: `ID;Country;Description;Designation;Price;Province;Variety;Winery;Year
1;US;Earthy berry flavor;Reserve;45.0;California;Pinot Noir;Castle;2018`,
		filter:      file.Data{Description: "berry", Country: "US"},
		startYear:   2015,
		endYear:     2020,
		expectCount: 1,
		numRuns:     1000,
	},
	{
		name: "MediumDataset_MultipleMatches",
		csvContent: `ID;Country;Description;Designation;Price;Province;Variety;Winery;Year` + "\n" +
			generateCSVLines(1000, "US", "California", "Cabernet", "WineryA", 2015, 2020),
		filter:      file.Data{Country: "US"},
		startYear:   2016,
		endYear:     2019,
		expectCount: 400,
		numRuns:     100,
	},
	{
		name: "LargeDataset_ComplexFilter",
		csvContent: `ID;Country;Description;Designation;Price;Province;Variety;Winery;Year` + "\n" +
			generateCSVLines(10000, "France", "Bordeaux", "Merlot", "ChateauX", 2000, 2020),
		filter:      file.Data{Variety: "Merlot", Winery: "ChateauX"},
		startYear:   2010,
		endYear:     2020,
		expectCount: 5000,
		numRuns:     10,
	},
}

func generateCSVLines(n int, country, province, variety, winery string, minYear, maxYear int) string {
	var lines string
	for i := 0; i < n; i++ {
		year := minYear + (i % (maxYear - minYear + 1))
		lines += "1;" + country + ";Description;Designation;25.0;" + province + ";" + variety + ";" + winery + ";" + string(rune(year)) + "\n"
	}
	return lines
}

func BenchmarkSearch(b *testing.B) {
	for _, bc := range benchmarkCases {
		b.Run(bc.name, func(b *testing.B) {
			// Создаем временный файл для теста
			tmpFile, err := os.CreateTemp("", "bench_index_*.csv")
			if err != nil {
				b.Fatalf("не удалось создать временный файл: %v", err)
			}
			defer os.Remove(tmpFile.Name())

			if _, err := tmpFile.WriteString(bc.csvContent); err != nil {
				b.Fatalf("ошибка записи в файл: %v", err)
			}
			tmpFile.Close()

			// Читаем данные и создаем индекс
			reader := file.NewDataImpl()
			err, filecsv := reader.Read(tmpFile.Name())
			if err != nil {
				b.Fatalf("ошибка чтения CSV: %v", err)
			}

			// Удаляем старый индекс, если есть
			_ = os.Remove("file.bleve")

			if err := index.Create(filecsv); err != nil {
				b.Fatalf("ошибка создания индекса: %v", err)
			}

			// Запускаем бенчмарк
			durations := make([]time.Duration, 0, bc.numRuns)
			for i := 0; i < bc.numRuns; i++ {
				start := time.Now()
				searchResult, err := index.Search(bc.filter, bc.startYear, bc.endYear)
				elapsed := time.Since(start)

				if err != nil {
					b.Fatalf("ошибка поиска: %v", err)
				}

				if int(searchResult.Total) != bc.expectCount {
					b.Errorf("ожидалось %d результатов, получено %d", bc.expectCount, searchResult.Total)
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
