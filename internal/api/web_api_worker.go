package api

import (
	"aurora/internal/log"
	"aurora/internal/request"
)

type workerHandler struct{}

func (*workerHandler) list(wait *WaitConn) {
	log.Runtime().Infof("%s", wait.GetRoute())
	defer func() { wait.Done() }()
	results, err := defaultApi.server.GetBackend().GetAllWorkersInfo()
	var filterResults []*request.WorkerResponse
	// Purge invalid worker
	for idx, result := range results {
		if isValid := result.IsValid(defaultApi.cfg.Center.BrokerApi); !isValid {
			results[idx] = nil
			req := request.WorkerRequest{
				UUID: result.UUID,
			}
			defaultApi.server.GetBackend().PurgeWorkerInfo(&req)
			continue
		}
		filterResults = append(filterResults, result)
	}
	if err != nil {
		wait.SetResult(err.Error(), "")
	}
	wait.SetResult("", filterResults)
}
