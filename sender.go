package sense

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/creamsensation/socketer"

	"github.com/creamsensation/auth"
	"github.com/creamsensation/sense/internal/constant/contentType"
	"github.com/creamsensation/sense/internal/constant/dataType"
)

type SendContext interface {
	Header() http.Header
	Status(statusCode int) SendContext
	Error(err any) error
	Text(value string) error
	Bool(value bool) error
	Json(value any) error
	Xml(value any) error
	Redirect(url string) error
	File(name string, bytes []byte) error
	Ws(name string) WsWriter
}

type sender struct {
	auth        auth.Manager
	interceptor *interceptor
	request     *request
	ws          map[string]socketer.Ws
	res         http.ResponseWriter
	bytes       []byte
	dataType    string
	contentType string
	value       string
	statusCode  int
}

func (s *sender) Header() http.Header {
	return s.res.Header()
}

func (s *sender) Status(statusCode int) SendContext {
	s.statusCode = statusCode
	return s
}

func (s *sender) Error(e any) error {
	var err error
	switch v := e.(type) {
	case nil:
		return s.Bool(true)
	case string:
		err = errors.New(v)
	case error:
		err = v
	default:
		err = errors.New(fmt.Sprintf("%v", e))
	}
	var bytes []byte
	if s.interceptor == nil || (s.interceptor != nil && s.interceptor.onError == nil) {
		bytes, err = wrapError(err)
	}
	if s.interceptor != nil && s.interceptor.onError != nil {
		bytes, err = wrapError(s.interceptor.onError(s.request, err))
	}
	s.bytes = bytes
	s.dataType = dataType.Error
	s.contentType = contentType.Json
	if s.statusCode == http.StatusOK {
		s.statusCode = http.StatusBadRequest
	}
	return err
}

func (s *sender) Json(value any) error {
	var bytes []byte
	var err error
	if s.interceptor == nil || (s.interceptor != nil && s.interceptor.onJson == nil) {
		bytes, err = wrapResult(value)
	}
	if s.interceptor != nil && s.interceptor.onJson != nil {
		bytes, err = wrapResult(s.interceptor.onJson(s.request, value))
	}
	s.bytes = bytes
	s.dataType = dataType.Json
	s.contentType = contentType.Json
	return err
}

func (s *sender) Xml(value any) error {
	var bytes []byte
	var err error
	if s.interceptor == nil || (s.interceptor != nil && s.interceptor.onXml == nil) {
		bytes, err = wrapResult(value)
	}
	if s.interceptor != nil && s.interceptor.onJson != nil {
		bytes, err = wrapResult(s.interceptor.onXml(s.request, value))
	}
	s.bytes = bytes
	s.dataType = dataType.Xml
	s.contentType = contentType.Xml
	return err
}

func (s *sender) Text(value string) error {
	var bytes []byte
	var err error
	if s.interceptor == nil || (s.interceptor != nil && s.interceptor.onText == nil) {
		bytes, err = wrapResult(value)
	}
	if s.interceptor != nil && s.interceptor.onText != nil {
		bytes, err = wrapResult(s.interceptor.onText(s.request, value))
	}
	s.bytes = bytes
	s.dataType = dataType.Text
	s.contentType = contentType.Json
	return err
}

func (s *sender) Bool(value bool) error {
	var bytes []byte
	var err error
	if s.interceptor == nil || (s.interceptor != nil && s.interceptor.onBool == nil) {
		bytes, err = wrapResult(value)
	}
	if s.interceptor != nil && s.interceptor.onBool != nil {
		bytes, err = wrapResult(s.interceptor.onBool(s.request, value))
	}
	s.bytes = bytes
	s.dataType = dataType.Bool
	s.contentType = contentType.Json
	return err
}

func (s *sender) Redirect(url string) error {
	s.value = url
	s.dataType = dataType.Redirect
	return nil
}

func (s *sender) File(name string, bytes []byte) error {
	s.value = name
	s.bytes = bytes
	s.dataType = dataType.Stream
	s.contentType = contentType.OctetStream
	return nil
}

func (s *sender) Ws(name string) WsWriter {
	if _, ok := s.ws[name]; !ok {
		panic(ErrorInvalidWebsocket)
	}
	return createWsWriter(s.ws, name, s.auth)
}
