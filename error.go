package sense

import "errors"

var (
	ErrorInvalidDatabase  = errors.New("invalid database")
	ErrorInvalidWebsocket = errors.New("invalid websocket")
	ErrorInvalidLang      = errors.New("invalid lang")
	ErrorInvalidMultipart = errors.New("request has not multipart content type")
	ErrorOpenFile         = errors.New("file cannot be opened")
	ErrorReadData         = errors.New("cannot read data")
)

type ErrorsWrapper[T any] struct {
	Errors T `json:"errors"`
}
