package openfaas

import (
	backend_iface "aurora/internal/backends/iface"
	"aurora/internal/common"
	"aurora/internal/config"
	"aurora/internal/faas/iface"
	"aurora/internal/log"
	"aurora/internal/request"
	"aurora/internal/utils"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type OpenFaas struct {
	common.Faas
	backend backend_iface.Backend
	client  *http.Client
	cookies []*http.Cookie
	once    sync.Once
}

// 应答结构
type Result struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func New(cnf *config.Faas, bk backend_iface.Backend) (iface.Faas, error) {
	return &OpenFaas{
		Faas:    common.NewFaas(cnf),
		client:  &http.Client{},
		once:    sync.Once{},
		backend: bk,
	}, nil
}

func (of *OpenFaas) New(name, lang, prefix string, opts ...*iface.NewOptions) error {
	url := fmt.Sprintf("%s/api/faasd/new", of.GetConfig().Endpoint)

	data := struct {
		Name   string `json:"name"`
		Lang   string `json:"lang"`
		Prefix string `json:"prefix"`
	}{
		Name:   name,
		Lang:   lang,
		Prefix: prefix,
	}
	ret, err := of.send(url, "POST", &data)
	if err != nil {
		return err
	}
	if ret.Message != "" {
		return errors.New(ret.Message)
	}
	d, err := json.Marshal(ret.Data)
	if err != nil {
		return err
	}
	details := struct {
		Output       string `json:"output"`
		Content      []byte `json:"content"`
		Dependencies []byte `json:"dependencies"`
	}{}
	err = json.Unmarshal(d, &details)
	if err != nil {
		return err
	}
	if strings.Contains(details.Output, "created") {
		// 写入数据库
		req := &request.OFDBRequest{
			UUID:         uuid.New().String(),
			Driver:       of.GetConfig().Driver,
			Name:         name,
			Lang:         lang,
			Content:      details.Content,
			Dependencies: details.Dependencies,
			Status:       "CREATED",
			Timestamp:    time.Now().Unix(),
		}
		err = of.backend.SetFaasInfo(req)
		if err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("uncommon output: %s", details.Output)
}

func (of *OpenFaas) Write(id, name, lang string, content, dependencies []byte) error {
	url := fmt.Sprintf("%s/api/faasd/write", of.GetConfig().Endpoint)
	data := struct {
		Name         string `json:"name"`
		Lang         string `json:"lang"`
		Content      []byte `json:"content"`
		Dependencies []byte `json:"dependencies"`
	}{
		Name:         name,
		Lang:         lang,
		Content:      content,
		Dependencies: dependencies,
	}
	ret, err := of.send(url, "POST", &data)
	if err != nil {
		return err
	}
	if ret.Message != "" {
		return errors.New(ret.Message)
	}
	// 更新数据库
	req := &request.OFDBRequest{
		UUID:         id,
		Content:      content,
		Dependencies: dependencies,
		Status:       "WRITTEN",
		Timestamp:    time.Now().Unix(),
	}
	err = of.backend.UpdateFaasInfo(req)
	if err != nil {
		return err
	}
	return nil
}

func (of *OpenFaas) Up(id, name string, opts ...*iface.UpOptions) error {
	url := fmt.Sprintf("%s/api/faasd/up", of.GetConfig().Endpoint)
	data := struct {
		Name string `json:"name"`
	}{
		Name: name,
	}
	ret, err := of.send(url, "POST", &data)
	if err != nil {
		return err
	}
	if ret.Message != "" {
		return errors.New(ret.Message)
	}

	d, err := json.Marshal(ret.Data)
	if err != nil {
		return err
	}
	details := struct {
		Output string `json:"output"`
	}{}
	err = json.Unmarshal(d, &details)
	if err != nil {
		return err
	}
	if strings.Contains(details.Output, "Deployed") {
		// Up后获取URL
		params, err := of.describe(name)
		if err != nil {
			return err
		}
		// 更新数据库
		req := &request.OFDBRequest{
			UUID:      id,
			Status:    "UP",
			URL:       params["URL"],
			Timestamp: time.Now().Unix(),
		}
		err = of.backend.UpdateFaasInfo(req)
		if err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("uncommon output: %s", details.Output)
}

func (of *OpenFaas) Delete(id, name string, opts ...*iface.DelOptions) error {
	url := fmt.Sprintf("%s/api/faasd/delete", of.GetConfig().Endpoint)
	data := struct {
		Name string `json:"name"`
	}{
		Name: name,
	}
	ret, err := of.send(url, "POST", &data)
	if err != nil {
		return err
	}
	if ret.Message != "" {
		return errors.New(ret.Message)
	}

	d, err := json.Marshal(ret.Data)
	if err != nil {
		return err
	}
	details := struct {
		Output string `json:"output"`
	}{}
	err = json.Unmarshal(d, &details)
	if err != nil {
		return err
	}
	// 更新数据库
	req := &request.OFDBRequest{
		UUID:      id,
		Status:    "DELETED",
		Timestamp: time.Now().Unix(),
	}
	err = of.backend.UpdateFaasInfo(req)
	if err != nil {
		return err
	}
	return nil
	return fmt.Errorf("uncommon output: %s", details.Output)
}

func (of *OpenFaas) describe(name string) (map[string]string, error) {
	url := fmt.Sprintf("%s/api/faasd/describe", of.GetConfig().Endpoint)
	data := struct {
		Name string `json:"name"`
	}{
		Name: name,
	}
	ret, err := of.send(url, "POST", &data)
	if err != nil {
		return nil, err
	}
	if ret.Message != "" {
		return nil, errors.New(ret.Message)
	}
	d, err := json.Marshal(ret.Data)
	if err != nil {
		return nil, err
	}
	details := struct {
		Output string `json:"output"`
	}{}
	err = json.Unmarshal(d, &details)
	if err != nil {
		return nil, err
	}
	params := utils.ExtractParams(details.Output)
	return params, nil
}

func (of *OpenFaas) List() ([]*request.OFDBResponse, error) {
	var newRets []*request.OFDBResponse
	// 查数据库
	rets, err := of.backend.GetAllFaasInfo()
	if err != nil {
		return nil, err
	}
	// 数据核查
	for idx, ret := range rets {
		if ret.Status == "DELETED" {
			continue
		}
		// 非UP需要重新部署，生产新的URL
		if ret.Status != "UP" {
			ret.URL = ""
		}
		// 修复URL缺失
		if ret.URL == "" && ret.Status == "UP" {
			// Up后获取URL
			params, err := of.describe(ret.Name)
			if err != nil {
				log.Runtime().Infof("Can`t obtain describe for %s, err: %v", ret.Name, err)
				continue
			}
			rets[idx].URL = params["URL"]
			// 更新数据库
			req := &request.OFDBRequest{
				UUID:      ret.UUID,
				URL:       params["URL"],
				Timestamp: time.Now().Unix(),
			}
			err = of.backend.UpdateFaasInfo(req)
			if err != nil {
				continue
			}
		}
		newRets = append(newRets, ret)
	}
	return newRets, nil
}

func (of *OpenFaas) SupportedLang() ([]string, error) {
	url := fmt.Sprintf("%s/api/faasd/support", of.GetConfig().Endpoint)
	ret, err := of.send(url, "GET", nil)
	if err != nil {
		return nil, err
	}
	if ret.Message != "" {
		return nil, errors.New(ret.Message)
	}
	langs := []string{}
	d, err := json.Marshal(ret.Data)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(d, &langs)
	if err != nil {
		return nil, err
	}
	return langs, nil
}

func (of *OpenFaas) login() error {
	url := fmt.Sprintf("%s/api/login", of.GetConfig().Endpoint)
	ret, err := of.send(url, "GET", nil)
	if err != nil {
		return err
	}
	if ret.Message != "" {
		return errors.New(ret.Message)
	}
	return nil
}

func (of *OpenFaas) send(url, method string, data interface{}) (*Result, error) {
	// 返回结果
	d, err := of.sendNotParse(url, method, data)
	if err != nil {
		return nil, err
	}
	ret := &Result{}
	// 初步同一结构解析
	err = json.Unmarshal(d, &ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (of *OpenFaas) sendNotParse(url, method string, data interface{}) ([]byte, error) {
	var body io.Reader
	if data != nil {
		d, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(d)
	}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	// 添加Cookie
	for _, cookie := range of.cookies {
		req.AddCookie(cookie)
	}
	// 授权请求
	if strings.Contains(url, "login") {
		q := req.URL.Query()
		q.Add("name", of.GetConfig().Name)
		q.Add("password", of.GetConfig().Password)
		req.URL.RawQuery = q.Encode()
	}
	// 发送请求
	rep, err := of.client.Do(req)
	if err != nil {
		return nil, err
	}
	if rep == nil {
		return nil, nil
	}
	// 关闭body
	defer rep.Body.Close()
	// 处理Set-Cookie
	if len(rep.Cookies()) > 0 {
		of.cookies = rep.Cookies()
	}
	// 处理未授权
	if rep.StatusCode == http.StatusUnauthorized && !strings.Contains(url, "login") {
		err = of.login()
		if err != nil {
			return nil, fmt.Errorf("Login failed: %s", err.Error())
		}
		// 消息重发
		return of.sendNotParse(url, method, data)
	} else if rep.StatusCode == http.StatusUnauthorized && strings.Contains(url, "login") {
		return nil, errors.New("密码或用户名错误")
	} else if rep.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Received a uncommon code: %d", req.Response.StatusCode)
	}

	d, err := ioutil.ReadAll(rep.Body)
	if err != nil {
		return nil, err
	}
	return d, nil
}
