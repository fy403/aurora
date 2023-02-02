package aliyunfc_test

import (
	"aurora/internal/config"
	"aurora/internal/faas/aliyunfc"
	"aurora/internal/faas/iface"
	"encoding/base64"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newAliyunFC() (iface.Faas, error) {
	if os.Getenv("ENDPOINT") == "" {
		return nil, errors.New("ENDPOINT is not defined")
	}
	if os.Getenv("ACCESS_KEY_ID") == "" {
		return nil, errors.New("ACCESS_KEY_ID is not defined")
	}
	if os.Getenv("ACCESS_KEY_SECRET") == "" {
		return nil, errors.New("ACCESS_KEY_SECRET is not defined")
	}
	faasCnf := &config.Faas{
		Driver:          "aliyunfc",
		Endpoint:        os.Getenv("ENDPOINT"),
		ServiceName:     "test",
		AccessKeyId:     os.Getenv("ACCESS_KEY_ID"),
		AccessKeySecret: os.Getenv("ACCESS_KEY_SECRET"),
	}
	fs, err := aliyunfc.New(faasCnf)
	if err != nil {
		return nil, err
	}
	return fs, nil
}

func TestList(t *testing.T) {
	fc, err := newAliyunFC()
	if assert.NoError(t, err) {
		assert.NotNil(t, fc)
	}
	resp, err := fc.List()
	if assert.NoError(t, err) {
		assert.NotNil(t, resp)
	}
	t.Log(len(resp))
}

func TestInvoke(t *testing.T) {
	fc, err := newAliyunFC()
	if assert.NoError(t, err) {
		assert.NotNil(t, fc)
	}
	resp, err := fc.Invoke("py3", "")
	if assert.NoError(t, err) {
		assert.NotNil(t, resp)
	}
	data, _ := base64.StdEncoding.DecodeString(resp)
	t.Logf("%#v", string(data))
}
