package config

import (
	"sync"

	"github.com/spf13/viper"
)

// Global App Config
type AppConfig struct {
	Help  string
	Files struct {
		Log     string
		Metrics string
	}
	Opentracing struct {
		ServiceName        string
		LocalAgentHostPort string
		LogSpans           bool
	}
	HTTPHealth struct {
		Port string
	}
	Auth struct {
	}
	Center *Config
	Worker *Config
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
