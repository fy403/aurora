package iface

import (
	"aurora/internal/config"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Repo interface {
	GetConfig() *config.Config
	UploadFile(fileName string, fileContent []byte) (primitive.ObjectID, error)
	UpdateFile(fileID primitive.ObjectID, fileName string, fileContent []byte) error
	DeleteFile(fileID primitive.ObjectID) error
	DownloadFile(fileID primitive.ObjectID) (fileContent []byte, err error)
	TestConnect() error
}
