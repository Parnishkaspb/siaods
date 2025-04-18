package file

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strconv"
)

type DataImpl struct {
	Data map[int]Data
}

func NewDataImpl() *DataImpl {
	return &DataImpl{
		Data: make(map[int]Data),
	}
}

func (d *DataImpl) Read(path string) (error, DataImpl) {
	file, err := os.Open(path)
	if err != nil {
		return errors.New("ошибка" + err.Error()), DataImpl{}
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ';'
	reader.FieldsPerRecord = -1

	records, err := reader.ReadAll()
	if err != nil {
		return errors.New("ошибка" + err.Error()), DataImpl{}
	}

	for i, record := range records[1:] {

		if len(record) < 9 {
			return fmt.Errorf("ошибка: недостаточно данных в строке %d", i+2), DataImpl{}

		}

		id, err := strconv.Atoi(record[0])
		if err != nil {
			return fmt.Errorf("ошибка преобразования ID в строке %d: %w", i+2, err), DataImpl{}
		}

		price64, err := strconv.ParseFloat(record[4], 32)
		if err != nil {
			return fmt.Errorf("ошибка преобразования Price в строке %d: %w", i+2, err), DataImpl{}
		}

		var year int
		if record[8] == "unknown" {
			year = 0
		} else {
			parsedYear, err := strconv.Atoi(record[8])
			if err != nil {
				return fmt.Errorf("ошибка преобразования Year в строке %d: %w", i+2, err), DataImpl{}
			}
			year = parsedYear
		}

		arrayData := Data{
			ID:          id,
			Country:     record[1],
			Description: record[2],
			Designation: record[3],
			Price:       float32(price64),
			Province:    record[5],
			Variety:     record[6],
			Winery:      record[7],
			Year:        year,
		}

		d.Data[arrayData.ID] = arrayData
	}

	return nil, *d
}
