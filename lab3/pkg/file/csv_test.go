package file_test

import (
	"lab3/pkg/file"
	"os"
	"testing"
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
