package sense

import (
	"github.com/creamsensation/filesystem"
	
	"github.com/creamsensation/quirk"
	"github.com/creamsensation/sense/config"
	
	"github.com/creamsensation/mailer"
)

type Config struct {
	App          config.App
	Cache        config.Cache
	Database     map[string]*quirk.DB
	Export       config.Export
	Filesystem   filesystem.Config
	Localization config.Localization
	Parser       config.Parser
	Router       config.Router
	Security     config.Security
	Smtp         mailer.Config
}

const (
	Main = "main"
)
