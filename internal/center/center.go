package center

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"time"

	opentracing_log "github.com/opentracing/opentracing-go/log"

	// "github.com/prometheus/client_golang/prometheus/promhttp"
	"aurora/internal/auth"
	mongobackend "aurora/internal/backends/mongo"
	"aurora/internal/backends/result"
	amqpbroker "aurora/internal/brokers/amqp"
	"aurora/internal/config"
	eagerlock "aurora/internal/locks/eager"
	"aurora/internal/log"
	"aurora/internal/opentracing/tracers"
	"aurora/internal/request"

	"aurora/internal/tasks"

	"github.com/google/uuid"
	opentracing "github.com/opentracing/opentracing-go"
)

type Center struct {
	server *Server
	srv    *http.Server
	cfg    *config.AppConfig
}

func NewCenter() *Center {
	return &Center{}
}

func (this *Center) HTTPHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (this *Center) HTTPAuth(w http.ResponseWriter, r *http.Request) {
	auth.Login(w, r)
}

func (this *Center) HTTPTasksTouch(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Must use POST", http.StatusBadRequest)
		return
	}

	if ok := auth.Authentication(w, r); !ok {
		return
	}

	strByte, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Read req.Body failed", http.StatusBadRequest)
		return
	}

	requestOBJ := &request.CenterRequest{}
	decoder := json.NewDecoder(bytes.NewReader(strByte))
	decoder.UseNumber()

	if err := decoder.Decode(requestOBJ); err != nil {
		http.Error(w, fmt.Sprintf("Unexpected request Unmarshal format: %v", requestOBJ), http.StatusBadRequest)
		return
	}

	if err := requestOBJ.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to validate format: %v", requestOBJ), http.StatusBadRequest)
		return
	}

	if err := requestOBJ.Inject(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	/*
	 * Send the corresponding query according to the TaskType matching result
	 */
	responseOBJ := request.CenterResponse{
		UUID:          requestOBJ.UUID,
		User:          requestOBJ.User,
		BatchID:       requestOBJ.BatchID,
		Timestamp:     time.Now().Local().Unix(),
		TaskType:      requestOBJ.TaskType,
		TaskResponses: []*request.TaskResponse{},
	}
	switch v := requestOBJ.TaskType; v {
	case "task":
		asyncResult := result.NewAsyncResult(requestOBJ.Signatures[0], this.server.backend)
		results, err := asyncResult.Touch()
		if err != nil {
			http.Error(w, fmt.Sprintf("Task has failed with error: %s", err.Error()), http.StatusBadGateway)
			return
		}
		responseOBJ.TaskResponses = append(responseOBJ.TaskResponses, &request.TaskResponse{
			Results:    tasks.InterfaceReadableResults(results),
			Signatures: requestOBJ.Signatures,
		})
	case "group":
		for _, signature := range requestOBJ.Signatures {
			asyncResult := result.NewAsyncResult(signature, this.server.backend)
			results, err := asyncResult.Touch()
			if err != nil {
				http.Error(w, fmt.Sprintf("Task has failed with error: %s", err.Error()), http.StatusBadGateway)
				return
			}
			responseOBJ.TaskResponses = append(responseOBJ.TaskResponses, &request.TaskResponse{
				Results: tasks.InterfaceReadableResults(results),
				Signatures: []*tasks.Signature{
					signature,
				},
			})
		}
	case "chord":
		chordAsyncResult := result.NewChordAsyncResult(requestOBJ.Signatures, requestOBJ.CallBack, this.server.backend)
		for _, asyncResult := range chordAsyncResult.GetGroupAsyncResults() {
			_, err := asyncResult.Touch()
			if err != nil {
				http.Error(w, fmt.Sprintf("Task has failed with error: %s", err.Error()), http.StatusBadGateway)
				return
			}
		}
		results, err := chordAsyncResult.GetChordAyncResults().Touch()
		if err != nil {
			http.Error(w, fmt.Sprintf("Task has failed with error: %s", err.Error()), http.StatusBadGateway)
			return
		}
		responseOBJ.TaskResponses = append(responseOBJ.TaskResponses, &request.TaskResponse{
			Results:    tasks.InterfaceReadableResults(results),
			Signatures: requestOBJ.Signatures,
			CallBack:   requestOBJ.CallBack,
		})
	case "chain":
		chainAsyncResult := result.NewChainAsyncResult(requestOBJ.Signatures, this.server.backend)
		var results []reflect.Value
		for _, asyncResult := range chainAsyncResult.GetAsyncResults() {
			results, err = asyncResult.Touch()
			if err != nil {
				http.Error(w, fmt.Sprintf("Task has failed with error: %s", err.Error()), http.StatusBadGateway)
				return
			}
		}
		responseOBJ.TaskResponses = append(responseOBJ.TaskResponses, &request.TaskResponse{
			Results:    tasks.InterfaceReadableResults(results),
			Signatures: requestOBJ.Signatures,
		})
	default:
		err := errors.New("Unexpected task type: " + v)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	data, err := json.Marshal(responseOBJ)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to json.Marshal responseOBJ: %v", responseOBJ), http.StatusBadGateway)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(data))
}

func (this *Center) HTTPTasksSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Must use POST", http.StatusBadRequest)
		return
	}

	if ok := auth.Authentication(w, r); !ok {
		return
	}

	strByte, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Read req.Body failed", http.StatusBadRequest)
		return
	}

	requestOBJ := &request.CenterRequest{}
	decoder := json.NewDecoder(bytes.NewReader(strByte))
	decoder.UseNumber()

	if err := decoder.Decode(requestOBJ); err != nil {
		http.Error(w, fmt.Sprintf("Unexpected request Unmarshal format: %v", requestOBJ), http.StatusBadRequest)
		return
	}

	if err := requestOBJ.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to validate format: %v", requestOBJ), http.StatusBadRequest)
		return
	}

	if err := requestOBJ.Inject(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	/*
	 * Lets start a span representing this run of the `send` command and
	 * set a batch id as baggage so it can travel all the way into
	 * the worker functions.
	 */
	span, ctx := opentracing.StartSpanFromContext(context.Background(), "send")
	defer span.Finish()

	// Generate an unique id for each request
	batchID := uuid.New().String()
	// sync server config
	requestOBJ.BatchID = batchID
	// Local Span details
	span.SetTag("task.type", requestOBJ.TaskType)
	span.LogFields(opentracing_log.String("params.format", fmt.Sprintf("%#v", requestOBJ)))
	span.LogFields(opentracing_log.String("batch.id", batchID))
	// Span Context：pass across process boundary
	span.SetBaggageItem("user.uuid", requestOBJ.UUID)
	span.SetBaggageItem("user.name", requestOBJ.User)
	span.SetBaggageItem("batch.id", batchID)

	log.Runtime().Infof("Starting batch: %s", batchID)
	time.Local, _ = time.LoadLocation("Asia/Beijing")

	responseOBJ := request.CenterResponse{
		UUID:          requestOBJ.UUID,
		User:          requestOBJ.User,
		BatchID:       requestOBJ.BatchID,
		Timestamp:     time.Now().Local().Unix(),
		TaskType:      requestOBJ.TaskType,
		TaskResponses: []*request.TaskResponse{},
	}

	switch v := requestOBJ.TaskType; v {
	case "task":
		asyncResultPtr, err := this.server.SendTaskWithContext(ctx, requestOBJ.Signatures[0])
		if err != nil {
			http.Error(w, fmt.Sprintf("Could not send task: %s", err.Error()), http.StatusExpectationFailed)
			return
		}
		// Try to obtain results,In time limit
		results, err := asyncResultPtr.GetWithTimeout(time.Duration(requestOBJ.TimeoutDuration)*time.Millisecond, time.Duration(requestOBJ.SleepDuration)*time.Millisecond)
		if err != nil && err != result.ErrTimeoutReached {
			http.Error(w, fmt.Sprintf("Task has failed with error: %s", err.Error()), http.StatusBadGateway)
			return
		}
		// Clean sensitive information
		tasks.CleanSignatureSensitiveInfo(asyncResultPtr.Signature)
		responseOBJ.TaskResponses = append(responseOBJ.TaskResponses, &request.TaskResponse{
			Results: tasks.InterfaceReadableResults(results),
			Signatures: []*tasks.Signature{
				asyncResultPtr.Signature,
			},
		})
	case "group":
		group, err := tasks.NewGroup(requestOBJ.Signatures...)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error creating group: %s", err.Error()), http.StatusBadGateway)
			return
		}

		asyncResults, err := this.server.SendGroupWithContext(ctx, group, requestOBJ.SendConcurrency)
		if err != nil {
			http.Error(w, fmt.Sprintf("Could not send group: %s", err.Error()), http.StatusBadGateway)
			return
		}
		size := len(asyncResults)
		for _, asyncResultPtr := range asyncResults {
			// Try to obtain results,In time limit
			results, err := asyncResultPtr.GetWithTimeout(time.Duration(requestOBJ.TimeoutDuration/size)*time.Millisecond, time.Duration(requestOBJ.SleepDuration)*time.Millisecond)
			if err != nil && err != result.ErrTimeoutReached {
				http.Error(w, fmt.Sprintf("Task has failed with error: %s", err.Error()), http.StatusBadGateway)
			}
			// Clean sensitive information
			tasks.CleanSignatureSensitiveInfo(asyncResultPtr.Signature)
			responseOBJ.TaskResponses = append(responseOBJ.TaskResponses, &request.TaskResponse{
				Results: tasks.InterfaceReadableResults(results),
				Signatures: []*tasks.Signature{
					asyncResultPtr.Signature,
				},
			})
		}
	case "chord":
		group, err := tasks.NewGroup(requestOBJ.Signatures...)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error creating group: %s", err.Error()), http.StatusBadGateway)
			return
		}

		chord, err := tasks.NewChord(group, requestOBJ.CallBack)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error creating chord: %s", err.Error()), http.StatusBadGateway)
			return
		}

		chordAsyncResult, err := this.server.SendChordWithContext(ctx, chord, requestOBJ.SendConcurrency)
		if err != nil {
			http.Error(w, fmt.Sprintf("Could not send chord: %s", err.Error()), http.StatusBadGateway)
			return
		}

		var signatures []*tasks.Signature
		for _, asyncResultPtr := range chordAsyncResult.GetGroupAsyncResults() {
			// Clean sensitive information
			tasks.CleanSignatureSensitiveInfo(asyncResultPtr.Signature)
			signatures = append(signatures, asyncResultPtr.Signature)
		}
		// Try to obtain results,In time limit
		results, err := chordAsyncResult.GetWithTimeout(time.Duration(requestOBJ.TimeoutDuration)*time.Millisecond, time.Duration(requestOBJ.SleepDuration)*time.Millisecond)
		if err != nil && err != result.ErrTimeoutReached {
			http.Error(w, fmt.Sprintf("Task has failed with error: %s", err.Error()), http.StatusBadGateway)
		}
		responseOBJ.TaskResponses = append(responseOBJ.TaskResponses, &request.TaskResponse{
			Results:    tasks.InterfaceReadableResults(results),
			Signatures: signatures,
			CallBack:   chordAsyncResult.GetChordAyncResults().Signature,
		})
	case "chain":
		chain, err := tasks.NewChain(requestOBJ.Signatures...)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error creating chain: %s", err), http.StatusBadGateway)
			return
		}

		chainAsyncResult, err := this.server.SendChainWithContext(ctx, chain)
		if err != nil {
			http.Error(w, fmt.Sprintf("Could not send chain: %s", err.Error()), http.StatusBadGateway)
			return
		}

		var signatures []*tasks.Signature
		for _, asyncResultPtr := range chainAsyncResult.GetAsyncResults() {
			// Clean sensitive information
			tasks.CleanSignatureSensitiveInfo(asyncResultPtr.Signature)
			signatures = append(signatures, asyncResultPtr.Signature)
		}
		// Try to obtain results,In time limit
		results, err := chainAsyncResult.GetWithTimeout(time.Duration(requestOBJ.TimeoutDuration)*time.Millisecond, time.Duration(requestOBJ.SleepDuration)*time.Millisecond)
		if err != nil && err != result.ErrTimeoutReached {
			http.Error(w, fmt.Sprintf("Task has failed with error: %s", err.Error()), http.StatusBadGateway)
		}
		responseOBJ.TaskResponses = append(responseOBJ.TaskResponses, &request.TaskResponse{
			Results:    tasks.InterfaceReadableResults(results),
			Signatures: signatures,
		})
	default:
		err := errors.New("Unexpected task type: " + v)
		span.SetTag("error", true)
		span.LogFields(
			opentracing_log.Error(err),
		)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	data, err := json.Marshal(responseOBJ)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to json.Marshal responseOBJ: %v", responseOBJ), http.StatusBadGateway)
		return
	}

	hasFinished := true
	for _, tasksResponse := range responseOBJ.TaskResponses {
		if len(tasksResponse.Results) == 0 {
			hasFinished = false
		}
	}
	if hasFinished {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusPartialContent)
	}
	w.Write([]byte(data))
}

