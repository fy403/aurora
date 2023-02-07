package worker

import (
	backendsiface "aurora/internal/backends/iface"
	"aurora/internal/brokers/errs"
	brokersiface "aurora/internal/brokers/iface"
	cachesiface "aurora/internal/cache/iface"
	"aurora/internal/center"
	"aurora/internal/config"
	lockiface "aurora/internal/locks/iface"
	"aurora/internal/log"
	"aurora/internal/opentracing/tracing"
	"aurora/internal/retry"
	"aurora/internal/tasks"
	"aurora/internal/utils"
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/opentracing/opentracing-go"
)

// Worker represents a single worker process
type Worker struct {
	server            *center.Server
	ConsumerTag       string
	Concurrency       int
	Queue             string
	errorHandler      func(err error)
	preTaskHandler    func(*tasks.Signature)
	postTaskHandler   func(*tasks.Signature)
	preConsumeHandler func(*Worker) bool
	cfg               *config.AppConfig
	// cancel context.CancelFunc
}

var (
	// ErrWorkerQuitGracefully is return when worker quit gracefully
	ErrWorkerQuitGracefully = errors.New("Worker quit gracefully")
	// ErrWorkerQuitGracefully is return when worker quit abruptly
	ErrWorkerQuitAbruptly = errors.New("Worker quit abruptly")
)

// NewWorker creates a worker instance
func (worker *Worker) NewWorker(
	cnf *config.Config, consumerTag string, concurrency int,
	brokerServer brokersiface.Broker,
	backendServer backendsiface.Backend,
	cache cachesiface.Cache,
	lock lockiface.Lock,
) *Worker {
	srv := center.NewServer(cnf, brokerServer, backendServer, cache, lock, true)
	return &Worker{
		server:      srv,
		ConsumerTag: consumerTag,
		Concurrency: concurrency,
		Queue:       "",
	}
}

// NewCustomQueueWorker creates Worker instance with Custom Queue
func (worker *Worker) NewCustomQueueWorker(
	cnf *config.Config, consumerTag string, concurrency int,
	brokerServer brokersiface.Broker,
	backendServer backendsiface.Backend,
	cache cachesiface.Cache,
	lock lockiface.Lock, queue string) *Worker {
	srv := center.NewServer(cnf, brokerServer, backendServer, cache, lock, true)
	return &Worker{
		server:      srv,
		ConsumerTag: consumerTag,
		Concurrency: concurrency,
		Queue:       queue,
	}
}

// Launch starts a new worker process. The worker subscribes
// to the default queue and processes incoming registered tasks
func (worker *Worker) Launch() error {
	errorsChan := make(chan error)

	worker.LaunchAsync(errorsChan)

	return <-errorsChan
}

// LaunchAsync is a non blocking version of Launch
func (worker *Worker) LaunchAsync(errorsChan chan<- error) {
	cnf := worker.GetServer().GetConfig()
	broker := worker.GetServer().GetBroker()

	// Log some useful information about worker configuration
	log.Runtime().Infof("Launching a worker with the following settings:")
	log.Runtime().Infof("- Broker: %s", RedactURL(cnf.Broker))
	log.Runtime().Infof("- Labels: %s", cnf.Labels)
	if worker.Queue == "" {
		log.Runtime().Infof("- DefaultQueue: %s", cnf.DefaultQueue)
	} else {
		log.Runtime().Infof("- CustomQueue: %s", worker.Queue)
	}
	log.Runtime().Infof("- ResultBackend: %s", RedactURL(cnf.ResultBackend))
	if cnf.AMQP != nil {
		log.Runtime().Infof("- AMQP: %s", cnf.AMQP.Exchange)
		log.Runtime().Infof("  - Exchange: %s", cnf.AMQP.Exchange)
		log.Runtime().Infof("  - ExchangeType: %s", cnf.AMQP.ExchangeType)
		log.Runtime().Infof("  - BindingKey: %s", cnf.AMQP.BindingKey)
		log.Runtime().Infof("  - PrefetchCount: %d", cnf.AMQP.PrefetchCount)
	}

	var signalWG sync.WaitGroup
	// Goroutine to start broker consumption and handle retries when broker connection dies
	go func() {
		for {
			retry, err := broker.StartConsuming(worker.ConsumerTag, worker.Concurrency, worker)

			if retry {
				if worker.errorHandler != nil {
					worker.errorHandler(err)
				} else {
					log.Runtime().Warnf("Broker failed with error: %s", err)
				}
			} else {
				signalWG.Wait()
				errorsChan <- err // stop the goroutine
				return
			}
		}
	}()
	if !cnf.NoUnixSignals {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		var signalsReceived uint

		// Goroutine Handle SIGINT and SIGTERM signals
		go func() {
			for s := range sig {
				log.Runtime().Warnf("Signal received: %v", s)
				signalsReceived++

				if signalsReceived < 2 {
					// After first Ctrl+C start quitting the worker gracefully
					log.Runtime().Warn("Waiting for running tasks to finish before shutting down")
					signalWG.Add(1)
					go func() {
						worker.Quit()
						errorsChan <- ErrWorkerQuitGracefully
						signalWG.Done()
					}()
				} else {
					// Abort the program when user hits Ctrl+C second time in a row
					errorsChan <- ErrWorkerQuitAbruptly
				}
			}
		}()
	}
}

