package iface

import (
	"aurora/internal/config"
)

type Repo interface {
	GetConfig() *config.Config
	UploadFile(fileName string, fileContent []byte) (string, error)
	UpdateFile(fileID string, fileName string, fileContent []byte) error
	DeleteFile(fileID string) error
	DownloadFile(fileID string) (fileContent []byte, err error)
	TestConnect() error
}
