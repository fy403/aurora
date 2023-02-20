package center

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/robfig/cron/v3"

	"aurora/internal/backends/result"
	"aurora/internal/config"
	"aurora/internal/constant"
	"aurora/internal/faas"
	"aurora/internal/faas/aliyunfc"
	"aurora/internal/faas/iface"
	"aurora/internal/faas/openfaas"
	"aurora/internal/log"
	"aurora/internal/opentracing/tracing"
	"aurora/internal/request"
	"aurora/internal/tasks"
	algorithm "aurora/internal/utils/algorithm"

	backendsiface "aurora/internal/backends/iface"
	cachesiface "aurora/internal/cache/iface"

	brokersiface "aurora/internal/brokers/iface"
	lockiface "aurora/internal/locks/iface"

	opentracing "github.com/opentracing/opentracing-go"
)

// Server is the main aurora object and stores all configuration
// All the tasks workers process are registered against the server
type Server struct {
	config            *config.Config
	registeredTasks   *sync.Map
	broker            brokersiface.Broker
	backend           backendsiface.Backend
	cache             cachesiface.Cache
	lock              lockiface.Lock
	scheduler         *cron.Cron
	prePublishHandler func(*tasks.Signature)
}

// NewServer creates Server instance
func NewServer(cnf *config.Config, brokerServer brokersiface.Broker, backendServer backendsiface.Backend, cacheServer cachesiface.Cache, lock lockiface.Lock, onlyCnf ...bool) *Server {
	srv := &Server{
		config:          cnf,
		registeredTasks: new(sync.Map),
		broker:          brokerServer,
		backend:         backendServer,
		cache:           cacheServer,
		lock:            lock,
		scheduler:       cron.New(),
	}

	// Run scheduler job: 除非显式声明为onlyCnf = true
	if len(onlyCnf) == 0 || (len(onlyCnf) > 0 && onlyCnf[0] == false) {
		go srv.scheduler.Run()
	}

	return srv
}

// GetBroker returns broker
func (server *Server) GetBroker() brokersiface.Broker {
	return server.broker
}

// SetBroker sets broker
func (server *Server) SetBroker(broker brokersiface.Broker) {
	server.broker = broker
}

// GetBackend returns backend
func (server *Server) GetBackend() backendsiface.Backend {
	return server.backend
}

func (server *Server) GetCache() cachesiface.Cache {
	return server.cache
}

// GetLock returns lock
func (server *Server) GetLock() lockiface.Lock {
	return server.lock
}

// SetBackend sets backend
func (server *Server) SetBackend(backend backendsiface.Backend) {
	server.backend = backend
}

// GetConfig returns connection object
func (server *Server) GetConfig() *config.Config {
	return server.config
}

// SetConfig sets config
func (server *Server) SetConfig(cnf *config.Config) {
	server.config = cnf
}

// SetPreTaskHandler Sets pre publish handler
func (server *Server) SetPreTaskHandler(handler func(*tasks.Signature)) {
	server.prePublishHandler = handler
}

// RegisterTasks registers all tasks at once
func (server *Server) RegisterTasks(namedTaskFuncs map[string]*request.Handler) error {
	for name, handler := range namedTaskFuncs {
		if err := tasks.ValidateTask(handler.Fn); err != nil {
			return err
		}
		server.registeredTasks.Store(name, handler.Fn)

		typ := reflect.TypeOf(handler.Fn)
		for idx := 0; idx < typ.NumIn(); idx++ {
			arg := tasks.Arg{
				Type: typ.In(idx).String(),
			}
			handler.InArgs = append(handler.InArgs, arg)
		}
		for idx := 0; idx < typ.NumOut(); idx++ {
			arg := tasks.Arg{
				Type: typ.Out(idx).String(),
			}
			handler.OutArgs = append(handler.OutArgs, arg)
		}
		handler.Name = name
	}
	server.broker.SetRegisteredTaskNames(server.GetRegisteredTaskNames())
	return nil
}

func (server *Server) RegisterFaas(faasCfg []*config.Faas) error {
	// Initialize Faas instance
	var fs iface.Faas
	var err error
	for _, fsCfg := range faasCfg {
		switch fsCfg.Driver {
		case constant.ALIYUNFC:
			fs, err = aliyunfc.New(fsCfg)
		case constant.OPENFASS:
			fs, err = openfaas.New(fsCfg)
		default:
			log.Runtime().Fatalf("Unknown faas instance: %s", fsCfg.Driver)
			return fmt.Errorf("Unknown faas instance: %s", fsCfg.Driver)
		}
		if err != nil {
			log.Runtime().Fatalf("Can`t create the faas instance: %s, err: %v", fsCfg.Driver, err)
			return err
		}
		faas.ExtantFaasMap[fsCfg.Driver] = fs
	}
	return nil
}

