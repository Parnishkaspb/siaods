package http

type Index interface {
	Create(title any, text string)
	Delete(id int)
}
