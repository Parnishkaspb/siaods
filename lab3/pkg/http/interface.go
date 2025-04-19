package http

type Front interface {
	startHandler()
	searchHandler()
	WriteApiResponse()
}
