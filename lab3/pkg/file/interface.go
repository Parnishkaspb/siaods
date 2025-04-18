package file

type Data struct {
	ID          int
	Country     string
	Description string
	Designation string
	Price       float32
	Province    string
	Variety     string
	Winery      string
	Year        int
}

type WorkWithFile interface {
	Read(path string) (error, bool)
}
