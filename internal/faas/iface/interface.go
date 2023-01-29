package iface

import (
	"aurora/internal/config"
	"aurora/internal/request"
)

type Faas interface {
	GetConfig() *config.Faas
	New(name, lang, prefix string, opts ...*NewOptions) error
	Write(id, name, lang string, content, dependencies []byte) error
	Up(id, name string, opts ...*UpOptions) error
	List() ([]*request.OFDBResponse, error)
	SupportedLang() ([]string, error)
	Delete(id, name string, opts ...*DelOptions) error
}
