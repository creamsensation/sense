package sense

import (
	"errors"
	"net/http"
	"slices"
	"strings"

	"github.com/creamsensation/sense/config"
)

func authMiddleware(firewalls []config.Firewall) Handler {
	return func(c Context) error {
		if len(firewalls) == 0 {
			return c.Continue()
		}
		session, err := c.Auth().Session().Get()
		if err != nil || session.Id == 0 {
			return c.Send().Status(http.StatusForbidden).Error(errors.New(http.StatusText(http.StatusForbidden)))
		}
		allowed := session.Super
		if !session.Super {
			for _, f := range firewalls {
				if len(f.Roles) == 0 {
					allowed = true
					continue
				}
				if slices.ContainsFunc(
					f.Roles, func(firewallRole string) bool {
						return slices.Contains(session.Roles, firewallRole)
					},
				) {
					allowed = true
				}
			}
		}
		if !allowed {
			return c.Send().Status(http.StatusForbidden).Error(errors.New(http.StatusText(http.StatusForbidden)))
		}
		if allowed {
			if err := c.Auth().Session().Renew(); err != nil {
				return c.Send().Status(http.StatusInternalServerError).Error(err)
			}
		}
		return c.Continue()
	}
}

func trailingSlashMiddleware() Handler {
	return func(c Context) error {
		path := c.Request().Path()
		if strings.HasSuffix(path, "/") {
			return c.Continue()
		}
		return c.Send().Redirect(path + "/")
	}
}
