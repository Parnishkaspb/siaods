package main

import (
	"fmt"
	"lab3/pkg/file"
)

func main() {
	dataImpl := file.NewDataImpl()

	filename := "EDAResult.csv"

	err, _ := dataImpl.Read(filename)
	if err != nil {
		panic(err)
	}

	//for id, d := range dataImpl.Data {
	//	fmt.Printf("ID: %d, Страна: %s, Сорт: %s\n", id, d.Country, d.Variety)
	//}

	fmt.Printf("Всего строк в DataSet", len(dataImpl.Data))

}
