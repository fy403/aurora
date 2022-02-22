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
	exampletasks "aurora/internal/example/tasks"
	eagerlock "aurora/internal/locks/eager"
	"aurora/internal/opentracing/tracers"

	"aurora/internal/log"
	"aurora/internal/tasks"
)

func (this *Worker) HTTPHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (this *Worker) StartHttpServer() (err error) {
	var port = this.cfg.HTTP.Port
	if port == "" {
		port = ":8080"
	}
	l, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}

	http.HandleFunc("/health", this.HTTPHealth)
	// http.Handle("/metrics", promhttp.Handler())

	go http.Serve(l, nil)
	log.Runtime().Infof("http started on %s", port)
	return nil
}

func (this *Worker) InitMetrics() (err error) {
	// if err = metrics.InitMetrics(global.Region(), config.AppTag, this.cfg.Files.Metrics, ""); err != nil {
	// 	return err
	// }
	return nil
}

func (this *Worker) InitLogs() (err error) {
	if err = log.InitLog(this.cfg.Files.Log); err != nil {
		return err
	}
	return nil
}

func (this *Worker) Init() (err error) {
	// load config
	if err = config.AppInitConfig(); err != nil {
		log.Runtime().Fatalf("config init error: %s", err.Error())
		return err
	}
	this.cfg = config.GetAppConfig()

	// init logs
	if err = this.InitLogs(); err != nil {
		log.Runtime().Errorf("logs init error: %s", err.Error())
	}

	// init metrics
	if err = this.InitMetrics(); err != nil {
		log.Runtime().Errorf("metrics init error: %s", err.Error())
	}

	// Only Load worker config
	var cfg = this.cfg.Worker
	if cfg == nil {
		log.Runtime().Fatal("cfg.Worker must be set")
		return
	}
	// If AMQP/MongoDB driver is used here
	if cfg.ResultBackend == "" || cfg.Broker == "" || (strings.Index(cfg.Broker, "amqp") != -1 && cfg.AMQP == nil) {
		log.Runtime().Fatal("cfg.Worker.AMQP must be set")
		return
	}

	// Create server instance
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

	lock := eagerlock.New()
	this.server = center.NewServer(cfg, broker, backend, lock, true)
	// Register example tasks
	tasksMap := map[string]interface{}{
		"add":               exampletasks.Add,
		"multiply":          exampletasks.Multiply,
		"sum_ints":          exampletasks.SumInts,
		"sum_floats":        exampletasks.SumFloats,
		"concat":            exampletasks.Concat,
		"split":             exampletasks.Split,
		"panic_task":        exampletasks.PanicTask,
		"long_running_task": exampletasks.LongRunningTask,
	}
	err = this.server.RegisterTasks(tasksMap)
	if err != nil {
		log.Runtime().Fatalf("RegisterTasks process error:", err)
		return
	}

	// Set worker subscribe configure
	hostname, err := os.Hostname()
	if err != nil {
		this.ConsumerTag = fmt.Sprintf("aurora_worker_id_%d", time.Now().Unix())
	} else {
		this.ConsumerTag = hostname
	}
	this.Concurrency = cfg.Concurrency
	this.Queue = cfg.Queue

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

	this.SetPostTaskHandler(postTaskHandler)
	this.SetErrorHandler(errorHandler)
	this.SetPreTaskHandler(preTaskHandler)
	return
}

func (this *Worker) Run() (err error) {
	// let worker run
	log.Runtime().Infof("Worker Has Running")

	// Setup opentracing
	opentracingCfg := this.cfg.Opentracing
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
	if err = this.StartHttpServer(); err != nil {
		log.Runtime().Errorf("http server start faild: %s", err.Error())
	}
	// Continuous operation until a CTRL+C
	return this.Launch()
}

func (this *Worker) Stop() (err error) {
	this.Quit()
	return
}
