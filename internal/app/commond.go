package app

import (
	"aurora/internal/log"
)

var defaultApp *App

func PreRun() error {
	log.Runtime().Debug("aurora PreRun started")
	defaultApp = NewApp()
	err := defaultApp.Init()
	if err != nil {
		log.Runtime().Errorf("app init faild: %s", err.Error())
	}
	return err
}

func Run() error {
	log.Runtime().Debug("aurora Run started")
	err := defaultApp.Run()
	if err != nil {
		log.Runtime().Errorf("app run faild: %s", err.Error())
	}
	return err
}

func PostRun() error {
	log.Runtime().Debug("aurora PostRun started")
	return defaultApp.Stop()
}