// CustomQueue returns Custom Queue of the running worker process
func (worker *Worker) CustomQueue() string {
	return worker.Queue
}

// Quit tears down the running worker process
func (worker *Worker) Quit() {
	worker.GetServer().GetBroker().StopConsuming()
}

// Process handles received tasks and triggers success/error callbacks
func (worker *Worker) Process(signature *tasks.Signature) error {
	// If the task is not registered with this worker, do not continue
	// but only return nil as we do not want to restart the worker process
	if !worker.GetServer().IsTaskRegistered(signature.Name) {
		return nil
	}

	taskFunc, err := worker.GetServer().GetRegisteredTask(signature.Name)
	if err != nil {
		return nil
	}

	// Update task state to RECEIVED
	if err = worker.GetServer().GetBackend().SetStateReceived(signature); err != nil {
		return fmt.Errorf("Set state to 'received' for task %s returned error: %s", signature.UUID, err)
	}

	// Prepare task for processing
	task, err := tasks.NewWithSignature(taskFunc, signature)
	// if this failed, it means the task is malformed, probably has invalid
	// signature, go directly to task failed without checking whether to retry
	if err != nil {
		worker.taskFailed(signature, err)
		return err
	}

	// try to extract trace span from headers and add it to the function context
	// so it can be used inside the function if it has context.Context as the first
	// argument. Start a new span if it isn't found.
	taskSpan := tracing.StartSpanFromHeaders(signature.Headers, signature.Name)
	tracing.AnnotateSpanWithSignatureInfo(taskSpan, signature)
	task.Context = opentracing.ContextWithSpan(task.Context, taskSpan)

	// Update task state to STARTED
	if err = worker.GetServer().GetBackend().SetStateStarted(signature); err != nil {
		return fmt.Errorf("Set state to 'started' for task %s returned error: %s", signature.UUID, err)
	}

	//Run handler before the task is called
	if worker.preTaskHandler != nil {
		worker.preTaskHandler(signature)
	}

	//Defer run handler for the end of the task
	if worker.postTaskHandler != nil {
		defer worker.postTaskHandler(signature)
	}

	// Call the task
	results, err := task.Call()
	if err != nil {
		// If a tasks.ErrRetryTaskLater was returned from the task,
		// retry the task after specified duration
		retriableErr, ok := interface{}(err).(tasks.ErrRetryTaskLater)
		if ok {
			return worker.retryTaskIn(signature, retriableErr.RetryIn())
		}

		// Otherwise, execute default retry logic based on signature.RetryCount
		// and signature.RetryTimeout values
		if signature.RetryCount > 0 {
			return worker.taskRetry(signature)
		}

		return worker.taskFailed(signature, err)
	}

	return worker.taskSucceeded(signature, results, task.Context)
}

// retryTask decrements RetryCount counter and republishes the task to the queue
func (worker *Worker) taskRetry(signature *tasks.Signature) error {
	// Update task state to RETRY
	if err := worker.GetServer().GetBackend().SetStateRetry(signature); err != nil {
		return fmt.Errorf("Set state to 'retry' for task %s returned error: %s", signature.UUID, err)
	}

	// Decrement the retry counter, when it reaches 0, we won't retry again
	signature.RetryCount--

	// Increase retry timeout
	signature.RetryTimeout = retry.FibonacciNext(signature.RetryTimeout)

	// Delay task by signature.RetryTimeout seconds
	eta := time.Now().UTC().Add(time.Second * time.Duration(signature.RetryTimeout))
	signature.ETA = &eta

	log.Runtime().Warnf("Task %s failed. Going to retry in %d seconds.", signature.UUID, signature.RetryTimeout)

	// Send the task back to the queue
	_, err := worker.GetServer().SendTask(signature)
	return err
}

