package http

import (
	"encoding/json"
	"github.com/blevesearch/bleve/v2"
	"lab3/pkg/file"
	"lab3/pkg/index"
	"net/http"
)

type SearchRequest struct {
	Country     string  `json:"country"`
	Description string  `json:"description"`
	Designation string  `json:"designation"`
	Price       float32 `json:"price"`
	Province    string  `json:"province"`
	Variety     string  `json:"variety"`
	Winery      string  `json:"winery"`
	Years       []int   `json:"years"`
}

type ApiResponse struct {
	Message string              `json:"message"`
	Data    *bleve.SearchResult `json:"data,omitempty"`
}

func startHandler() {
	http.HandleFunc(
		"/start",
		func(w http.ResponseWriter, r *http.Request) {
			dataImpl := file.NewDataImpl()
			filename := "EDAResult.csv"

			err, filecsv := dataImpl.Read(filename)
			if err != nil {
				http.Error(w, "Ошибка чтения CSV: "+err.Error(), http.StatusInternalServerError)
				return
			}

			err = index.Create(filecsv)
			if err != nil {
				http.Error(w, "Ошибка индексации: "+err.Error(), http.StatusInternalServerError)
				return
			}

			WriteApiResponse(w, nil, "Индекс успешно создан", http.StatusOK)
		},
	)
}

func searchHandler() {
	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Метод не разрешён", http.StatusMethodNotAllowed)
			return
		}

		var req SearchRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "Ошибка парсинга JSON: "+err.Error(), http.StatusBadRequest)
			return
		}

		if len(req.Years) < 2 {
			http.Error(w, "Поле 'years' должно содержать минимум 2 значения (start и end)", http.StatusBadRequest)
			return
		}

		filter := file.Data{
			ID:          0,
			Country:     req.Country,
			Description: req.Description,
			Designation: req.Designation,
			Price:       req.Price,
			Province:    req.Province,
			Variety:     req.Variety,
			Winery:      req.Winery,
		}

		result, err := index.Search(filter, req.Years[0], req.Years[1])
		if err != nil {
			http.Error(w, "Ошибка поиска: "+err.Error(), http.StatusInternalServerError)
			return
		}

		WriteApiResponse(w, result, "Поиск завершён", http.StatusOK)
	})
}

func WriteApiResponse(w http.ResponseWriter, result *bleve.SearchResult, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	response := ApiResponse{
		Message: message,
		Data:    result,
	}
	json.NewEncoder(w).Encode(response)
}

func StartServer() {
	startHandler()
	searchHandler()
}
