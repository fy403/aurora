package mongo

import (
	"aurora/internal/request"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SetWorkerInfo sets worker runtime info
func (b *Backend) SetWorkerInfo(req *request.WorkerRequest) error {
	workerMeta := &request.WorkerMeta{
		UUID:      req.UUID,
		SpecQueue: req.SpecQueue,
		Metrics:   req.Metrics,
		Handlers:  req.Handlers,
		Labels:    req.Labels,
		CreateAt:  req.Timestamp,
	}
	_, err := b.workersCollection().InsertOne(context.Background(), workerMeta)
	return err
}

// GetAllWorkersInfo gets worker runtime info
func (b *Backend) GetAllWorkersInfo() ([]*request.WorkerResponse, error) {
	count, err := b.workersCollection().CountDocuments(context.Background(), bson.D{})
	if err != nil {
		return nil, err
	}
	results := make([]*request.WorkerResponse, 0, count)
	cursor, err := b.workersCollection().Find(context.Background(), bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())
	// Iterate the cursor and print out each document until the cursor is
	// exhausted or there is an error getting the next document.
	for cursor.Next(context.TODO()) {
		// A new result variable should be declared for each document.
		var result request.WorkerMeta
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}
		results = append(results, &request.WorkerResponse{
			UUID:      result.UUID,
			SpecQueue: result.SpecQueue,
			Metrics:   result.Metrics,
			Handlers:  result.Handlers,
			Labels:    result.Labels,
			Timestamp: result.CreateAt,
		})
	}
	if err = cursor.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

// UpdateWorkerInfo updates worker runtime info
func (b *Backend) UpdateWorkerInfo(req *request.WorkerRequest) error {
	workerMeta := &request.WorkerMeta{
		UUID:      req.UUID,
		SpecQueue: req.SpecQueue,
		Metrics:   req.Metrics,
		Handlers:  req.Handlers,
		Labels:    req.Labels,
		CreateAt:  req.Timestamp,
	}
	update := bson.M{"$set": workerMeta}
	_, err := b.workersCollection().UpdateOne(context.Background(), bson.M{"_id": workerMeta.UUID}, update, options.Update().SetUpsert(true))
	return err
}

// PurgeWorkerInfo purges worker runtime info
func (b *Backend) PurgeWorkerInfo(req *request.WorkerRequest) error {
	_, err := b.workersCollection().DeleteOne(context.Background(), bson.M{"_id": req.UUID})
	return err
}
