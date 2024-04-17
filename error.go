package sense

import "errors"

var (
	ErrorInvalidDatabase   = errors.New("invalid database")
	ErrorInvalidWebsocket  = errors.New("invalid websocket")
	ErrorInvalidLang       = errors.New("invalid lang")
	ErrorInvalidMultipart  = errors.New("request has not multipart content type")
	ErrorOpenFile          = errors.New("file cannot be opened")
	ErrorReadData          = errors.New("cannot read data")
	ErrorPointerTarget     = errors.New("target must be a pointer")
	ErrorQueryParamMissing = errors.New("query param is missing")
	ErrorPathValueMissing  = errors.New("path value is missing")
)

type ErrorsWrapper[T any] struct {
	Errors T `json:"errors"`
}