// RegisterTask registers a single task
func (server *Server) RegisterTask(name string, taskFunc interface{}) error {
	if err := tasks.ValidateTask(taskFunc); err != nil {
		return err
	}
	server.registeredTasks.Store(name, taskFunc)
	server.broker.SetRegisteredTaskNames(server.GetRegisteredTaskNames())
	return nil
}

// IsTaskRegistered returns true if the task name is registered with this broker
func (server *Server) IsTaskRegistered(name string) bool {
	_, ok := server.registeredTasks.Load(name)
	return ok
}

// GetRegisteredTask returns registered task by name
func (server *Server) GetRegisteredTask(name string) (interface{}, error) {
	taskFunc, ok := server.registeredTasks.Load(name)
	if !ok {
		return nil, fmt.Errorf("Task not registered error: %s", name)
	}
	return taskFunc, nil
}

// SendTaskWithContext will inject the trace context in the signature headers before publishing it
func (server *Server) SendTaskWithContext(ctx context.Context, signature *tasks.Signature) (*result.AsyncResult, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "SendTask", tracing.ProducerOption(), tracing.AuroraTag)
	defer span.Finish()

	// tag the span with some info about the signature
	signature.Headers = tracing.HeadersWithSpan(signature.Headers, span)

	// Make sure result backend is defined
	if server.backend == nil {
		return nil, errors.New("Result backend required")
	}

	// Auto generate a UUID if not set already
	if signature.UUID == "" {
		taskID := uuid.New().String()
		signature.UUID = fmt.Sprintf("task_%v", taskID)
	}

	// Set initial task state to PENDING
	if err := server.backend.SetStatePending(signature); err != nil {
		return nil, fmt.Errorf("Set state pending error: %s", err)
	}

	if server.prePublishHandler != nil {
		server.prePublishHandler(signature)
	}

	if err := server.broker.Publish(ctx, signature); err != nil {
		return nil, fmt.Errorf("Publish message error: %s", err)
	}

	return result.NewAsyncResult(signature, server.backend), nil
}

// SendTask publishes a task to the default queue
func (server *Server) SendTask(signature *tasks.Signature) (*result.AsyncResult, error) {
	return server.SendTaskWithContext(context.Background(), signature)
}

// SendChainTask will extract tracer context info from header and update now span into header
func (server *Server) SendChainTask(signature *tasks.Signature) (*result.AsyncResult, error) {
	return server.SendChainTaskWithContext(context.Background(), signature)
}

// SendChainTaskWithContext will inject the trace context in the signature headers before publishing it
func (server *Server) SendChainTaskWithContext(ctx context.Context, signature *tasks.Signature) (*result.AsyncResult, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "SendChainTask", tracing.ProducerOption(), tracing.AuroraTag)
	defer span.Finish()

	// tag the span with some info about the signature
	signature.Headers = tracing.HeadersWithSpan(signature.Headers, span)

	// update signature`s span info into now span
	// inject the tracing span into the tasks OnSuccess signature headers
	for _, sig := range signature.OnSuccess {
		sig.Headers = tracing.HeadersWithSpan(sig.Headers, span)
	}

	// Make sure result backend is defined
	if server.backend == nil {
		return nil, errors.New("Result backend required")
	}

	// Auto generate a UUID if not set already
	if signature.UUID == "" {
		taskID := uuid.New().String()
		signature.UUID = fmt.Sprintf("task_%v", taskID)
	}

	// Set initial task state to PENDING
	if err := server.backend.SetStatePending(signature); err != nil {
		return nil, fmt.Errorf("Set state pending error: %s", err)
	}

	if server.prePublishHandler != nil {
		server.prePublishHandler(signature)
	}

	if err := server.broker.Publish(context.Background(), signature); err != nil {
		return nil, fmt.Errorf("Publish message error: %s", err)
	}

	return result.NewAsyncResult(signature, server.backend), nil
}

