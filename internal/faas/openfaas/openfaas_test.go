package openfaas_test

import (
	"aurora/internal/config"
	"aurora/internal/faas/iface"
	"aurora/internal/faas/openfaas"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newOpenFaas() (iface.Faas, error) {
	faasCnf := &config.Faas{
		Driver:   "openfaas",
		Endpoint: "http://localhost:8080",
		Name:     "admin",
		Password: "9DTyNVs0ABQYT4jQ9yEuDtcVzD4gQXfkh4XdssGQ5tFx2UHucXMrqthvuoU3XV2",
	}
	fs, err := openfaas.New(faasCnf)
	if err != nil {
		return nil, err
	}
	return fs, nil
}

func TestList(t *testing.T) {
	fc, err := newOpenFaas()
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
	fc, err := newOpenFaas()
	if assert.NoError(t, err) {
		assert.NotNil(t, fc)
	}
	resp, err := fc.Invoke("node-demo", "")
	if assert.NoError(t, err) {
		assert.NotNil(t, resp)
	}
	data, _ := base64.StdEncoding.DecodeString(resp)
	t.Logf("%#v", string(data))
}
