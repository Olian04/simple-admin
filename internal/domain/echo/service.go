package echo

import "strings"

type Service struct{}

type Request struct {
	Message string `json:"message"`
}

type Response struct {
	Message string `json:"message"`
}

func NewService() Service {
	return Service{}
}

func (Service) Echo(req Request) Response {
	return Response{Message: strings.TrimSpace(req.Message)}
}
