package openfaas_test

import (
	"aurora/internal/config"
	"aurora/internal/faas/iface"
	"aurora/internal/faas/openfaas"
)

func newOpenFaas() (iface.Faas, error) {
	faasCnf := &config.Faas{
		Driver:   "openfaas",
		Endpoint: "http://localhost:18080",
		Name:     "admin",
		Password: "123456",
	}
	fs, err := openfaas.New(faasCnf)
	if err != nil {
		return nil, err
	}
	return fs, nil
}

// func TestNew(t *testing.T) {
// 	if os.Getenv("MONGODB_URL") == "" {
// 		t.Skip("MONGODB_URL is not defined")
// 	}

// 	fs, err := newOpenFaas()
// 	if assert.NoError(t, err) {
// 		assert.NotNil(t, fs)
// 	}
// 	err = fs.New(name, "golang-http", fs.GetConfig().Prefix)
// 	assert.NoError(t, err)
// }
