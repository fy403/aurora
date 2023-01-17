package center

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"

	"aurora/internal/backends/result"
	"aurora/internal/config"
	"aurora/internal/log"
	"aurora/internal/opentracing/tracing"
	"aurora/internal/tasks"
	"aurora/internal/utils"
	algorithm "aurora/internal/utils/algorithm"

	backendsiface "aurora/internal/backends/iface"
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
	lock              lockiface.Lock
	scheduler         *cron.Cron
	prePublishHandler func(*tasks.Signature)
}

// NewServer creates Server instance
func NewServer(cnf *config.Config, brokerServer brokersiface.Broker, backendServer backendsiface.Backend, lock lockiface.Lock, onlyCnf ...bool) *Server {
	srv := &Server{
		config:          cnf,
		registeredTasks: new(sync.Map),
		broker:          brokerServer,
		backend:         backendServer,
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
func (server *Server) RegisterTasks(namedTaskFuncs map[string]interface{}) error {
	for _, task := range namedTaskFuncs {
		if err := tasks.ValidateTask(task); err != nil {
			return err
		}
	}
	for k, v := range namedTaskFuncs {
		server.registeredTasks.Store(k, v)
	}
	server.broker.SetRegisteredTaskNames(server.GetRegisteredTaskNames())
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

// RegisterPeriodicTask register a periodic task which will be triggered periodically
func (server *Server) RegisterPeriodicTask(spec, name string, signature *tasks.Signature) error {
	//check spec
	schedule, err := cron.ParseStandard(spec)
	if err != nil {
		return err
	}

	f := func() {
		//get lock
		err := server.lock.LockWithRetries(utils.GetLockName(name, spec), schedule.Next(time.Now()).UnixNano()-1)
		if err != nil {
			return
		}

		//send task
		_, err = server.SendTask(tasks.CopySignature(signature))
		if err != nil {
			log.Runtime().Errorf("periodic task failed. task name is: %s. error is %s", name, err.Error())
		}
	}

	_, err = server.scheduler.AddFunc(spec, f)
	return err
}

// RegisterPeriodicChain register a periodic chain which will be triggered periodically
func (server *Server) RegisterPeriodicChain(spec, name string, signatures ...*tasks.Signature) error {
	//check spec
	schedule, err := cron.ParseStandard(spec)
	if err != nil {
		return err
	}

	f := func() {
		// new chain
		chain, _ := tasks.NewChain(tasks.CopySignatures(signatures...)...)

		//get lock
		err := server.lock.LockWithRetries(utils.GetLockName(name, spec), schedule.Next(time.Now()).UnixNano()-1)
		if err != nil {
			return
		}

		//send task
		_, err = server.SendChain(chain)
		if err != nil {
			log.Runtime().Errorf("periodic task failed. task name is: %s. error is %s", name, err.Error())
		}
	}

	_, err = server.scheduler.AddFunc(spec, f)
	return err
}

// RegisterPeriodicGroup register a periodic group which will be triggered periodically
func (server *Server) RegisterPeriodicGroup(spec, name string, sendConcurrency int, signatures ...*tasks.Signature) error {
	//check spec
	schedule, err := cron.ParseStandard(spec)
	if err != nil {
		return err
	}

	f := func() {
		// new group
		group, _ := tasks.NewGroup(tasks.CopySignatures(signatures...)...)

		//get lock
		err := server.lock.LockWithRetries(utils.GetLockName(name, spec), schedule.Next(time.Now()).UnixNano()-1)
		if err != nil {
			return
		}

		//send task
		_, err = server.SendGroup(group, sendConcurrency)
		if err != nil {
			log.Runtime().Errorf("periodic task failed. task name is: %s. error is %s", name, err.Error())
		}
	}

	_, err = server.scheduler.AddFunc(spec, f)
	return err
}

// RegisterPeriodicChord register a periodic chord which will be triggered periodically
func (server *Server) RegisterPeriodicChord(spec, name string, sendConcurrency int, callback *tasks.Signature, signatures ...*tasks.Signature) error {
	//check spec
	schedule, err := cron.ParseStandard(spec)
	if err != nil {
		return err
	}

	f := func() {
		// new chord
		group, _ := tasks.NewGroup(tasks.CopySignatures(signatures...)...)
		chord, _ := tasks.NewChord(group, tasks.CopySignature(callback))

		//get lock
		err := server.lock.LockWithRetries(utils.GetLockName(name, spec), schedule.Next(time.Now()).UnixNano()-1)
		if err != nil {
			return
		}

		//send task
		_, err = server.SendChord(chord, sendConcurrency)
		if err != nil {
			log.Runtime().Errorf("periodic task failed. task name is: %s. error is %s", name, err.Error())
		}
	}

	_, err = server.scheduler.AddFunc(spec, f)
	return err
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