// SendChainWithContext will inject the trace context in all the signature headers before publishing it
func (server *Server) SendChainWithContext(ctx context.Context, chain *tasks.Chain) (*result.ChainAsyncResult, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "SendChain", tracing.ProducerOption(), tracing.AuroraTag, tracing.WorkflowChainTag)
	defer span.Finish()

	tracing.AnnotateSpanWithChainInfo(span, chain)
	_, err := server.SendChainTask(chain.Tasks[0])
	if err != nil {
		return nil, err
	}
	return result.NewChainAsyncResult(chain.Tasks, server.backend), nil
}

// SendChain triggers a chain of tasks
func (server *Server) SendChain(chain *tasks.Chain) (*result.ChainAsyncResult, error) {
	return server.SendChainWithContext(context.Background(), chain)
}

// SendGroupWithContext will inject the trace context in all the signature headers before publishing it
func (server *Server) SendGroupWithContext(ctx context.Context, group *tasks.Group, sendConcurrency int) ([]*result.AsyncResult, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "SendGroup", tracing.ProducerOption(), tracing.AuroraTag, tracing.WorkflowGroupTag)
	defer span.Finish()

	tracing.AnnotateSpanWithGroupInfo(span, group, sendConcurrency)

	// Make sure result backend is defined
	if server.backend == nil {
		return nil, errors.New("Result backend required")
	}

	asyncResults := make([]*result.AsyncResult, len(group.Tasks))

	var wg sync.WaitGroup
	wg.Add(len(group.Tasks))
	errorsChan := make(chan error, len(group.Tasks)*2)

	// Init group
	server.backend.InitGroup(group.GroupUUID, group.GetUUIDs())

	// Init the tasks Pending state first
	for _, signature := range group.Tasks {
		if err := server.backend.SetStatePending(signature); err != nil {
			errorsChan <- err
			continue
		}
	}

	pool := make(chan struct{}, sendConcurrency)
	go func() {
		for i := 0; i < sendConcurrency; i++ {
			pool <- struct{}{}
		}
	}()

	for i, signature := range group.Tasks {

		if sendConcurrency > 0 {
			<-pool
		}

		go func(s *tasks.Signature, index int) {
			defer wg.Done()

			// Publish task

			err := server.broker.Publish(ctx, s)

			if sendConcurrency > 0 {
				pool <- struct{}{}
			}

			if err != nil {
				errorsChan <- fmt.Errorf("Publish message error: %s", err)
				return
			}

			asyncResults[index] = result.NewAsyncResult(s, server.backend)
		}(signature, i)
	}

	done := make(chan int)
	go func() {
		wg.Wait()
		done <- 1
	}()

	select {
	case err := <-errorsChan:
		return asyncResults, err
	case <-done:
		return asyncResults, nil
	}
}

// SendGroup triggers a group of parallel tasks
func (server *Server) SendGroup(group *tasks.Group, sendConcurrency int) ([]*result.AsyncResult, error) {
	return server.SendGroupWithContext(context.Background(), group, sendConcurrency)
}

// SendChordWithContext will inject the trace context in all the signature headers before publishing it
func (server *Server) SendChordWithContext(ctx context.Context, chord *tasks.Chord, sendConcurrency int) (*result.ChordAsyncResult, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "SendChord", tracing.ProducerOption(), tracing.AuroraTag, tracing.WorkflowChordTag)
	defer span.Finish()

	tracing.AnnotateSpanWithChordInfo(span, chord, sendConcurrency)

	_, err := server.SendGroupWithContext(opentracing.ContextWithSpan(ctx, span), chord.Group, sendConcurrency)
	if err != nil {
		return nil, err
	}

	return result.NewChordAsyncResult(
		chord.Group.Tasks,
		chord.Callback,
		server.backend,
	), nil
}

// SendChord triggers a group of parallel tasks with a callback
func (server *Server) SendChord(chord *tasks.Chord, sendConcurrency int) (*result.ChordAsyncResult, error) {
	return server.SendChordWithContext(context.Background(), chord, sendConcurrency)
}

// SendChordCallback will extract trace context and return a new span from signatuire.ChordCallback
func (server *Server) SendChordCallback(signature *tasks.Signature) (*result.AsyncResult, error) {
	return server.SendChordCallbackWithContext(context.Background(), signature)
}