func (this *Center) StartHttpServer() (err error) {
	var port = this.cfg.HTTP.Port
	if port == "" {
		port = ":4332"
	}
	l, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/health", this.HTTPHealth)
	mux.HandleFunc("/auth", this.HTTPAuth)
	mux.HandleFunc("/tasks/send", this.HTTPTasksSend)
	mux.HandleFunc("/tasks/touch", this.HTTPTasksTouch)
	// http.Handle("/metrics", promhttp.Handler())
	this.srv = &http.Server{Handler: mux}
	go this.srv.Serve(l)
	log.Runtime().Infof("http started on %s", port)
	return nil
}

func (this *Center) InitMetrics() (err error) {
	// if err = metrics.InitMetrics(global.Region(), config.AppTag, this.cfg.Files.Metrics, ""); err != nil {
	// 	return err
	// }
	return nil
}

func (this *Center) InitLogs() (err error) {
	if err = log.InitLog(this.cfg.Files.Log); err != nil {
		return err
	}
	return nil
}

func (this *Center) InitAuth() (err error) {
	if err = auth.Init(this.cfg.Auth); err != nil {
		return err
	}
	return nil
}

func (this *Center) Init() (err error) {
	// load config
	if err = config.AppInitConfig(); err != nil {
		log.Runtime().Errorf("config init error: %s", err.Error())
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

	// init auth
	if err = this.InitAuth(); err != nil {
		log.Runtime().Errorf("auth init error: %s", err.Error())
	}

	// Only Load Center Config
	var cfg = this.cfg.Center
	if cfg == nil {
		log.Runtime().Fatal("cfg.Center must be set")
		return
	}
	// If AMQP/MongoDB driver is used here
	if cfg.ResultBackend == "" || cfg.Broker == "" || (strings.Index(cfg.Broker, "amqp") != -1 && cfg.AMQP == nil) {
		log.Runtime().Fatal("cfg.Center.AMQP must be set")
		return
	}

	// Create server instance
	broker := amqpbroker.New(cfg)
	backend, err := mongobackend.New(cfg)
	if err != nil {
		log.Runtime().Fatalf("Unable to instantiate a mongobackend: %v", err)
		return
	}

	if err = broker.TestConnect(); err != nil {
		log.Runtime().Fatalf("Can`t build a connection to broker: %v", err)
		return
	}
	if err = backend.TestConnect(); err != nil {
		log.Runtime().Fatalf("Can`t build a connection to backend: %v", err)
		return
	}

	lock := eagerlock.New()
	this.server = NewServer(cfg, broker, backend, lock)
	return
}

func (this *Center) Run() (err error) {
	// let center run
	log.Runtime().Infof("Center Has Running")

	// Setup opentracing
	opentracingCfg := this.cfg.Opentracing
	serviceName := "aurora_center"
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

	// wait for CRTL+C to stop
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint
	return
}

func (this *Center) Stop() (err error) {
	// close http server
	err = this.srv.Close()
	return
}
