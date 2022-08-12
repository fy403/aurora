package example

import (
	"aurora/internal/log"
	"aurora/internal/model"
	"time"
)

func init() {
	model.ExtantTaskMap["long_running_task"] = LongRunningTask
}

// LongRunningTask ...
func LongRunningTask() error {
	log.Runtime().Info("Long running task started")
	for i := 0; i < 10; i++ {
		log.Runtime().Info(string(10 - i))
		time.Sleep(1 * time.Second)
	}
	log.Runtime().Info("Long running task finished")
	return nil
}
