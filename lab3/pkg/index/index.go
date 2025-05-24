package index

import (
	"fmt"
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search/query"
	"lab3/pkg/file"
	"os"
)

var index bleve.Index

var indexPath = "test.bleve"

//var indexPath = "file.bleve"

func Create(filecsv file.DataImpl) error {
	fmt.Println("начало создание индекса")

	_ = os.RemoveAll(indexPath) // Полное удаление старого индекса

	index, err := bleve.New(indexPath, bleve.NewIndexMapping())
	if err != nil {
		return fmt.Errorf("ошибка при создании индекса: %w", err)
	}
	defer index.Close()

	for _, doc := range filecsv.Data {
		fmt.Println(doc.ID)
		err := index.Index(fmt.Sprintf("%d", doc.ID), doc)
		if err != nil {
			return fmt.Errorf("ошибка при индексации документа ID=%d: %w", doc.ID, err)
		}
	}

	fmt.Println("окончание создание индекса")
	return nil
}

func CreateInMemory(filecsv file.DataImpl) error {
	//fmt.Println("начало создание индекса в памяти")

	index, err := bleve.NewMemOnly(bleve.NewIndexMapping())
	if err != nil {
		return fmt.Errorf("ошибка при создании in-memory индекса: %w", err)
	}

	for _, doc := range filecsv.Data {
		err := index.Index(fmt.Sprintf("%d", doc.ID), doc)
		if err != nil {
			return fmt.Errorf("ошибка при индексации документа ID=%d: %w", doc.ID, err)
		}
	}

	//fmt.Println("окончание создание индекса в памяти")
	return nil
}

func openIndex() error {
	if index != nil {
		return nil
	}

	var err error
	index, err = bleve.Open(indexPath)
	return err
}

func Search(filter file.Data, startYear, endYear int) (*bleve.SearchResult, error) {
	//fmt.Println("начало поиска индекса")
	err := openIndex()
	if err != nil {
		return nil, fmt.Errorf("ошибка при открытии индекса: %w", err)
	}

	query := buildSearchQuery(filter, startYear, endYear)

	searchRequest := bleve.NewSearchRequest(query)
	searchRequest.Fields = []string{"ID", "Title", "Year", "Description", "Country", "Variety", "Winery"}

	searchResult, err := index.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("ошибка при выполнении поиска: %w", err)
	}
	//fmt.Println("конец создание индекса")
	return searchResult, nil
}

func Print(result *bleve.SearchResult) {
	fmt.Println("Результаты поиска:")
	for _, hit := range result.Hits {
		fmt.Printf("ID: %v\nYear: %v\nDescription: %v\n\n",
			hit.ID, hit.Fields["Year"], hit.Fields["Description"])
	}
}

func buildSearchQuery(filter file.Data, startYear, endYear int) query.Query {
	boolQuery := bleve.NewBooleanQuery()

	// По описанию
	if filter.Description != "" {
		q := bleve.NewMatchQuery(filter.Description)
		q.SetField("Description")
		boolQuery.AddMust(q)
	}

	// Строковые поля
	addFieldQuery := func(fieldName, value string) {
		if value != "" {
			q := bleve.NewMatchQuery(value)
			q.SetField(fieldName)
			boolQuery.AddMust(q)
		}
	}

	addFieldQuery("Country", filter.Country)
	addFieldQuery("Designation", filter.Designation)
	addFieldQuery("Province", filter.Province)
	addFieldQuery("Variety", filter.Variety)
	addFieldQuery("Winery", filter.Winery)

	// По году
	if startYear > 0 || endYear > 0 {
		var startPtr, endPtr *float64
		if startYear > 0 {
			start := float64(startYear)
			startPtr = &start
		}
		if endYear > 0 {
			end := float64(endYear)
			endPtr = &end
		}
		q := bleve.NewNumericRangeQuery(startPtr, endPtr)
		q.SetField("Year")
		boolQuery.AddMust(q)
	} else if filter.Year > 0 {
		year := float64(filter.Year)
		q := bleve.NewNumericRangeQuery(&year, &year)
		q.SetField("Year")
		boolQuery.AddMust(q)
	}

	return boolQuery
}
