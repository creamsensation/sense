package sense

import (
	"context"
	"net/http"
	"sync"

	"github.com/creamsensation/auth"
	"github.com/creamsensation/cache"
	"github.com/creamsensation/cookie"
	"github.com/creamsensation/filesystem"
	"github.com/creamsensation/mailer"
	"github.com/creamsensation/quirk"
	"github.com/creamsensation/socketer"
)

type Context interface {
	Auth(dbname ...string) auth.Manager
	Cache() cache.Client
	Cookie() cookie.Cookie
	Config() Config
	Continue() error
	Db(dbname ...string) *quirk.Quirk
	Email() mailer.Mailer
	Files() filesystem.Client
	Lang() LangContext
	Parse() ParseContext
	Request() RequestContext
	Send() SendContext
	Translate(key string, args ...map[string]any) string
}

type handlerContext struct {
	context.Context
	config  Config
	res     http.ResponseWriter
	req     *http.Request
	mu      *sync.Mutex
	cookie  cookie.Cookie
	files   filesystem.Client
	lang    lang
	parse   *parser
	request *request
	send    *sender
}

func createHandlerContext(
	config Config, res http.ResponseWriter, req *http.Request, interceptor *interceptor, ws map[string]socketer.Ws,
) *handlerContext {
	ctx := context.Background()
	hc := &handlerContext{
		Context: ctx,
		config:  config,
		res:     res,
		req:     req,
		mu:      &sync.Mutex{},
		cookie:  cookie.New(req, res, formatPath(config.Router.Prefix)+"/"),
		files:   filesystem.New(ctx, config.Filesystem),
		parse:   &parser{req: req, limit: config.Parser.Limit},
		request: &request{req: req},
	}
	hc.lang = lang{config: config.Localization, cookie: hc.cookie}
	hc.send = &sender{
		request: hc.request, res: res, statusCode: http.StatusOK, interceptor: interceptor, ws: ws,
		auth: hc.Auth(),
	}
	return hc
}

func (c *handlerContext) Auth(dbname ...string) auth.Manager {
	var db *quirk.DB
	var ok bool
	dbn := Main
	if len(dbname) > 0 {
		dbn = dbname[0]
	}
	if len(c.config.Database) > 0 {
		db, ok = c.config.Database[dbn]
		if !ok {
			panic(ErrorInvalidDatabase)
		}
	}
	return auth.New(
		db,
		c.req,
		c.res,
		c.cookie,
		c.Cache(),
		c.config.Security.Auth,
	)
}

func (c *handlerContext) Cache() cache.Client {
	return cache.New(c.Context, c.config.Cache.Memory, c.config.Cache.Redis)
}

func (c *handlerContext) Config() Config {
	return c.config
}

func (c *handlerContext) Cookie() cookie.Cookie {
	return c.cookie
}

func (c *handlerContext) Continue() error {
	c.mu.Unlock()
	return nil
}

func (c *handlerContext) Db(dbname ...string) *quirk.Quirk {
	dbn := Main
	if len(dbname) > 0 {
		dbn = dbname[0]
	}
	db, ok := c.config.Database[dbn]
	if !ok {
		panic(ErrorInvalidDatabase)
	}
	return quirk.New(db)
}

func (c *handlerContext) Email() mailer.Mailer {
	return mailer.New(c.config.Smtp)
}

func (c *handlerContext) Files() filesystem.Client {
	return c.files
}

func (c *handlerContext) Lang() LangContext {
	return c.lang
}

func (c *handlerContext) Parse() ParseContext {
	return c.parse
}

func (c *handlerContext) Request() RequestContext {
	return c.request
}

func (c *handlerContext) Send() SendContext {
	return c.send
}

func (c *handlerContext) Translate(key string, args ...map[string]any) string {
	return key
}
