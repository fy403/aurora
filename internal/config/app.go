package config

import (
	"sync"

	"github.com/gorilla/sessions"
	"github.com/spf13/viper"
)

// Global App Config
type AppConfig struct {
	Help        string
	Files       *Files
	Opentracing *Opentracing
	HTTP        *HTTP
	Auth        *Auth
	Center      *Config
	Worker      *Config
	Web         *Web
}

type Files struct {
	Log     string
	Metrics string
}
type Opentracing struct {
	ServiceName       string
	CollectorEndpoint string
	LogSpans          bool
}
type HTTP struct {
	Port string
}
type Auth struct {
	New_authentication_key string
	New_encryption_key     string
	Old_authentication_key string
	Old_encryption_key     string
	DefaultSessionOption   *sessions.Options
	Users                  map[string]string
}
type Web struct {
	StaticFS bool
	WebIndex string
	WebAddr  string
	FilePath string
}

var defaultAppConfig *AppConfig
var defaultAppConfigOnce sync.Once

func GetAppConfig() *AppConfig {
	return defaultAppConfig
}

func AppInitConfig() (err error) {
	defaultAppConfigOnce.Do(func() {
		defaultAppConfig = &AppConfig{}
		err = viper.Unmarshal(defaultAppConfig)
	})
	return
}
