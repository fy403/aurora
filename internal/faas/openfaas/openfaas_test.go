package openfaas_test

import (
	"aurora/internal/backends/mongo"
	"aurora/internal/config"
	"aurora/internal/faas/iface"
	"aurora/internal/faas/openfaas"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	id = "020d6cec-ad22-4079-99c7-9ba0e89d3d1b"
)

func newOpenFaas() (iface.Faas, error) {
	cnf := &config.Config{
		ResultBackend:   os.Getenv("MONGODB_URL"),
		ResultsExpireIn: 30,
	}
	backend, err := mongo.New(cnf)
	if err != nil {
		return nil, err
	}
	faasCnf := &config.Faas{
		Driver:   "openfaas",
		Endpoint: "http://localhost:18080",
		Name:     "admin",
		Password: "123456",
		Prefix:   "fy403",
	}
	fs, err := openfaas.New(faasCnf, backend)
	if err != nil {
		return nil, err
	}
	return fs, nil
}

func TestNew(t *testing.T) {
	if os.Getenv("MONGODB_URL") == "" {
		t.Skip("MONGODB_URL is not defined")
	}

	fs, err := newOpenFaas()
	if assert.NoError(t, err) {
		assert.NotNil(t, fs)
	}
	err = fs.New("go-demo-test", "golang-http", fs.GetConfig().Prefix)
	assert.NoError(t, err)
}

func TestUp(t *testing.T) {
	if os.Getenv("MONGODB_URL") == "" {
		t.Skip("MONGODB_URL is not defined")
	}

	fs, err := newOpenFaas()
	if assert.NoError(t, err) {
		assert.NotNil(t, fs)
	}
	err = fs.Up(id, "go-demo-test")
	assert.NoError(t, err)
}

func TestDelete(t *testing.T) {
	if os.Getenv("MONGODB_URL") == "" {
		t.Skip("MONGODB_URL is not defined")
	}

	fs, err := newOpenFaas()
	if assert.NoError(t, err) {
		assert.NotNil(t, fs)
	}
	// 去数据库找_id
	err = fs.Delete(id, "go-demo-test")
	assert.NoError(t, err)
}

func TestList(t *testing.T) {
	if os.Getenv("MONGODB_URL") == "" {
		t.Skip("MONGODB_URL is not defined")
	}

	fs, err := newOpenFaas()
	if assert.NoError(t, err) {
		assert.NotNil(t, fs)
	}
	instances, err := fs.List()
	if assert.NoError(t, err) {
		assert.NotNil(t, instances)
	}
	t.Logf("%#v", instances)
}

func TestSupportedLang(t *testing.T) {
	if os.Getenv("MONGODB_URL") == "" {
		t.Skip("MONGODB_URL is not defined")
	}

	fs, err := newOpenFaas()
	if assert.NoError(t, err) {
		assert.NotNil(t, fs)
	}
	langs, err := fs.SupportedLang()
	if assert.NoError(t, err) {
		assert.NotNil(t, langs)
	}
	t.Logf("%v", langs)
}
