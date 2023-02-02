package api

import (
	"aurora/internal/faas"
	"aurora/internal/log"
	"aurora/internal/request"
)

type faasHandler struct{}

func (*faasHandler) ListInstance(wait *WaitConn) {
	log.Runtime().Infof("%s", wait.GetRoute())
	defer func() { wait.Done() }()
	var allInstances []*request.FaasResponse
	for _, fs := range faas.ExtantFaasMap {
		instances, err := fs.List()
		if err != nil {
			log.Runtime().Infof("Can`t obtain faas(%s) instances, err: %v", fs.GetConfig().Driver, err)
			continue
		}
		allInstances = append(allInstances, instances...)
	}
	wait.SetResult("", allInstances)
}
