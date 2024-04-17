package sense

import (
	"net/http"
	"slices"
	"strings"
	
	"github.com/creamsensation/socketer"
	
	"github.com/creamsensation/sense/config"
)

type Router interface {
	Static(path, dir string) Router
	Use(handler Handler) Router
	Group(pathPrefix string) Router
	Head(path string, handler Handler)
	Get(path string, handler Handler)
	Post(path string, handler Handler)
	Options(path string, handler Handler)
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
	pathPrefix  string
	middlewares []Handler
	routes      *[]Route
	options     []string
	heads       []string
	ws          map[string]socketer.Ws
}

func createRouter(args routerArgs) *router {
	return &router{
		config:      args.config,
		mux:         args.mux,
		pathPrefix:  args.pathPrefix,
		middlewares: args.middlewares,
		routes:      args.routes,
		options:     make([]string, 0),
		heads:       make([]string, 0),
		ws:          make(map[string]socketer.Ws),
	}
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
	return createRouter(
		routerArgs{
			config:      r.config,
			mux:         r.mux,
			routes:      r.routes,
			pathPrefix:  r.pathPrefix + formatPath(pathPrefix),
			middlewares: r.middlewares,
		},
	)
}

func (r *router) Ws(path, name string, handler Handler) {
	path = formatPath(path)
	r.ws[name] = socketer.New()
	route := r.addRoute("WS", path)
	r.mux.HandleFunc(
		createRoutePattern("", r.pathPrefix, path),
		createWsHandlerFunc(
			handlerFuncArgs{
				config:      r.config,
				route:       route,
				handler:     handler,
				middlewares: r.middlewares,
				ws:          r.ws,
				name:        name,
			},
		),
	)
}

func (r *router) Get(path string, handler Handler) {
	path = formatPath(path)
	route := r.addRoute(http.MethodGet, path)
	r.createHeadHandleFunc(path, route)
	r.createCanonicalHandleFunc(http.MethodGet, path)
	r.mux.HandleFunc(
		createRoutePattern(http.MethodGet, r.pathPrefix, path),
		createHandlerFunc(
			handlerFuncArgs{
				config:      r.config,
				route:       route,
				handler:     handler,
				middlewares: r.middlewares,
			},
		),
	)
}

func (r *router) Post(path string, handler Handler) {
	path = formatPath(path)
	route := r.addRoute(http.MethodPost, path)
	r.createOptionsHandleFunc(path, route)
	r.createCanonicalHandleFunc(http.MethodPost, path)
	r.mux.HandleFunc(
		createRoutePattern(http.MethodPost, r.pathPrefix, path),
		createHandlerFunc(
			handlerFuncArgs{
				config:      r.config,
				route:       route,
				handler:     handler,
				middlewares: r.middlewares,
			},
		),
	)
}

func (r *router) Put(path string, handler Handler) {
	path = formatPath(path)
	route := r.addRoute(http.MethodPut, path)
	r.createCanonicalHandleFunc(http.MethodPut, path)
	r.mux.HandleFunc(
		createRoutePattern(http.MethodPut, r.pathPrefix, path),
		createHandlerFunc(
			handlerFuncArgs{
				config:      r.config,
				route:       route,
				handler:     handler,
				middlewares: r.middlewares,
			},
		),
	)
}

func (r *router) Patch(path string, handler Handler) {
	path = formatPath(path)
	route := r.addRoute(http.MethodPatch, path)
	r.createCanonicalHandleFunc(http.MethodPatch, path)
	r.mux.HandleFunc(
		createRoutePattern(http.MethodPatch, r.pathPrefix, path),
		createHandlerFunc(
			handlerFuncArgs{
				config:      r.config,
				route:       route,
				handler:     handler,
				middlewares: r.middlewares,
			},
		),
	)
}

func (r *router) Delete(path string, handler Handler) {
	path = formatPath(path)
	route := r.addRoute(http.MethodDelete, path)
	r.createCanonicalHandleFunc(http.MethodDelete, path)
	r.mux.HandleFunc(
		createRoutePattern(http.MethodDelete, r.pathPrefix, path),
		createHandlerFunc(
			handlerFuncArgs{
				config:      r.config,
				route:       route,
				handler:     handler,
				middlewares: r.middlewares,
			},
		),
	)
}

func (r *router) Options(path string, handler Handler) {
	path = formatPath(path)
	route := r.addRoute(http.MethodOptions, path)
	r.createCanonicalHandleFunc(http.MethodOptions, path)
	r.mux.HandleFunc(
		createRoutePattern(http.MethodOptions, r.pathPrefix, path),
		createHandlerFunc(
			handlerFuncArgs{
				config:      r.config,
				route:       route,
				handler:     handler,
				middlewares: r.middlewares,
			},
		),
	)
}

func (r *router) Head(path string, handler Handler) {
	path = formatPath(path)
	route := r.addRoute(http.MethodHead, path)
	r.createCanonicalHandleFunc(http.MethodHead, path)
	r.mux.HandleFunc(
		createRoutePattern(http.MethodHead, r.pathPrefix, path),
		createHandlerFunc(
			handlerFuncArgs{
				config:      r.config,
				route:       route,
				handler:     handler,
				middlewares: r.middlewares,
			},
		),
	)
}

func (r *router) createCanonicalHandleFunc(method string, path string) {
	if !strings.HasSuffix(path, "/") {
		r.mux.HandleFunc(
			createRoutePattern(method, r.pathPrefix, path+"/"),
			createHandlerCanonicalRedirect(),
		)
	}
}

func (r *router) createOptionsHandleFunc(path string, route Route) {
	if slices.Contains(r.options, route.Path) {
		return
	}
	r.options = append(r.options, route.Path)
	r.createCanonicalHandleFunc(http.MethodOptions, path)
	r.mux.HandleFunc(
		createRoutePattern(http.MethodOptions, r.pathPrefix, path),
		createEmptyHandlerResponse(),
	)
}

func (r *router) createHeadHandleFunc(path string, route Route) {
	if slices.Contains(r.heads, route.Path) {
		return
	}
	r.heads = append(r.heads, route.Path)
	r.createCanonicalHandleFunc(http.MethodHead, path)
	r.mux.HandleFunc(
		createRoutePattern(http.MethodHead, r.pathPrefix, path),
		createEmptyHandlerResponse(),
	)
}

func (r *router) addRoute(method string, path string) Route {
	p := r.pathPrefix + path
	route := Route{
		Method:    method,
		Path:      p,
		Firewalls: findFirewallsWithPath(p, r.config.Security.Firewalls),
	}
	*r.routes = append(*r.routes, route)
	return route
}
