package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

type AuroraConnector struct {
	client      *http.Client
	authRequest *AuthRequest
	cookies     []*http.Cookie
	loginUrl    string
	tasksUrl    string
	connError   chan error
}

// NewAuroraConnector create a client instance
func NewAuroraConnector(loginUrl, tasksUrl string) *AuroraConnector {
	defaultTransport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   15 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConnsPerHost: 2,
		MaxIdleConns:        0,
		IdleConnTimeout:     10 * time.Second,
	}
	return &AuroraConnector{
		client: &http.Client{
			Timeout:   time.Second * time.Duration(15),
			Transport: defaultTransport,
		},
		authRequest: &AuthRequest{},
		loginUrl:    loginUrl,
		tasksUrl:    tasksUrl,
	}
}

// Init set some personal information for login
func (conn *AuroraConnector) Init(userName, password string) error {
	conn.authRequest.Name = userName
	conn.authRequest.Password = password
	return nil
}

// Login send a logit post http request to Aurora
func (conn *AuroraConnector) login() error {
	var err error
	requestOBJ, err := json.Marshal(conn.authRequest)
	if err != nil {
		return err
	}
	bodyReader := bytes.NewReader(requestOBJ)
	request, err := http.NewRequest("POST", conn.loginUrl, bodyReader)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := conn.client.Do(request)
	if err != nil {
		return err
	}
	if response == nil {
		return errors.New("Response is nil")
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(response.Body)
		return fmt.Errorf("Login failed error: %s", string(body))
	}
	conn.cookies = response.Cookies()
	return nil
}

// SendSync send a sync http request to Aurora
func (conn *AuroraConnector) SendSync(requestOBJ *CenterRequest) (*CenterResponse, error) {
	var err error
	req, err := json.Marshal(requestOBJ)
	if err != nil {
		return nil, err
	}
	bodyReader := bytes.NewReader(req)
	request, err := http.NewRequest("POST", conn.tasksUrl, bodyReader)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Cache-control", " no-cache")
	for _, c := range conn.cookies {
		request.AddCookie(c)
	}
	response, err := conn.client.Do(request)
	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, errors.New("Response is nil")
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	switch response.StatusCode {
	case http.StatusOK:
		responseOBJ := &CenterResponse{}
		err = json.Unmarshal(body, responseOBJ)
		if err != nil {
			return nil, err
		}
		return responseOBJ, nil
	case http.StatusForbidden:
		err := conn.login()
		if err != nil {
			return nil, err
		}
		return conn.SendSync(requestOBJ)
	default:
		return nil, fmt.Errorf("Send fail: %s", string(body))
	}
}

func (conn *AuroraConnector) Close() error {
	conn.client.CloseIdleConnections()
	return nil
}