// SendChordCallbackWithContext will inject the trace context in the signature headers before publishing it
func (server *Server) SendChordCallbackWithContext(ctx context.Context, signature *tasks.Signature) (*result.AsyncResult, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "SendChordCallback", tracing.ProducerOption(), tracing.AuroraTag)
	defer span.Finish()

	// tag the span with some info about the signature
	signature.Headers = tracing.HeadersWithSpan(signature.Headers, span)
	// Make sure result backend is defined
	if server.backend == nil {
		return nil, errors.New("Result backend required")
	}

	// Auto generate a UUID if not set already
	if signature.UUID == "" {
		taskID := uuid.New().String()
		signature.UUID = fmt.Sprintf("task_%v", taskID)
	}

	// Set initial task state to PENDING
	if err := server.backend.SetStatePending(signature); err != nil {
		return nil, fmt.Errorf("Set state pending error: %s", err)
	}

	if server.prePublishHandler != nil {
		server.prePublishHandler(signature)
	}

	if err := server.broker.Publish(context.Background(), signature); err != nil {
		return nil, fmt.Errorf("Publish message error: %s", err)
	}

	return result.NewAsyncResult(signature, server.backend), nil
}

// GetRegisteredTaskNames returns slice of registered task names
func (server *Server) GetRegisteredTaskNames() []string {
	taskNames := make([]string, 0)

	server.registeredTasks.Range(func(key, value interface{}) bool {
		taskNames = append(taskNames, key.(string))
		return true
	})
	return taskNames
}

// SendGraph triggers a graph of tasks
func (server *Server) SendGraph(graph *tasks.Graph) ([]*result.AsyncResult, error) {
	return server.SendGraphWithContext(context.Background(), graph)
}

func (server *Server) SendGraphWithContext(ctx context.Context, graph *tasks.Graph) ([]*result.AsyncResult, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "SendGraph", tracing.ProducerOption(), tracing.AuroraTag, tracing.WorkflowGroupTag)
	defer span.Finish()

	tracing.AnnotateSpanWithGraphInfo(span, graph)

	// Make sure result backend is defined
	if server.backend == nil {
		return nil, errors.New("Result backend required")
	}

	asyncResults := make([]*result.AsyncResult, len(graph.Vertexes))

	errorsChan := make(chan error, len(graph.Vertexes)*2)

	// Init group
	server.backend.InitGraph(graph)

	// Init the tasks Pending state first
	for i, signature := range graph.Vertexes {
		if err := server.backend.SetStatePending(signature); err != nil {
			errorsChan <- err
			continue
		}
		asyncResults[i] = result.NewAsyncResult(signature, server.backend)
	}

	// TopologySort
	initialTasks, ok := algorithm.TopologySort(graph)
	if !ok {
		return nil, errors.New("TopologySort failed")
	}

	var wg sync.WaitGroup
	wg.Add(len(initialTasks))

	for _, signature := range initialTasks {
		go func(s *tasks.Signature) {
			defer wg.Done()

			// Publish task
			err := server.broker.Publish(ctx, s)

			if err != nil {
				errorsChan <- fmt.Errorf("Publish message error: %s", err)
				return
			}
		}(signature)
	}

	done := make(chan int)
	go func() {
		wg.Wait()
		done <- 1
	}()

	select {
	case err := <-errorsChan:
		return asyncResults, err
	case <-done:
		return asyncResults, nil
	}
}

func (server *Server) GetAllWorkersInfo() ([]*request.WorkerResponse, error) {
	keys, err := server.GetCache().Keys(constant.WOKERKEYS)
	if err != nil {
		return nil, err
	}
	var workerMetaStrs = make([]interface{}, 0, len(keys))
	for _, key := range keys {
		workerMetaStr, err := server.GetCache().Get(key.(string))
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

func (server *Server) UpdateWorkerInfo(req *request.WorkerRequest) error {
	data, err := server.GetCache().Get(fmt.Sprintf(constant.WOKERMETAFORMAT, req.UUID))
	// 可能已经过期，重新Add
	if err != nil {
		if err.Error() == redis.Nil.Error() {
			return server.SetWorkerInfo(req)
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
	return server.GetCache().Add(fmt.Sprintf(constant.WOKERMETAFORMAT, req.UUID), data)
}

func (server *Server) PurgeWorkerInfo(req *request.WorkerRequest) error {
	return server.GetCache().Del(fmt.Sprintf(constant.WOKERMETAFORMAT, req.UUID))
}

func (server *Server) SetWorkerInfo(req *request.WorkerRequest) error {
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
	return server.GetCache().Add(fmt.Sprintf(constant.WOKERMETAFORMAT, workerMeta.UUID), data)
}
