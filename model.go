package sense

import (
	"net/http"
	
	"github.com/creamsensation/socketer"
)

type Assert interface {
	string | int | float32 | float64 | bool
}

type routerArgs struct {
	config      Config
	mux         *http.ServeMux
	routes      *[]Route
	pathPrefix  string
	middlewares []Handler
}

type handlerFuncArgs struct {
	config      Config
	route       Route
	handler     Handler
	middlewares []Handler
	ws          map[string]socketer.Ws
	name        string
}

type handlerContextArgs struct {
	config Config
	req    *http.Request
	res    http.ResponseWriter
	ws     map[string]socketer.Ws
}
