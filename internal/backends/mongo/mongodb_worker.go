package mongo

import (
	"aurora/internal/request"
	"context"

	"go.mongodb.org/mongo-driver/bson"
)

// SetWorkerInfo sets worker runtime info
func (b *Backend) SetWorkerInfo(req *request.WorkerRequest) error {
	workerMeta := &request.WorkerMeta{
		UUID:      req.UUID,
		SpecQueue: req.SpecQueue,
		Metrics:   req.Metrics,
		Handlers:  req.Handlers,
		Labels:    req.Labels,
		CreatedAt: req.Timestamp,
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
			Timestamp: result.CreatedAt,
		})
	}
	if err = cursor.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

// UpdateWorkerInfo updates worker runtime info
func (b *Backend) UpdateWorkerInfo(req *request.WorkerRequest) error {
	oldMeta, err := b.findOneInWC(req.UUID)
	workerMeta := oldMeta
	// 局部更新
	if req.SpecQueue != "" {
		workerMeta.SpecQueue = req.SpecQueue
	}
	if len(req.Metrics) > 0 {
		workerMeta.Metrics = req.Metrics
	}
	if len(req.Handlers) > 0 {
		workerMeta.Handlers = req.Handlers
	}
	if len(req.Labels) > 0 {
		workerMeta.Labels = req.Labels
	}
	workerMeta.CreatedAt = req.Timestamp
	update := bson.M{"$set": workerMeta}
	_, err = b.workersCollection().UpdateOne(context.Background(), bson.M{"_id": req.UUID}, update)
	return err
}

// PurgeWorkerInfo purges worker runtime info
func (b *Backend) PurgeWorkerInfo(req *request.WorkerRequest) error {
	_, err := b.workersCollection().DeleteOne(context.Background(), bson.M{"_id": req.UUID})
	return err
}

func (b *Backend) findOneInWC(id string) (oldMeta *request.WorkerMeta, err error) {
	filter := bson.D{
		{"_id", id},
	}
	result := b.workersCollection().FindOne(context.Background(), filter)
	oldMeta = &request.WorkerMeta{}
	err = result.Decode(oldMeta)
	if err != nil {
		return nil, err
	}
	return oldMeta, nil
}