// taskRetryIn republishes the task to the queue with ETA of now + retryIn.Seconds()
func (worker *Worker) retryTaskIn(signature *tasks.Signature, retryIn time.Duration) error {
	// Update task state to RETRY
	if err := worker.GetServer().GetBackend().SetStateRetry(signature); err != nil {
		return fmt.Errorf("Set state to 'retry' for task %s returned error: %s", signature.UUID, err)
	}

	// Delay task by retryIn duration
	eta := time.Now().UTC().Add(retryIn)
	signature.ETA = &eta

	log.Runtime().Warnf("Task %s failed. Going to retry in %.0f seconds.", signature.UUID, retryIn.Seconds())

	// Send the task back to the queue
	_, err := worker.GetServer().SendTask(signature)
	return err
}

// taskSucceeded updates the task state and triggers success callbacks or a
// chord callback if this was the last task of a group with a chord callback
func (worker *Worker) taskSucceeded(signature *tasks.Signature, taskResults []*tasks.TaskResult, ctx context.Context) (err error) {
	// Update task state to SUCCESS
	if err := worker.GetServer().GetBackend().SetStateSuccess(signature, taskResults); err != nil {
		return fmt.Errorf("Set state to 'success' for task %s returned error: %s", signature.UUID, err)
	}

	// Log human readable results of the processed task
	// var debugResults = "[]"
	// results, err := tasks.ReflectTaskResults(taskResults)
	// if err != nil {
	// 	log.Runtime().Warn(err.Error())
	// } else {
	// 	debugResults = tasks.HumanReadableResults(results)
	// }
	// log.Runtime().Debugf("Processed task %s. Results = %s", signature.UUID, debugResults)

	// Trigger success callbacks
	for _, successTask := range signature.OnSuccess {
		if signature.Immutable == false {
			// Pass results of the task to success callbacks
			for _, taskResult := range taskResults {
				successTask.Args = append(successTask.Args, tasks.Arg{
					Type:  taskResult.Type,
					Value: taskResult.Value,
				})
			}
		}

		worker.GetServer().SendChainTaskWithContext(ctx, successTask)
	}

	// Trigger graph next sequences
	if signature.GraphUUID != "" {
		// 不允许重复读取，互斥锁
		expiration := 4 * time.Second
		err = worker.GetServer().GetLock().LockWithRetries(utils.GetLockName(signature.GraphUUID, ""), int64(expiration))
		if err != nil {
			return err
		}
		lockedTime := time.Now()
		defer func() {
			if time.Since(lockedTime) < expiration {
				worker.GetServer().GetLock().UnLock(utils.GetLockName(signature.GraphUUID, ""))
			}
		}()
		isFinished, err := worker.GetServer().GetBackend().GraphCompleted(signature.GraphUUID, signature.GraphTaskCount)
		if err != nil {
			return fmt.Errorf("Get complemented state for graph %s returned error: %s", signature.GraphUUID, err)
		}
		if isFinished {
			return nil
		}

		graph, err := worker.GetServer().GetBackend().GraphStates(signature.GraphUUID)
		if err != nil {
			return fmt.Errorf("Get state for graph %s returned error: %s", signature.GraphUUID, err)
		}
		inDegrees := make([]int, graph.VexNum)
		for i := 0; i < graph.VexNum; i++ {
			for j := 0; j < graph.VexNum; j++ {
				if graph.Edge[i][j] == 1 {
					inDegrees[j]++
				}
			}
		}
		index := -1
		for idx, vertex := range graph.Vertexes {
			if vertex.UUID == signature.UUID {
				index = idx
			}
		}
		if index == -1 {
			return fmt.Errorf("graph %s not contain %s", signature.GraphUUID, signature.UUID)
		}
		for i := 0; i < graph.VexNum; i++ {
			if graph.Edge[index][i] == 1 {
				inDegrees[i]--
				graph.Edge[index][i] = 0
				if inDegrees[i] == 0 {
					// Send the next graph task
					_, err = worker.GetServer().SendTaskWithContext(ctx, graph.Vertexes[i])
					if err != nil {
						return err
					}
				}
			}
		}
		if err = worker.GetServer().GetBackend().UpdateGraphStates(graph); err != nil {
			return err
		}
		return nil
	}

	// If the task was not part of a group, just return
	if signature.GroupUUID == "" {
		return nil
	}

	// There is no chord callback, just return
	if signature.ChordCallback == nil {
		return nil
	}

	// Check if all task in the group has completed
	groupCompleted, err := worker.GetServer().GetBackend().GroupCompleted(
		signature.GroupUUID,
		signature.GroupTaskCount,
	)
	if err != nil {
		return fmt.Errorf("Completed check for group %s returned error: %s", signature.GroupUUID, err)
	}

	// If the group has not yet completed, just return
	if !groupCompleted {
		return nil
	}

	// Defer purging of group meta queue if we are using AMQP backend
	// if worker.hasAMQPBackend() {
	// 	defer worker.GetServer().GetBackend().PurgeGroupMeta(signature.GroupUUID)
	// }
	//  读写互斥锁，运行重复读
	expiration := 4 * time.Second
	err = worker.GetServer().GetLock().Lock(utils.GetLockName(signature.GroupUUID, signature.ChordCallback.UUID), int64(expiration))
	if err != nil {
		// 枪锁失败，退出触发后续
		return nil
	}
	// Trigger chord callback
	shouldTrigger, err := worker.GetServer().GetBackend().TriggerChord(signature.GroupUUID)
	if err != nil {
		return fmt.Errorf("Triggering chord for group %s returned error: %s", signature.GroupUUID, err)
	}

	// Chord has already been triggered
	if !shouldTrigger {
		return nil
	}

	// Get task states
	taskStates, err := worker.GetServer().GetBackend().GroupTaskStates(
		signature.GroupUUID,
		signature.GroupTaskCount,
	)
	if err != nil {
		log.Runtime().Errorf(
			"Failed to get tasks states for group:[%s]. Task count:[%d]. The chord may not be triggered. Error:[%s]",
			signature.GroupUUID,
			signature.GroupTaskCount,
			err,
		)
		return nil
	}

	// Append group tasks' return values to chord task if it's not immutable
	for _, taskState := range taskStates {
		if !taskState.IsSuccess() {
			return nil
		}

		if signature.ChordCallback.Immutable == false {
			// Pass results of the task to the chord callback
			for _, taskResult := range taskState.Results {
				signature.ChordCallback.Args = append(signature.ChordCallback.Args, tasks.Arg{
					Type:  taskResult.Type,
					Value: taskResult.Value,
				})
			}
		}
	}

	// Send the chord task
	_, err = worker.GetServer().SendChordCallbackWithContext(ctx, signature.ChordCallback)
	if err != nil {
		return err
	}

	return nil
}

