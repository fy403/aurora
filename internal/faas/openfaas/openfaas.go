package openfaas

import (
	"aurora/internal/common"
	"aurora/internal/config"
	"aurora/internal/faas/iface"
	"aurora/internal/request"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type OpenFaas struct {
	common.Faas
	client *http.Client
	once   sync.Once
}

type ListFunctionsOutput struct {
	Name        string            `json:"name"`
	Image       string            `json:"image"`
	Namespace   string            `json:"namespace"`
	EnvProcess  string            `json:"envProcess"`
	EnvVars     map[string]string `json:"envVars"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	CreatedAt   string            `createdAt:""`
}

func New(cnf *config.Faas) (iface.Faas, error) {
	return &OpenFaas{
		Faas:   common.NewFaas(cnf),
		client: &http.Client{},
		once:   sync.Once{},
	}, nil
}

func (of *OpenFaas) List() ([]*request.FaasResponse, error) {
	endpoint, _ := GetAccessPoint(of.GetConfig().Endpoint)
	req, err := http.NewRequest(http.MethodGet, endpoint+"/system/functions", http.NoBody)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(of.GetConfig().Name, of.GetConfig().Password)
	res, err := of.client.Do(req)
	if err != nil {
		return nil, err
	}
	if res.Body != nil {
		defer res.Body.Close()
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	listFunctionsOutput := make([]*ListFunctionsOutput, 0)
	err = json.Unmarshal(resBody, &listFunctionsOutput)
	if err != nil {
		return nil, err
	}
	var resps = make([]*request.FaasResponse, 0, len(listFunctionsOutput))
	for _, fcMeta := range listFunctionsOutput {
		resp := &request.FaasResponse{
			FunctionID:   uuid.New().String(),
			FunctionName: fcMeta.Name,
			Driver:       of.GetConfig().Driver,
			Description:  "OpenFaas接入测试成功通知",
			Timestamp:    time.Now().Unix(),
		}
		resps = append(resps, resp)
	}
	return resps, nil
}

func (of *OpenFaas) Invoke(functionName string, args string) (string, error) {
	endpoint, _ := GetAccessPoint(of.GetConfig().Endpoint)
	req, err := http.NewRequest(http.MethodGet, endpoint+"/function/"+functionName, strings.NewReader(args))
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(of.GetConfig().Name, of.GetConfig().Password)
	res, err := of.client.Do(req)
	if err != nil {
		return "", err
	}
	if res.Body != nil {
		defer res.Body.Close()
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	// mongodb直接存储该结果会以object形式，因此需要编码
	sEnc := base64.StdEncoding.EncodeToString(resBody)
	return sEnc, nil
}

func GetAccessPoint(endpointInput string) (endpoint, host string) {
	httpPrefix := "http://"
	httpsPrefix := "https://"
	if HasPrefix(endpointInput, httpPrefix) {
		host = endpointInput[len(httpPrefix):]
		return endpointInput, host
	} else if HasPrefix(endpointInput, httpsPrefix) {
		host = endpointInput[len(httpsPrefix):]
		return endpointInput, host
	}
	return httpPrefix + endpointInput, endpointInput
}

// HasPrefix check endpoint prefix
func HasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[0:len(prefix)] == prefix
}
