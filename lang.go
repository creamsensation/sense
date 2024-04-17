package sense

import (
	"time"
	
	"github.com/creamsensation/cookie"
	"github.com/creamsensation/sense/config"
)

type LangContext interface {
	Exists() bool
	Get() string
	Set(langCode string)
	CreateIfNotExists()
}

type lang struct {
	config config.Localization
	cookie cookie.Cookie
}

const (
	LangCookieKey = "X-Lang"
)

const (
	langDuration = 365 * 24 * time.Hour
)

func (l lang) Exists() bool {
	if !l.config.Enabled {
		return false
	}
	return len(l.Get()) > 0
}

func (l lang) CreateIfNotExists() {
	if !l.config.Enabled {
		return
	}
	langCode := l.cookie.Get(LangCookieKey)
	if len(langCode) != 0 {
		return
	}
	langCode = l.getMainLangCode()
	l.Set(langCode)
}

func (l lang) Get() string {
	if !l.config.Enabled {
		return ""
	}
	langCode := l.cookie.Get(LangCookieKey)
	return langCode
}

func (l lang) Set(langCode string) {
	if !l.config.Enabled {
		return
	}
	if !l.langCodeExists(langCode) {
		panic(ErrorInvalidLang)
	}
	l.cookie.Set(LangCookieKey, langCode, langDuration)
}

func (l lang) getMainLangCode() string {
	for _, item := range l.config.Languages {
		if item.Main {
			return item.Code
		}
	}
	return ""
}

func (l lang) langCodeExists(langCode string) bool {
	exists := false
	for _, item := range l.config.Languages {
		if item.Code == langCode {
			exists = true
		}
	}
	return exists
}
