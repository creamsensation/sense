package sense

import (
	"net/http"
	"strings"

	"github.com/creamsensation/sense/config"
	"github.com/creamsensation/socketer"
)

type Router interface {
	Hook() Hook
	Intercept() Interceptor
	Static(path, dir string) Router
	Use(handler Handler) Router
	Group(pathPrefix string) Router
	Get(path string, handler Handler)
	Post(path string, handler Handler)
	Put(path string, handler Handler)
	Patch(path string, handler Handler)
	Delete(path string, handler Handler)
	Ws(path, name string, handler Handler)
}

type Route struct {
	Method    string
	Path      string
	Firewalls []config.Firewall
}

type router struct {
	config      Config
	mux         *http.ServeMux
	hook        *hook
	interceptor *interceptor
	pathPrefix  string
	middlewares []Handler
	routes      *[]Route
	ws          map[string]socketer.Ws
}

func createRouter(
	config Config, mux *http.ServeMux, routes *[]Route, pathPrefix string, middlewares []Handler,
) *router {
	return &router{
		config:      config,
		mux:         mux,
		hook:        &hook{},
		interceptor: &interceptor{},
		pathPrefix:  pathPrefix,
		middlewares: middlewares,
		routes:      routes,
		ws:          make(map[string]socketer.Ws),
	}
}

func (r *router) Hook() Hook {
	return r.hook
}

func (r *router) Intercept() Interceptor {
	return r.interceptor
}

func (r *router) Use(handler Handler) Router {
	r.middlewares = append(r.middlewares, handler)
	return r
}

func (r *router) Static(path, dir string) Router {
	path = formatPath(path) + "/"
	r.mux.Handle(http.MethodGet+" "+path, http.StripPrefix(path, http.FileServer(http.Dir(dir))))
	return r
}

func (r *router) Group(pathPrefix string) Router {
	return createRouter(r.config, r.mux, r.routes, r.pathPrefix+formatPath(pathPrefix), r.middlewares)
}

func (r *router) Ws(path, name string, handler Handler) {
	r.ws[name] = socketer.New()
	route := r.addRoute("WS", path)
	r.mux.HandleFunc(
		createRoutePattern("", r.pathPrefix, path),
		createWsHandlerFunc(r.config, route, handler, r.middlewares, r.interceptor, r.ws, name),
	)
}

func (r *router) Get(path string, handler Handler) {
	route := r.addRoute(http.MethodGet, path)
	r.mux.HandleFunc(
		createRoutePattern(http.MethodGet, r.pathPrefix, path),
		createHandlerFunc(r.config, route, handler, r.middlewares, r.hook, r.interceptor, r.ws),
	)
}

func (r *router) Post(path string, handler Handler) {
	route := r.addRoute(http.MethodPost, path)
	r.mux.HandleFunc(
		createRoutePattern(http.MethodPost, r.pathPrefix, path),
		createHandlerFunc(r.config, route, handler, r.middlewares, r.hook, r.interceptor, r.ws),
	)
}

func (r *router) Put(path string, handler Handler) {
	route := r.addRoute(http.MethodPut, path)
	r.mux.HandleFunc(
		createRoutePattern(http.MethodPut, r.pathPrefix, path),
		createHandlerFunc(r.config, route, handler, r.middlewares, r.hook, r.interceptor, r.ws),
	)
}

func (r *router) Patch(path string, handler Handler) {
	route := r.addRoute(http.MethodPatch, path)
	r.mux.HandleFunc(
		createRoutePattern(http.MethodPatch, r.pathPrefix, path),
		createHandlerFunc(r.config, route, handler, r.middlewares, r.hook, r.interceptor, r.ws),
	)
}

func (r *router) Delete(path string, handler Handler) {
	route := r.addRoute(http.MethodDelete, path)
	r.mux.HandleFunc(
		createRoutePattern(http.MethodDelete, r.pathPrefix, path),
		createHandlerFunc(r.config, route, handler, r.middlewares, r.hook, r.interceptor, r.ws),
	)
}

func (r *router) addRoute(method string, path string) Route {
	p := r.pathPrefix + formatPath(path)
	n := len(p)
	if n == 0 {
		p += "/"
	}
	if p != "/" {
		p = strings.TrimSuffix(p, "/")
	}
	route := Route{
		Method:    method,
		Path:      p,
		Firewalls: findFirewallsWithPath(p, r.config.Security.Firewalls),
	}
	*r.routes = append(*r.routes, route)
	return route
}
