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
	Gateway     *Config
	Worker      *Config
	Faas        []*Faas
}

type Faas struct {
	Driver   string
	Endpoint string
	// aliyunfc
	ServiceName     string
	AccessKeyId     string
	AccessKeySecret string
	// openfaas
	Name     string
	Password string
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
	NewAuthenticationKey string
	NewEncryptionKey     string
	OldAuthenticationKey string
	OldEncryptionKey     string
	DefaultSessionOption *sessions.Options
	Users                map[string]string
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
