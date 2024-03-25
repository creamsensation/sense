package sense

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
)

type Sense interface {
	Router
	Run(address string)
}

type sense struct {
	context.Context
	*router
	config Config
	mux    *http.ServeMux
	routes *[]Route
}

func New(config Config) Sense {
	mux := http.NewServeMux()
	routes := make([]Route, 0)
	return &sense{
		Context: context.Background(),
		router:  createRouter(config, mux, &routes, formatPath(config.Router.Prefix), []Handler{}),
		config:  config,
		mux:     mux,
		routes:  &routes,
	}
}

func (s *sense) Run(address string) {
	s.beforeRun(address)
	log.Fatalln(http.ListenAndServe(address, s.mux))
}

func (s *sense) beforeRun(address string) {
	s.printRoutes()
	if strings.HasPrefix(address, ":") {
		address = "localhost" + address
	}

	fmt.Printf(
		"%s%s%s %s\n",
		WhiteColor.Render("Sense ["),
		BlueColor.Render(s.config.App.Name),
		WhiteColor.Render("] running on ->"),
		BlueColor.Render(address),
	)
}

func (s *sense) printRoutes() {
	fmt.Println(WhiteColor.Underline(true).Bold(true).Render("Routes:"))
	for _, route := range *s.routes {
		fmt.Printf(
			"%s %s\n", EmeraldColor.Bold(true).Underline(false).Render(route.Method),
			WhiteColor.Bold(false).Underline(false).Render(route.Path),
		)
	}
	fmt.Println(Divider)
}
