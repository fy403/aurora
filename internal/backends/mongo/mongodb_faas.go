package mongo

import (
	"aurora/internal/repo/mongo"
	"aurora/internal/request"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (b *Backend) SetFaasInfo(req *request.OFDBRequest) error {
	// 写入内容文件
	r, _ := mongo.New(b.GetConfig())
	contentFileID, err := r.UploadFile(req.Name, req.Content)
	if err != nil {
		return err
	}
	// 写入依赖文件
	dependenciesFileID, err := r.UploadFile(req.Name, req.Dependencies)
	if err != nil {
		return err
	}

	oFDBMeta := &request.OFDBMeta{
		UUID:                 req.UUID,
		Driver:               req.Driver,
		Name:                 req.Name,
		Lang:                 req.Lang,
		Content_File_ID:      contentFileID,
		Dependencies_File_ID: dependenciesFileID,
		Status:               req.Status,
		CreatedAt:            req.Timestamp,
	}
	_, err = b.faasCollection().InsertOne(context.Background(), oFDBMeta)
	return err
}

func (b *Backend) GetAllFaasInfo() ([]*request.OFDBResponse, error) {
	count, err := b.faasCollection().CountDocuments(context.Background(), bson.D{})
	if err != nil {
		return nil, err
	}
	results := make([]*request.OFDBResponse, 0, count)
	cursor, err := b.faasCollection().Find(context.Background(), bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())
	// Iterate the cursor and print out each document until the cursor is
	// exhausted or there is an error getting the next document.
	for cursor.Next(context.TODO()) {
		// A new result variable should be declared for each document.
		var result request.OFDBMeta
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}
		r, _ := mongo.New(b.GetConfig())

		content, err := r.DownloadFile(result.Content_File_ID)
		if err != nil {
			return nil, err
		}
		dependencies, err := r.DownloadFile(result.Dependencies_File_ID)
		if err != nil {
			return nil, err
		}
		results = append(results, &request.OFDBResponse{
			UUID:         result.UUID,
			Driver:       result.Driver,
			Name:         result.Name,
			URL:          result.URL,
			Lang:         result.Lang,
			Content:      content,
			Dependencies: dependencies,
			Status:       result.Status,
			Timestamp:    result.CreatedAt,
		})
	}
	if err = cursor.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

func (b *Backend) UpdateFaasInfo(req *request.OFDBRequest) error {
	var contentFileID, dependenciesFileID primitive.ObjectID
	// 先查再更新
	oldMeta, err := b.findOneInFC(req.UUID)
	if err != nil {
		return err
	}
	oFDBMeta := oldMeta
	// 写入内容文件
	r, _ := mongo.New(b.GetConfig())
	if len(req.Content) > 0 {
		contentFileID, err = r.UploadFile(req.Name, req.Content)
		if err != nil {
			return err
		}
		oFDBMeta.Content_File_ID = contentFileID
	}
	// 写入依赖文件
	if len(req.Dependencies) > 0 {
		dependenciesFileID, err = r.UploadFile(req.Name, req.Dependencies)
		if err != nil {
			return err
		}
		oFDBMeta.Dependencies_File_ID = dependenciesFileID
	}
	// 局部更新
	if req.Driver != "" {
		oFDBMeta.Driver = req.Driver
	}
	if req.Lang != "" {
		oFDBMeta.Lang = req.Lang
	}
	if req.URL != "" {
		oFDBMeta.URL = req.URL
	}
	if req.Status != "" {
		oFDBMeta.Status = req.Status
	}
	oFDBMeta.CreatedAt = req.Timestamp
	update := bson.M{"$set": oFDBMeta}
	_, err = b.faasCollection().UpdateOne(context.Background(), bson.M{"_id": oFDBMeta.UUID}, update)
	return err
}

func (b *Backend) PurgeFaasInfo(req *request.OFDBRequest) error {
	_, err := b.faasCollection().DeleteOne(context.Background(), bson.M{"_id": req.UUID})
	return err
}

func (b *Backend) findOneInFC(id string) (oldMeta *request.OFDBMeta, err error) {
	filter := bson.D{
		{"_id", id},
	}
	result := b.faasCollection().FindOne(context.Background(), filter)
	oldMeta = &request.OFDBMeta{}
	err = result.Decode(oldMeta)
	if err != nil {
		return nil, err
	}
	return oldMeta, nil
}