// taskFailed updates the task state and triggers error callbacks
func (worker *Worker) taskFailed(signature *tasks.Signature, taskErr error) error {
	// Update task state to FAILURE
	if err := worker.GetServer().GetBackend().SetStateFailure(signature, taskErr.Error()); err != nil {
		return fmt.Errorf("Set state to 'failure' for task %s returned error: %s", signature.UUID, err)
	}

	if worker.errorHandler != nil {
		worker.errorHandler(taskErr)
	} else {
		log.Runtime().Errorf("Failed processing task %s. Error = %v", signature.UUID, taskErr)
	}

	// Trigger error callbacks
	for _, errorTask := range signature.OnError {
		// Pass error as a first argument to error callbacks
		args := append([]tasks.Arg{{
			Type:  "string",
			Value: taskErr.Error(),
		}}, errorTask.Args...)
		errorTask.Args = args
		worker.GetServer().SendTask(errorTask)
	}

	if signature.StopTaskDeletionOnError {
		return errs.ErrStopTaskDeletion
	}

	return nil
}

// Returns true if the worker uses AMQP backend
// func (worker *Worker) hasAMQPBackend() bool {
// 	_, ok := worker.GetServer().GetBackend().(*amqp.Backend)
// 	return ok
// }

// SetErrorHandler sets a custom error handler for task errors
// A default behavior is just to log the error after all the retry attempts fail
func (worker *Worker) SetErrorHandler(handler func(err error)) {
	worker.errorHandler = handler
}

// SetPreTaskHandler sets a custom handler func before a job is started
func (worker *Worker) SetPreTaskHandler(handler func(*tasks.Signature)) {
	worker.preTaskHandler = handler
}

// SetPostTaskHandler sets a custom handler for the end of a job
func (worker *Worker) SetPostTaskHandler(handler func(*tasks.Signature)) {
	worker.postTaskHandler = handler
}

// SetPreConsumeHandler sets a custom handler for the end of a job
func (worker *Worker) SetPreConsumeHandler(handler func(*Worker) bool) {
	worker.preConsumeHandler = handler
}

// GetServer returns server
func (worker *Worker) GetServer() *center.Server {
	return worker.server
}

func (worker *Worker) PreConsumeHandler() bool {
	if worker.preConsumeHandler == nil {
		return true
	}

	return worker.preConsumeHandler(worker)
}

func RedactURL(urlString string) string {
	u, err := url.Parse(urlString)
	if err != nil {
		return urlString
	}
	return fmt.Sprintf("%s://%s", u.Scheme, u.Host)
}
