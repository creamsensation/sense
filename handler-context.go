package sense

import (
	"context"
	"net/http"
	"sync"
	
	"github.com/creamsensation/filesystem"
	"github.com/creamsensation/validator"
	
	"github.com/creamsensation/auth"
	"github.com/creamsensation/cookie"
	
	"github.com/creamsensation/cache"
	"github.com/creamsensation/mailer"
	"github.com/creamsensation/quirk"
)

type Context interface {
	Auth(dbname ...string) auth.Manager
	Cache() cache.Client
	Cookie() cookie.Cookie
	Config() Config
	Continue() error
	Db(dbname ...string) *quirk.Quirk
	Email() mailer.Mailer
	Export() ExportContext
	Files() filesystem.Client
	Lang() LangContext
	Parse() ParseContext
	Request() RequestContext
	Send() SendContext
	Translate(key string, args ...map[string]any) string
	Validate(s validator.Schema, data any) (bool, ErrorsWrapper[validator.Errors])
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

func createHandlerContext(args handlerContextArgs) *handlerContext {
	ctx := context.Background()
	hc := &handlerContext{
		Context: ctx,
		config:  args.config,
		res:     args.res,
		req:     args.req,
		mu:      &sync.Mutex{},
		cookie:  cookie.New(args.req, args.res, formatPath(args.config.Router.Prefix)+"/"),
		files:   filesystem.New(ctx, args.config.Filesystem),
		parse:   &parser{req: args.req, limit: args.config.Parser.Limit},
		request: &request{req: args.req},
	}
	hc.lang = lang{config: args.config.Localization, cookie: hc.cookie}
	hc.send = &sender{
		request:    hc.request,
		res:        args.res,
		statusCode: http.StatusOK,
		ws:         args.ws,
		auth:       hc.Auth(),
	}
	hc.Lang().CreateIfNotExists()
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

func (c *handlerContext) Export() ExportContext {
	return createExport(c.config.Export)
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
	if !c.config.Localization.Enabled {
		return key
	}
	return c.config.Localization.Translator.Translate(c.Lang().Get(), key, args...)
}

func (c *handlerContext) Validate(s validator.Schema, data any) (bool, ErrorsWrapper[validator.Errors]) {
	m := c.config.Localization.Validator
	var messages validator.Messages
	if !c.config.Localization.Enabled {
		messages = validator.Messages{
			Email:     m.Email,
			Required:  m.Required,
			MinText:   m.MinText,
			MaxText:   m.MaxText,
			MinNumber: m.MinNumber,
			MaxNumber: m.MaxNumber,
		}
	}
	if c.config.Localization.Enabled {
		messages = validator.Messages{
			Email:     c.Translate(m.Email),
			Required:  c.Translate(m.Required),
			MinText:   c.Translate(m.MinText),
			MaxText:   c.Translate(m.MaxText),
			MinNumber: c.Translate(m.MinNumber),
			MaxNumber: c.Translate(m.MaxNumber),
		}
	}
	v := validator.New(
		validator.Config{
			Messages: messages,
		},
	)
	ok, errs := v.Json(s, data)
	return ok, ErrorsWrapper[validator.Errors]{errs}
}
