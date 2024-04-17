# Sense
Simple and intuitive Go web framework

**IMPORTANT: API is not frozen, bugs expected!**

## Features

- CRUD
- Static files
- Groups
- URL, body parser
- Middlewares
- Error recovery
- Auth
- DBAL
- Cache
- Filesystem
- Localization
- Data exports
- Validation
- Emails
- Websockets

## Examples

### Config
```go
config := sense.Config{
    App: config.App{
        Name: "example",
    },
    Cache: config.Cache{
        Memory: memory.New("./cache")
        Redis: redis.New(&redis.Options{...})
    },
    Database: map[string]*quirk.DB{
        sense.Main: quirk.MustConnect(...)
    },
    Export: config.Export{
        Gotenberg: config.Gotenberg{
            Endpoint: "http://localhost:3000",
        },
    },
    Filesystem: filesystem.New(
        context.Background(),
        filesystem.Config{
            Driver: filesystem.Local,
            Dir: "./files"
        },
    ),
    Localization: config.Localization{
        Enabled: true,
        Languages: []config.Language{
            {Main: true, Code: "cs"},
            {Code: "en"},
        },
        Translator: translator.New(
            translator.Config{
                Dir:      "./static/locales",
                FileType: translator.Yaml,
            },
        ),
        Validator: validator.Messages{
            Email:     "error.field.email",
            Required:  "error.field.required",
            MinText:   "error.field.min-text",
            MaxText:   "error.field.max-text",
            MinNumber: "error.field.min-number",
            MaxNumber: "error.field.max-number",
        },
    },
    Router: config.Router{
        Prefix:  "",
        Recover: true,
    },
    Security: config.Security{
        Auth: auth.Config{
          Roles: map[string]auth.Role{
            "owner": {Super: true},
            "sales": {},
          },
          Duration: 24 * time.Hour,
        },
        Firewalls: []config.Firewall{},
    },
    Smtp: mailer.Config{
        Host:     "smtp.example.com",
        Port:     25,
        User:     "exampleuser",
        Password: "examplepass",
    },
}
```

### New instance
```go
app := sense.New(config)
```

### Get
```go
app.Get("/{id}", func(c sense.Context) error {
    var id int
    c.Parse().MustPathValue("id", &id)
    return c.Send().Json(map[string]int{
      "id": id
    })
}) 
```

### Post
```go
app.Post("/", func(c sense.Context) error {
    type example struct {
        Id int `json:"id"`
    } 
    var body example
    c.Parse().MustJson(&body)
    return c.Send().Json(map[string]int{
      "id": body.Id,
    })
}) 
```

### Group
```go
versionOne := app.Group("/v1")
{
    user := versionOne.Group("/user")
    {
        user.Get("/", user_handler.GetAll())
        user.Get("/{id}", user_handler.GetOne())
        user.Post("/", user_handler.CreateOne())
        user.Put("/", user_handler.UpdateOne())
        user.Delete("/{id}", user_handler.RemoveOne())
    }
}
```

### Start an app
```go
app.Run(":8000")
```