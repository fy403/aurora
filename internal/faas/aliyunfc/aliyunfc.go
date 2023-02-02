package aliyunfc

import (
	"aurora/internal/common"
	"aurora/internal/config"
	"aurora/internal/faas/iface"
	"aurora/internal/request"
	"encoding/base64"
	"time"

	"github.com/aliyun/fc-go-sdk"
)

type AliyunFC struct {
	common.Faas
	client *fc.Client
}

// 应答结构
type Result struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func New(cnf *config.Faas) (iface.Faas, error) {
	apiVersion := "2016-08-15"
	client, err := fc.NewClient(cnf.Endpoint, apiVersion, cnf.AccessKeyId, cnf.AccessKeySecret)
	if err != nil {
		return nil, err
	}
	return &AliyunFC{
		Faas:   common.NewFaas(cnf),
		client: client,
	}, nil
}

func (afc *AliyunFC) List() ([]*request.FaasResponse, error) {
	// 获取服务。
	listFunctionsOutput, err := afc.client.ListFunctions(fc.NewListFunctionsInput(afc.GetConfig().ServiceName))
	if err != nil {
		return nil, err
	}
	resps := make([]*request.FaasResponse, 0, len(listFunctionsOutput.Functions))
	for _, fcMeta := range listFunctionsOutput.Functions {
		resp := &request.FaasResponse{
			FunctionID:   *fcMeta.FunctionID,
			FunctionName: *fcMeta.FunctionName,
			Driver:       afc.GetConfig().Driver,
			Description:  *fcMeta.Description,
			Timestamp:    time.Now().Unix(),
		}
		resps = append(resps, resp)
	}
	return resps, nil
}

func (afc *AliyunFC) Invoke(functionName string) (string, error) {
	invokeInput := fc.NewInvokeFunctionInput(afc.GetConfig().ServiceName, functionName).WithLogType("None")
	invokeOutput, err := afc.client.InvokeFunction(invokeInput)
	if err != nil {
		return "", err
	}
	// mongodb直接存储该结果会以object形式，因此需要编码
	sEnc := base64.StdEncoding.EncodeToString([]byte(invokeOutput.String()))
	return sEnc, nil
}
