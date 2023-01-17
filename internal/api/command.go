package api

import (
	"aurora/internal/log"
)

var defaultApi *Api

func PreRun() error {
	log.Runtime().Debug("Web PreRun started")
	defaultApi = NewApi()
	err := defaultApi.Init()
	if err != nil {
		log.Runtime().Errorf("Web init faild: %s", err.Error())
	}
	return err
}

func Run() error {
	log.Runtime().Debug("Web Run started")
	err := defaultApi.Run()
	if err != nil {
		log.Runtime().Errorf("Web run faild: %s", err.Error())
	}
	return err
}

func PostRun() error {
	log.Runtime().Debug("Web PostRun started")
	return defaultApi.Stop()
}
