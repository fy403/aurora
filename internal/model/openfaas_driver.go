package model

import (
	"aurora/internal/request"
)

func init() {
	ExtantTaskMap["openfaas_driver"] = &request.Handler{
		Usage: "openfaas的驱动句柄: in[0]为函数名字, in[1]为string; 返回string, error",
		Fn:    OpenfaasDriver,
	}
}

func OpenfaasDriver(functionName string, args string) (string, error) {
	// resp, err := http.Post(url, "application/json; charset=utf-8", strings.NewReader(args))
	// if err != nil {
	// 	return "", err
	// }
	// if resp.Body != nil {
	// 	defer resp.Body.Close()
	// }
	// body, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	return "", err
	// }
	// return string(body), nil
	return "", nil
}
