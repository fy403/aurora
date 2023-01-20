package worker

import (
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strings"
	"time"

	// "github.com/prometheus/client_golang/prometheus/promhttp"
	mongobackend "aurora/internal/backends/mongo"
	amqpbroker "aurora/internal/brokers/amqp"
	"aurora/internal/center"
	"aurora/internal/config"
	eagerlock "aurora/internal/locks/eager"
	"aurora/internal/model"
	"aurora/internal/opentracing/tracers"
	"aurora/internal/request"

	"aurora/internal/log"
	"aurora/internal/tasks"

	"github.com/google/uuid"
)

func (worker *Worker) httpHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (worker *Worker) startHttpServer() (err error) {
	var port = worker.cfg.HTTP.Port
	if port == "" {
		port = ":8080"
	}
	l, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}

	http.HandleFunc("/health", worker.httpHealth)
	// http.Handle("/metrics", promhttp.Handler())

	go http.Serve(l, nil)
	log.Runtime().Infof("http started on %s", port)
	return nil
}

// Set Worker info to backend
func (worker *Worker) register() (err error) {
	labels := worker.cfg.Worker.Labels
	if len(labels) == 0 {
		return nil
	}
	queueName := uuid.New().String()
	go func() {
		for {
			// CreateSpecQueue and Continuous consumption
			retry, err := worker.server.GetBroker().CreateSpecQueue(queueName, worker.ConsumerTag, worker.Concurrency, worker)
			if retry {
				if worker.errorHandler != nil {
					worker.errorHandler(err)
				} else {
					log.Runtime().Warnf("Broker failed with error: %s", err)
				}
			} else {
				log.Runtime().Fatal("register daemon has dead with too many retry")
				return
			}
		}
	}()
	_id := "worker_" + uuid.New().String()
	handlers := []*request.Handler{}
	for _, handler := range model.ExtantTaskMap {
		handlers = append(handlers, handler)
	}
	req := request.WorkerRequest{
		UUID:      _id,
		SpecQueue: queueName,
		Metrics:   nil,
		Handlers:  handlers,
		Labels:    worker.cfg.Worker.Labels,
		Timestamp: time.Now().Unix(),
	}

	results, err := worker.server.GetBackend().GetAllWorkersInfo()
	var filterResults []*request.WorkerResponse
	// Purge invalid worker
	for idx, result := range results {
		if isValid := result.IsValid(worker.cfg.Center.BrokerApi); !isValid || result.SpecQueue == queueName {
			results[idx] = nil
			req := request.WorkerRequest{
				UUID: result.UUID,
			}
			worker.server.GetBackend().PurgeWorkerInfo(&req)
			continue
		}
		filterResults = append(filterResults, result)
	}

	err = worker.server.GetBackend().SetWorkerInfo(&req)
	return
}

func (worker *Worker) initMetrics() (err error) {
	// if err = metrics.InitMetrics(global.Region(), config.AppTag, worker.cfg.Files.Metrics, ""); err != nil {
	// 	return err
	// }
	return nil
}

func (worker *Worker) initLogs() (err error) {
	if err = log.InitLog(worker.cfg.Files.Log); err != nil {
		return err
	}
	return nil
}

func (worker *Worker) Init() (err error) {
	// load config
	if err = config.AppInitConfig(); err != nil {
		log.Runtime().Fatalf("config init error: %s", err.Error())
		return err
	}
	worker.cfg = config.GetAppConfig()

	// init logs
	if err = worker.initLogs(); err != nil {
		log.Runtime().Errorf("logs init error: %s", err.Error())
	}

	// init metrics
	if err = worker.initMetrics(); err != nil {
		log.Runtime().Errorf("metrics init error: %s", err.Error())
	}

	// Only Load worker config
	var cfg = worker.cfg.Worker
	if cfg == nil {
		log.Runtime().Fatal("cfg.Worker must be set")
		return
	}
	// If AMQP/MongoDB driver is used here
	if cfg.ResultBackend == "" || cfg.Broker == "" || (strings.Index(cfg.Broker, "amqp") != -1 && cfg.AMQP == nil) {
		log.Runtime().Fatal("cfg.Worker.AMQP must be set")
		return
	}

	// Create broker, backend instance
	broker := amqpbroker.New(cfg)
	backend, err := mongobackend.New(cfg)
	if err != nil {
		log.Runtime().Fatalf("Unable to instantiate a mongobackend:", err)
		return
	}

	// Test broker and backend connection
	if err = broker.TestConnect(); err != nil {
		log.Runtime().Fatalf("Can`t build a connection to broker: %v", err)
		return
	}
	if err = backend.TestConnect(); err != nil {
		log.Runtime().Fatalf("Can`t build a connection to backend: %v", err)
		return
	}

	// Create server instance
	lock := eagerlock.New()
	worker.server = center.NewServer(cfg, broker, backend, lock, true)

	// Register example tasks
	err = worker.server.RegisterTasks(model.ExtantTaskMap)
	if err != nil {
		log.Runtime().Fatalf("RegisterTasks process error:", err)
		return
	}

	log.Runtime().Infof("RegisterTasks are %v", worker.server.GetRegisteredTaskNames())

	// Set worker subscribe configure
	hostname, err := os.Hostname()
	if err != nil {
		worker.ConsumerTag = fmt.Sprintf("aurora_worker_id_%d", time.Now().Unix())
	} else {
		worker.ConsumerTag = hostname
	}
	worker.Concurrency = cfg.Concurrency
	// Required, setting worker subscribe queue
	worker.Queue = cfg.Queue

	// Set task state transition handler
	errorHandler := func(err error) {
		log.Runtime().Errorf("I am an error handler:", err)
	}
	preTaskHandler := func(signature *tasks.Signature) {
		log.Runtime().Infof("I am a start of task handler for:", signature.Name)
	}
	postTaskHandler := func(signature *tasks.Signature) {
		log.Runtime().Infof("I am an end of task handler for:", signature.Name)
	}

	worker.SetPostTaskHandler(postTaskHandler)
	worker.SetErrorHandler(errorHandler)
	worker.SetPreTaskHandler(preTaskHandler)

	// Register the queue of the current instance with the center and subscribe
	if err = worker.register(); err != nil {
		log.Runtime().Fatalf("Can`t register instance queue to center: %v", err)
		return
	}
	return
}

func (worker *Worker) Run() (err error) {
	// let worker run
	log.Runtime().Infof("Worker Has Running")

	// Setup opentracing
	opentracingCfg := worker.cfg.Opentracing
	serviceName := "aurora_worker"
	if opentracingCfg.ServiceName != "" {
		serviceName = opentracingCfg.ServiceName
	}
	cleanup, err := tracers.SetupTracer(serviceName, opentracingCfg.CollectorEndpoint, opentracingCfg.LogSpans)
	if err != nil {
		log.Runtime().Fatalf("Unable to instantiate a tracer:", err)
	}
	defer cleanup()

	// start a http server
	if err = worker.startHttpServer(); err != nil {
		log.Runtime().Errorf("http server start faild: %s", err.Error())
	}
	// Continuous operation until a CTRL+C
	return worker.Launch()
}

func (worker *Worker) Stop() (err error) {
	worker.Quit()
	return
}
