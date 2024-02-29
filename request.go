package sense

import (
	"net/http"

	"github.com/creamsensation/sense/internal/constant/header"
)

type RequestContext interface {
	ContentType() string
	Header() http.Header
	Host() string
	Ip() string
	Is() RequestIsContext
	Method() string
	Path() string
	Protocol() string
	Raw() *http.Request
	UserAgent() string
}

type RequestIsContext interface {
	Get() bool
	Post() bool
	Put() bool
	Patch() bool
	Delete() bool
}

type request struct {
	req *http.Request
}

func (r *request) ContentType() string {
	return r.req.Header.Get(header.ContentType)
}

func (r *request) Header() http.Header {
	return r.req.Header
}

func (r *request) Host() string {
	return r.Protocol() + "://" + r.req.Host
}

func (r *request) Ip() string {
	return r.req.Header.Get("X-Forwarded-For")
}

func (r *request) Is() RequestIsContext {
	return r
}

func (r *request) Method() string {
	return r.req.Method
}

func (r *request) Path() string {
	return r.req.URL.Path
}

func (r *request) Protocol() string {
	if r.req.TLS == nil {
		return "http"
	}
	return "https"
}

func (r *request) Raw() *http.Request {
	return r.req
}

func (r *request) UserAgent() string {
	return r.req.Header.Get(header.UserAgent)
}

func (r *request) Get() bool {
	return r.req.Method == http.MethodGet
}

func (r *request) Post() bool {
	return r.req.Method == http.MethodPost
}

func (r *request) Put() bool {
	return r.req.Method == http.MethodPut
}

func (r *request) Patch() bool {
	return r.req.Method == http.MethodPatch
}

func (r *request) Delete() bool {
	return r.req.Method == http.MethodDelete
}

func PathValue[T Assert](c RequestContext, key string, defaultValue ...T) T {
	value := c.Raw().PathValue(key)
	if len(value) == 0 && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return assertStringToType[T](value)
}

func Query[T Assert](c RequestContext, key string, defaultValue ...T) T {
	value := c.Raw().URL.Query().Get(key)
	if len(value) == 0 && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return assertStringToType[T](value)
}
