package config

import (
	"github.com/creamsensation/translator"
	"github.com/creamsensation/validator"
)

type Localization struct {
	Enabled    bool
	Languages  []Language
	Translator translator.Translator
	Validator  validator.Messages
}

type Language struct {
	Main bool
	Code string
}
