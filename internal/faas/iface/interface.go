package iface

import (
	"aurora/internal/config"
	"aurora/internal/request"
)

type Faas interface {
	GetConfig() *config.Faas
	List() ([]*request.FaasResponse, error)
	Invoke(string, string) (string, error)
}
