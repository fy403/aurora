package worker

import (
	cachesiface "aurora/internal/cache/iface"
	"aurora/internal/constant"
	"aurora/internal/request"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/goccy/go-json"
)

func GetAllWorkersInfo(cache cachesiface.Cache) ([]*request.WorkerResponse, error) {
	keys, err := cache.Keys(constant.WorkerKeys)
	if err != nil {
		return nil, err
	}
	var workerMetaStrs = make([]interface{}, 0, len(keys))
	for _, key := range keys {
		workerMetaStr, err := cache.Get(key.(string))
		if err != nil {
			continue
		}
		workerMetaStrs = append(workerMetaStrs, workerMetaStr)
	}
	var resps = make([]*request.WorkerResponse, 0, len(workerMetaStrs))
	for _, workerMetaStr := range workerMetaStrs {
		var workerMeta request.WorkerMeta
		strData, ok := workerMetaStr.(string)
		if !ok {
			return nil, fmt.Errorf("cache.Get data is not string: %#v", workerMetaStr)
		}
		err = json.Unmarshal([]byte(strData), &workerMeta)
		if err != nil {
			return nil, err
		}
		resps = append(resps, &request.WorkerResponse{
			UUID:      workerMeta.UUID,
			SpecQueue: workerMeta.SpecQueue,
			Metrics:   workerMeta.Metrics,
			Handlers:  workerMeta.Handlers,
			Labels:    workerMeta.Labels,
			Timestamp: workerMeta.CreatedAt,
		})
	}
	return resps, nil
}

func UpdateWorkerInfo(cache cachesiface.Cache, req *request.WorkerRequest) error {
	data, err := cache.Get(fmt.Sprintf(constant.WorkerMetaFormat, req.UUID))
	// 可能已经过期，重新Add
	if err != nil {
		if err.Error() == redis.Nil.Error() {
			return SetWorkerInfo(cache, req)
		} else {
			return err
		}
	}
	var workerMeta request.WorkerMeta
	strData, ok := data.(string)
	if !ok {
		return fmt.Errorf("cache.Get data is not string: %#v", data)
	}
	err = json.Unmarshal([]byte(strData), &workerMeta)
	if err != nil {
		return err
	}
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
	workerMeta.UUID = req.UUID
	workerMeta.CreatedAt = req.Timestamp
	return cache.Add(fmt.Sprintf(constant.WorkerMetaFormat, req.UUID), data)
}

func PurgeWorkerInfo(cache cachesiface.Cache, req *request.WorkerRequest) error {
	return cache.Del(fmt.Sprintf(constant.WorkerMetaFormat, req.UUID))
}

func SetWorkerInfo(cache cachesiface.Cache, req *request.WorkerRequest) error {
	workerMeta := &request.WorkerMeta{
		UUID:      req.UUID,
		SpecQueue: req.SpecQueue,
		Metrics:   req.Metrics,
		Handlers:  req.Handlers,
		Labels:    req.Labels,
		CreatedAt: req.Timestamp,
	}
	data, err := json.Marshal(workerMeta)
	if err != nil {
		return err
	}
	return cache.Add(fmt.Sprintf(constant.WorkerMetaFormat, workerMeta.UUID), data)
}
