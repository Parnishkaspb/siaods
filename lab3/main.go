// package main
//
// import (
//
//	"fmt"
//	"lab3/pkg/file"
//	"lab3/pkg/index"
//
// )
//
//	func main() {
//		dataImpl := file.NewDataImpl()
//
//		filename := "EDAResult.csv"
//
//		err, filecsv := dataImpl.Read(filename)
//		if err != nil {
//			panic(err)
//		}
//
//		err = index.Create(filecsv)
//		if err != nil {
//			fmt.Println(err)
//		}
//
//		filter := file.Data{
//			Description: "earthy berry",
//			Country:     "US",
//			Variety:     "Pinot Noir",
//		}
//
//		result, err := index.Search(filter, 2015, 2020)
//		if err != nil {
//			fmt.Println(err)
//		}
//		index.Print(result)
//	}
package main

import (
	http2 "lab3/pkg/http"
	"log"
	"net/http"
)

func main() {
	http2.StartServer()
	log.Println("Сервер запущен на http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Ошибка запуска сервера:", err)
	}
}
