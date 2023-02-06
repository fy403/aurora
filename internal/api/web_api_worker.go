package api

import (
	"aurora/internal/log"
	"aurora/internal/request"
	workerutils "aurora/internal/utils/worker"
)

type workerHandler struct{}

func (*workerHandler) list(wait *WaitConn) {
	log.Runtime().Infof("%s", wait.GetRoute())
	defer func() { wait.Done() }()
	results, err := workerutils.GetAllWorkersInfo(defaultApi.server.GetCache())
	var filterResults []*request.WorkerResponse
	// Purge invalid worker
	for idx, result := range results {
		if isValid := result.IsValid(defaultApi.cfg.Gateway.BrokerApi); !isValid {
			results[idx] = nil
			req := request.WorkerRequest{
				UUID: result.UUID,
			}
			workerutils.PurgeWorkerInfo(defaultApi.server.GetCache(), &req)
			continue
		}
		filterResults = append(filterResults, result)
	}
	if err != nil {
		wait.SetResult(err.Error(), "")
	}
	wait.SetResult("", filterResults)
}
