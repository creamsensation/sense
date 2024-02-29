package config

import "github.com/creamsensation/translator"

type Localization struct {
	Enabled    bool
	Languages  []Language
	Translator translator.Translator
}

type Language struct {
	Main bool
	Code string
}
