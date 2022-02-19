package center

import (
	"aurora/internal/log"
)

var defaultCenter *Center

func PreRun() error {
	log.Runtime().Debug("Center PreRun started")
	defaultCenter = NewCenter()
	err := defaultCenter.Init()
	if err != nil {
		log.Runtime().Errorf("Center init faild: %s", err.Error())
	}
	return err
}

func Run() error {
	log.Runtime().Debug("Center Run started")
	err := defaultCenter.Run()
	if err != nil {
		log.Runtime().Errorf("Center run faild: %s", err.Error())
	}
	return err
}

func PostRun() error {
	log.Runtime().Debug("Center PostRun started")
	return defaultCenter.Stop()
}
