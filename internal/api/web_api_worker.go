package api

import (
	"aurora/internal/log"
	"aurora/internal/request"
)

type workerHandler struct{}

func (*workerHandler) list(wait *WaitConn) {
	log.Runtime().Infof("%s", wait.GetRoute())
	defer func() { wait.Done() }()
	results, err := defaultApi.GetServer().GetAllWorkersInfo()
	var filterResults []*request.WorkerResponse
	// Purge invalid worker
	for _, result := range results {
		// if isValid := result.IsValid(defaultApi.cfg.Gateway.BrokerApi); !isValid {
		// 	results[idx] = nil
		// 	req := request.WorkerRequest{
		// 		UUID: result.UUID,
		// 	}
		// 	defaultApi.GetServer().PurgeWorkerInfo(&req)
		// 	continue
		// }
		filterResults = append(filterResults, result)
	}
	if err != nil {
		wait.SetResult(err.Error(), "")
	}
	wait.SetResult("", filterResults)
}
