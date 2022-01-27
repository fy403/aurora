package config

import (
	"sync"

	"github.com/spf13/viper"
)

// type JobConfig struct {
// 	Name              string
// 	Provide           string
// 	GatherNames       []string
// 	StoreNames        []string
// 	Gathers           []*GatherConfig
// 	Stores            []*StoreConfig
// 	Params            util.Params
// 	MinOutputInterval time.Duration
// }

// func (this *JobConfig) CheckName() {
// 	this.Name = strings.ToLower(this.Name)
// 	if this.Params == nil {
// 		this.Params = make(util.Params)
// 	}
// 	if _, ok := this.Params.GetString("JobName"); !ok {
// 		this.Params["JobName"] = this.Name
// 	}
// }

// func (this *JobConfig) DeepCopy() *JobConfig {
// 	var newCfg = *this
// 	newCfg.Params = make(util.Params).Merge(this.Params)
// 	newCfg.Gathers = []*GatherConfig{}
// 	for _, gatherCfg := range this.Gathers {
// 		newCfg.Gathers = append(newCfg.Gathers, gatherCfg.DeepCopy(this.Params))
// 	}
// 	newCfg.Stores = []*StoreConfig{}
// 	for _, storeCfg := range this.Stores {
// 		newCfg.Stores = append(newCfg.Stores, storeCfg.DeepCopy(this.Params))
// 	}
// 	return &newCfg
// }

type AppConfig struct {
	Help  string
	Files struct {
		Log     string
		Metrics string
	}
	Port string
}

var defaultConfig *AppConfig
var defaultConfigOnce sync.Once

func Config() *AppConfig {
	return defaultConfig
}

func InitConfig() error {
	defaultConfig = &AppConfig{}
	return viper.Unmarshal(defaultConfig)
}
