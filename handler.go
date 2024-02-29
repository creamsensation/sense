package sense

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/creamsensation/sense/internal/constant/contentType"
	"github.com/creamsensation/sense/internal/constant/dataType"
	"github.com/creamsensation/sense/internal/constant/header"
	"github.com/creamsensation/sense/internal/constant/model"
	"github.com/creamsensation/socketer"
)

type Handler func(c Context) error

func createHandlerFunc(
	config Config, route Route, handler Handler, middlewares []Handler, hook *hook, interceptor *interceptor,
	ws map[string]socketer.Ws,
) func(
	http.ResponseWriter, *http.Request,
) {
	return func(res http.ResponseWriter, req *http.Request) {
		var err error
		c := createHandlerContext(config, res, req, interceptor, ws)
		defer createRecover(c.request, res, interceptor)
		middlewares = applyInternalMiddlewares(route, middlewares)
		for _, middleware := range middlewares {
			c.mu.Lock()
			err = middleware(c)
			if err != nil {
				c.mu.Unlock()
				createHandlerResponse(c, hook, err)
				return
			}
		}
		if len(c.send.dataType) == 0 {
			err = handler(c)
		}
		createHandlerResponse(c, hook, err)
	}
}

func createWsHandlerFunc(
	config Config, route Route, handler Handler, middlewares []Handler, interceptor *interceptor,
	ws map[string]socketer.Ws,
	name string,
) func(
	http.ResponseWriter, *http.Request,
) {
	return func(res http.ResponseWriter, req *http.Request) {
		var err error
		c := createHandlerContext(config, res, req, interceptor, ws)
		defer createRecover(c.request, res, interceptor)
		middlewares = applyInternalMiddlewares(route, middlewares)
		for _, middleware := range middlewares {
			c.mu.Lock()
			err = middleware(c)
			if err != nil {
				c.mu.Unlock()
				return
			}
		}
		var id int
		if err := ws[name].OnRead(
			func(bytes []byte) {
				c.parse.bytes = bytes
				if err := handler(c); err != nil {
					panic(err) // TODO
				}
				c.parse.bytes = nil
			},
		).Serve(req, res, id); err != nil {
			panic(err)
		}
	}
}

func createHandlerResponse(c *handlerContext, hook *hook, err error) {
	if err != nil {
		errorBytes, err := wrapError(err)
		if err != nil {
			http.Error(c.res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			callHookIfExists(
				c, hook, dataType.Error, contentType.Text, []byte(http.StatusText(http.StatusInternalServerError)),
			)
			return
		}
		c.res.Header().Set(header.ContentType, contentType.Json)
		if c.send.statusCode == http.StatusOK {
			c.res.WriteHeader(http.StatusInternalServerError)
		}
		_, err = c.res.Write(errorBytes)
		if err != nil {
			http.Error(c.res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			callHookIfExists(
				c, hook, dataType.Error, contentType.Text, []byte(http.StatusText(http.StatusInternalServerError)),
			)
			return
		}
		return
	}
	if c.send.dataType == dataType.Redirect {
		if c.send.statusCode == http.StatusOK {
			c.send.statusCode = http.StatusFound
		}
		http.Redirect(c.res, c.req, c.send.value, c.send.statusCode)
		return
	}
	if c.send.dataType == dataType.Stream {
		c.res.Header().Set(header.ContentDisposition, fmt.Sprintf("attachment;filename=%s", c.send.value))
		c.res.Header().Set(header.ContentLength, fmt.Sprintf("%d", len(c.send.bytes)))
	}
	c.res.Header().Set(header.ContentType, c.send.contentType)
	c.res.WriteHeader(c.send.statusCode)
	if _, err = c.res.Write(c.send.bytes); err != nil {
		http.Error(c.res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		callHookIfExists(c, hook, dataType.Error, contentType.Text, []byte(http.StatusText(http.StatusInternalServerError)))
		return
	}
	callHookIfExists(c, hook, c.send.dataType, c.send.contentType, c.send.bytes)
}

func callHookIfExists(c *handlerContext, hook *hook, dt, ct string, bytes []byte) {
	switch dt {
	case dataType.Error:
		if hook.onError != nil {
			switch ct {
			case contentType.Json:
				var e model.Error
				_ = json.Unmarshal(bytes, &e)
				hook.onError(c.Request(), errors.New(e.Error))
			case contentType.Text:
				hook.onError(c.Request(), errors.New(string(bytes)))
			}
		}
	}
}

func createRecover(request *request, res http.ResponseWriter, interceptor *interceptor) {
	if e := recover(); e != nil {
		var bytes []byte
		err := errors.New(fmt.Sprintf("%v", e))
		if interceptor == nil || (interceptor != nil && interceptor.onError == nil) {
			bytes, err = wrapError(err)
		}
		if interceptor != nil && interceptor.onError != nil {
			bytes, err = wrapError(interceptor.onError(request, err))
		}
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		res.Header().Set(header.ContentType, contentType.Json)
		res.WriteHeader(http.StatusInternalServerError)
		if _, err = res.Write(bytes); err != nil {
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		return
	}
}

func applyInternalMiddlewares(route Route, middlewares []Handler) []Handler {
	if len(route.Firewalls) > 0 {
		middlewares = append(middlewares, authMiddleware(route.Firewalls))
	}
	return middlewares
}
