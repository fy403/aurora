package model

import (
	"aurora/internal/faas"
	"aurora/internal/request"
	"errors"
)

func init() {
	ExtantTaskMap["openfaas_driver"] = &request.Handler{
		Usage: "openfaas的驱动句柄: in[0]为函数名字, in[1]为string; 返回string, error",
		Fn:    OpenfaasDriver,
	}
}

func OpenfaasDriver(functionName string, args string) (string, error) {
	driver := "openfaas"
	// 调用函数。
	afc, ok := faas.ExtantFaasMap[driver]
	if !ok {
		return "", errors.New("Not found faas instance: " + driver)
	}
	return afc.Invoke(functionName, args)
}
