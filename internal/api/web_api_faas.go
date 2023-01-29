package api

import (
	"aurora/internal/faas"
	"aurora/internal/log"
	"aurora/internal/request"
	"fmt"
)

type faasHandler struct{}

func (*faasHandler) ListInstance(wait *WaitConn) {
	log.Runtime().Infof("%s", wait.GetRoute())
	defer func() { wait.Done() }()
	var allInstances []*request.OFDBResponse
	for _, fs := range faas.ExtantFaasMap {
		instances, err := fs.List()
		if err != nil {
			log.Runtime().Infof("Can`t obtain faas(%s) instances", fs.GetConfig().Driver)
			continue
		}
		allInstances = append(allInstances, instances...)
	}
	wait.SetResult("", allInstances)
}

func (*faasHandler) InstanceSupportedLangs(wait *WaitConn) {
	log.Runtime().Infof("%s", wait.GetRoute())
	defer func() { wait.Done() }()
	allSupportedLangs := make(map[string][]string)
	for _, fs := range faas.ExtantFaasMap {
		langs, err := fs.SupportedLang()
		if err != nil {
			log.Runtime().Infof("Can`t obtain faas(%s) langs", fs.GetConfig().Driver)
			continue
		}
		allSupportedLangs[fs.GetConfig().Driver] = langs
	}
	wait.SetResult("", allSupportedLangs)
}

func (*faasHandler) CreateInstance(wait *WaitConn, req struct {
	Driver string `json:"driver"`
	Name   string `json:"name"`
	Lang   string `json:"lang"`
}) {
	log.Runtime().Infof("%s", wait.GetRoute())
	defer func() { wait.Done() }()
	fs, ok := faas.ExtantFaasMap[req.Driver]
	if !ok {
		wait.SetResult(fmt.Sprintf("This driver(%s) not found", req.Driver), "")
		return
	}
	err := fs.New(req.Name, req.Lang, fs.GetConfig().Driver)
	if err != nil {
		log.Runtime().Infof("Can`t create faas instance, err: %v", err)
		wait.SetResult(err.Error(), "")
		return
	}
}
func (*faasHandler) WriteInstance(wait *WaitConn, req struct {
	Driver       string `json:"driver"`
	Name         string `json:"name"`
	Lang         string `json:"lang"`
	UUID         string `json:"uuid"`
	Content      []byte `json:"content"`
	Dependencies []byte `json:"dependencies"`
}) {
	log.Runtime().Infof("%s", wait.GetRoute())
	defer func() { wait.Done() }()
	fs, ok := faas.ExtantFaasMap[req.Driver]
	if !ok {
		wait.SetResult(fmt.Sprintf("This driver(%s) not found", req.Driver), "")
		return
	}
	err := fs.Write(req.UUID, req.Name, req.Lang, req.Content, req.Dependencies)
	if err != nil {
		log.Runtime().Infof("Can`t write faas instance, err: %v", err)
		wait.SetResult(err.Error(), "")
		return
	}
}
func (*faasHandler) UpInstance(wait *WaitConn, req struct {
	Driver string `json:"driver"`
	Name   string `json:"name"`
	UUID   string `json:"uuid"`
}) {
	log.Runtime().Infof("%s", wait.GetRoute())
	defer func() { wait.Done() }()
	fs, ok := faas.ExtantFaasMap[req.Driver]
	if !ok {
		wait.SetResult(fmt.Sprintf("This driver(%s) not found", req.Driver), "")
		return
	}
	err := fs.Up(req.UUID, req.Name)
	if err != nil {
		log.Runtime().Infof("Can`t up faas instance, err: %v", err)
		wait.SetResult(err.Error(), "")
		return
	}
}
func (*faasHandler) DeleteInstance(wait *WaitConn, req struct {
	Driver string `json:"driver"`
	Name   string `json:"name"`
	UUID   string `json:"uuid"`
}) {
	log.Runtime().Infof("%s", wait.GetRoute())
	defer func() { wait.Done() }()
	fs, ok := faas.ExtantFaasMap[req.Driver]
	if !ok {
		wait.SetResult(fmt.Sprintf("This driver(%s) not found", req.Driver), "")
		return
	}
	err := fs.Delete(req.UUID, req.Name)
	if err != nil {
		log.Runtime().Infof("Can`t delete faas instance, err: %v", err)
		wait.SetResult(err.Error(), "")
		return
	}
}
