package index_test

import (
	"lab3/pkg/index"
	"os"
	"testing"

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
