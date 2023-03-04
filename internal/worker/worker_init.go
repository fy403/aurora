package worker

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"
	"strings"
	"time"

	// "github.com/prometheus/client_golang/prometheus/promhttp"
	mongobackend "aurora/internal/backends/mongo"
	amqpbroker "aurora/internal/brokers/amqp"
	eagercache "aurora/internal/cache/eager"
	cachesiface "aurora/internal/cache/iface"
	rediscache "aurora/internal/cache/redis"
	"aurora/internal/center"
	"aurora/internal/config"
	eagerlock "aurora/internal/locks/eager"
	locksiface "aurora/internal/locks/iface"
	redislock "aurora/internal/locks/redis"
	"aurora/internal/log"
	"aurora/internal/model"
	"aurora/internal/opentracing/tracers"
	"aurora/internal/request"
	"aurora/internal/tasks"
	"aurora/internal/utils"

	"github.com/google/uuid"
)

func (worker *Worker) httpHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (worker *Worker) startHttpServer() (err error) {
	// var port = worker.cfg.HTTP.Port
	// if port == "" {
	// 	port = ":8080"
	// }
	// l, err := net.Listen("tcp", port)
	// if err != nil {
	// 	return err
	// }

	// http.HandleFunc("/health", worker.httpHealth)
	// http.Handle("/metrics", promhttp.Handler())

	// go http.Serve(l, nil)
	// log.Runtime().Infof("http started on %s", port)
	return nil
}

// Set Worker info to backend
func (worker *Worker) register() (err error) {
	labels := worker.cfg.Worker.Labels
	if len(labels) == 0 {
		return nil
	}
	queueName := strconv.Itoa(int(utils.Hash32WithMap(labels)))

	handlers := []*request.Handler{}
	for _, handler := range model.ExtantTaskMap {
		handlers = append(handlers, handler)
	}
	req := request.WorkerRequest{
		UUID:      uuid.New().String(),
		SpecQueue: queueName,
		Metrics:   nil,
		Handlers:  handlers,
		Labels:    worker.cfg.Worker.Labels,
		Timestamp: time.Now().Unix(),
	}

	errChan := make(chan error, 10)
	errChan <- worker.GetServer().SetWorkerInfo(&req)
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

	go func() {
		ticker := time.NewTicker(time.Second * 15)
		defer ticker.Stop()
		for {
			select {
			case err := <-errChan:
				if err != nil {
					log.Runtime().Infof("SetWorkerInfo has occur some err: %v", err)
				}
			case <-ticker.C:
				handlers = handlers[:0]
				for _, handler := range model.ExtantTaskMap {
					handlers = append(handlers, handler)
				}
				worker.server.RegisterTasks(model.ExtantTaskMap)
				errChan <- worker.GetServer().SetWorkerInfo(&req)
			}
		}
	}()

	// Purge invalid worker
	// results, err := worker.GetServer().GetAllWorkersInfo()
	// for idx, result := range results {
	// 	if isValid := result.IsValid(worker.cfg.Worker.BrokerApi); !isValid && result.SpecQueue != queueName {
	// 		results[idx] = nil
	// 		req := request.WorkerRequest{
	// 			UUID: result.UUID,
	// 		}
	// 		worker.GetServer().PurgeWorkerInfo(&req)
	// 		continue
	// 	}
	// }
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

	var lock locksiface.Lock
	var cache cachesiface.Cache
	if strings.Contains(cfg.Lock, "redis") {
		// 分布式锁
		lock = redislock.New(cfg)
	} else {
		// 本地锁
		lock = eagerlock.New()
	}
	if strings.Contains(cfg.Cache, "redis") {
		// 分布式缓存
		cache = rediscache.New(cfg)
	} else {
		// 本地缓存
		cache = eagercache.New()
	}
	// Create server instance
	worker.server = center.NewServer(cfg, broker, backend, cache, lock, true)

	// Register example tasks
	err = worker.server.RegisterTasks(model.ExtantTaskMap)

	if err != nil {
		log.Runtime().Fatalf("RegisterTasks process error:", err)
		return
	}

	// Register faas instance
	err = worker.server.RegisterFaas(worker.cfg.Faas)
	if err != nil {
		log.Runtime().Fatalf("RegisterFaas process error:", err)
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
