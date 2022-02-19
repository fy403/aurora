package worker

import (
	"aurora/internal/log"
)

var defaultWorker *Worker

func PreRun() error {
	log.Runtime().Debug("Worker PreRun started")
	defaultWorker = &Worker{}
	err := defaultWorker.Init()
	if err != nil {
		log.Runtime().Errorf("Worker init faild: %s", err.Error())
	}
	return err
}

func Run() error {
	log.Runtime().Debug("Worker Run started")
	err := defaultWorker.Run()
	if err != nil {
		log.Runtime().Errorf("Worker run faild: %s", err.Error())
	}
	return err
}

func PostRun() error {
	log.Runtime().Debug("Worker PostRun started")
	return defaultWorker.Stop()
}
