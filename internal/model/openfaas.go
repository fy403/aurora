package model

import (
	"aurora/internal/request"
	"io/ioutil"
	"net/http"
	"strings"
)

func init() {
	ExtantTaskMap["openfaas_driver"] = &request.Handler{
		Usage: "openfaas的驱动句柄: in[0]为url, in[1]为string; 返回string, error",
		Fn:    OpenfaasDriver,
	}
}

func OpenfaasDriver(url string, args string) (string, error) {
	resp, err := http.Post(url, "application/json; charset=utf-8", strings.NewReader(args))
	if err != nil {
		return "", err
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
