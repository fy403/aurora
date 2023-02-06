package mongo

import (
	"aurora/internal/common"
	"aurora/internal/config"
	"aurora/internal/log"
	"aurora/internal/repo/iface"
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Repo struct {
	common.Repo
	client *mongo.Client
	once   sync.Once
}

func New(cnf *config.Config) (iface.Repo, error) {
	return &Repo{
		Repo: common.NewRepo(cnf),
		once: sync.Once{},
	}, nil
}

func (r *Repo) UploadFile(fileName string, fileContent []byte) (primitive.ObjectID, error) {
	bucket := r.getGridFSBucket("")
	fileID, err := bucket.UploadFromStream(fileName, bytes.NewBuffer(fileContent))
	if err != nil {
		log.Runtime().Error(err.Error())
		return primitive.NilObjectID, err
	}
	return fileID, nil
}

func (r *Repo) UpdateFile(fileID primitive.ObjectID, fileName string, fileContent []byte) error {
	bucket := r.getGridFSBucket("")
	err := r.DeleteFile(fileID)
	if err != nil {
		log.Runtime().Error(err.Error())
		return err
	}
	err = bucket.UploadFromStreamWithID(fileID, fileName, bytes.NewBuffer(fileContent))
	if err != nil {
		log.Runtime().Error(err.Error())
		return err
	}
	return nil
}

func (r *Repo) DeleteFile(fileID primitive.ObjectID) error {
	bucket := r.getGridFSBucket("")
	if err := bucket.Delete(fileID); err != nil && err != gridfs.ErrFileNotFound {
		log.Runtime().Error(err.Error())
		return err
	}
	return nil
}

func (r *Repo) DownloadFile(fileID primitive.ObjectID) (fileContent []byte, err error) {
	bucket := r.getGridFSBucket("")
	fileBuffer := bytes.NewBuffer(nil)
	if _, err = bucket.DownloadToStream(fileID, fileBuffer); err != nil {
		log.Runtime().Error(err.Error())
		return nil, err
	}
	return fileBuffer.Bytes(), nil
}

// collName:文件集合名称 fileID:文件ID，必须唯一，否则会覆盖
// fileName:文件名称 fileContent:文件内容
func (r *Repo) getGridFSBucket(collName string) *gridfs.Bucket {
	database := "aurora"
	if r.GetConfig().MongoDB != nil {
		database = r.GetConfig().MongoDB.Database
	}
	r.once.Do(func() {
		r.connect()
	})
	var mongoDatabase = r.client.Database(database)

	var bucket *gridfs.Bucket
	// 使用默认文件集合名称
	if collName == "" || collName == options.DefaultName {
		bucket, _ = gridfs.NewBucket(mongoDatabase)
	} else {
		// 使用传入的文件集合名称
		bucketOptions := options.GridFSBucket().SetName(collName)
		bucket, _ = gridfs.NewBucket(mongoDatabase, bucketOptions)
	}
	return bucket
}

// connect creates the underlying mgo connection if it doesn't exist
// creates required indexes for our collections
func (r *Repo) connect() error {
	client, err := r.dial()
	if err != nil {
		return err
	}
	r.client = client

	// database := "aurora"

	// if r.GetConfig().MongoDB != nil {
	// 	database = r.GetConfig().MongoDB.Database
	// }

	// err = r.createMongoIndexes(database)
	// if err != nil {
	// 	return err
	// }
	return nil
}

// dial connects to mongo with TLSConfig if provided
// else connects via ResultBackend uri
func (r *Repo) dial() (*mongo.Client, error) {

	if r.GetConfig().MongoDB != nil && r.GetConfig().MongoDB.Client != nil {
		return r.GetConfig().MongoDB.Client, nil
	}

	uri := r.GetConfig().ResultBackend
	if strings.HasPrefix(uri, "mongodb://") == false &&
		strings.HasPrefix(uri, "mongodb+srv://") == false {
		uri = fmt.Sprintf("mongodb://%s", uri)
	}

	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		return nil, err
	}

	return client, nil
}

func (r *Repo) TestConnect() error {
	client, err := r.dial()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Ping(ctx, nil)
	if err != nil {
		return err
	}
	client.Disconnect(ctx)
	return nil
}
