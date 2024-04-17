package sense

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	
	"github.com/creamsensation/sense/internal/constant/contentType"
	"github.com/creamsensation/sense/internal/constant/dataType"
	"github.com/creamsensation/sense/internal/constant/header"
)

type Handler func(c Context) error

func createHandlerFunc(args handlerFuncArgs) func(
	http.ResponseWriter, *http.Request,
) {
	return func(res http.ResponseWriter, req *http.Request) {
		var err error
		c := createHandlerContext(
			handlerContextArgs{
				config: args.config,
				req:    req,
				res:    res,
			},
		)
		if args.config.Router.Recover {
			defer createRecover(res)
		}
		args.middlewares = applyInternalMiddlewares(args.route, args.middlewares)
		for _, middleware := range args.middlewares {
			c.mu.Lock()
			err = middleware(c)
			if err != nil {
				c.mu.Unlock()
				createHandlerResponse(c, err)
				return
			}
		}
		if len(c.send.dataType) == 0 {
			err = args.handler(c)
		}
		createHandlerResponse(c, err)
	}
}

func createWsHandlerFunc(args handlerFuncArgs) func(
	http.ResponseWriter, *http.Request,
) {
	return func(res http.ResponseWriter, req *http.Request) {
		var err error
		c := createHandlerContext(
			handlerContextArgs{
				config: args.config,
				req:    req,
				res:    res,
				ws:     args.ws,
			},
		)
		if args.config.Router.Recover {
			defer createRecover(res)
		}
		args.middlewares = applyInternalMiddlewares(args.route, args.middlewares)
		for _, middleware := range args.middlewares {
			c.mu.Lock()
			err = middleware(c)
			if err != nil {
				c.mu.Unlock()
				return
			}
		}
		var id int
		if err := args.ws[args.name].OnRead(
			func(bytes []byte) {
				c.parse.bytes = bytes
				if err := args.handler(c); err != nil {
					panic(err)
				}
				c.parse.bytes = nil
			},
		).Serve(req, res, id); err != nil {
			panic(err)
		}
	}
}

func createEmptyHandlerResponse() func(
	http.ResponseWriter, *http.Request,
) {
	return func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
	}
}

func createHandlerCanonicalRedirect() func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		uri := req.Header.Get("X-Forwarded-Uri")
		if len(uri) == 0 {
			uri = req.RequestURI
		}
		http.Redirect(res, req, strings.TrimSuffix(uri, "/"), http.StatusMovedPermanently)
	}
}

func createHandlerResponse(c *handlerContext, err error) {
	if err != nil {
		errorBytes, err := wrapError(err)
		if err != nil {
			http.Error(c.res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		c.res.Header().Set(header.ContentType, contentType.Json)
		if c.send.statusCode == http.StatusOK {
			c.res.WriteHeader(http.StatusInternalServerError)
		}
		_, err = c.res.Write(errorBytes)
		if err != nil {
			http.Error(c.res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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
		return
	}
}

func createRecover(res http.ResponseWriter) {
	if e := recover(); e != nil {
		var bytes []byte
		err := errors.New(fmt.Sprintf("%v", e))
		bytes, err = wrapError(err)
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
		middlewares = append(middlewares, trailingSlashMiddleware())
		middlewares = append(middlewares, authMiddleware(route.Firewalls))
	}
	return middlewares
}
