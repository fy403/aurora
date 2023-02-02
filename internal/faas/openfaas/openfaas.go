package openfaas

import (
	"aurora/internal/common"
	"aurora/internal/config"
	"aurora/internal/faas/iface"
	"aurora/internal/request"
	"net/http"
	"sync"
)

type OpenFaas struct {
	common.Faas
	client *http.Client
	once   sync.Once
}

// 应答结构
type Result struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func New(cnf *config.Faas) (iface.Faas, error) {
	return &OpenFaas{
		Faas:   common.NewFaas(cnf),
		client: &http.Client{},
		once:   sync.Once{},
	}, nil
}

func (of *OpenFaas) List() ([]*request.FaasResponse, error) {
	return nil, nil
}

func (of *OpenFaas) Invoke(functionName string) (string, error) {
	return "", nil
}
