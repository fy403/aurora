package model

import (
	"aurora/internal/log"
	"time"
)

func init() {
	ExtantTaskMap["long_running_task"] = LongRunningTask
}

// LongRunningTask ...
func LongRunningTask(duration int64) error {
	log.Runtime().Info("Long running task started")
	for i := int64(0); i < duration; i++ {
		log.Runtime().Info(string(10 - i))
		time.Sleep(1 * time.Second)
	}
	log.Runtime().Info("Long running task finished")
	return nil
}
