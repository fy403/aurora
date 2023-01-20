package model

import (
	"aurora/internal/log"
	"aurora/internal/request"
	"time"
)

func init() {
	ExtantTaskMap["long_running_task"] = &request.Handler{
		Name:  "long_running_task",
		Usage: "默认长时间运行指定时间,接受一个int64类型参数作为默认时间,单位秒,返回error",
		Fn:    LongRunningTask,
	}
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
