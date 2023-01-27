package iface

import "aurora/internal/config"

type Repo interface {
	GetConfig() *config.Config
	UploadFile(fileName string, fileContent []byte) error
	UpdateFile(fileName string, fileContent []byte) error
	DeleteFile(fileName string) error
	DownloadFile(fileName string) (fileContent []byte, err error)
	TestConnect() error
}
