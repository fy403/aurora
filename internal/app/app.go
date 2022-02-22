package app

import (
	"context"
	"net"
	"net/http"
	_ "net/http/pprof"
	"time"

	// "github.com/prometheus/client_golang/prometheus/promhttp"

	"aurora/internal/config"
	"aurora/internal/log"
)

type App struct {
	cfg    *config.AppConfig
	cancel context.CancelFunc
}

func NewApp() *App {
	return &App{}
}

func (this *App) HTTPHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (this *App) StartHttpServer() (err error) {
	var port = this.cfg.HTTP.Port
	if port == "" {
		port = ":80"
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

func (this *App) InitMetrics() (err error) {
	// if err = metrics.InitMetrics(global.Region(), config.AppTag, this.cfg.Files.Metrics, ""); err != nil {
	// 	return err
	// }
	return nil
}

func (this *App) InitLogs() (err error) {
	if err = log.InitLog(this.cfg.Files.Log); err != nil {
		return err
	}
	return nil
}

func (this *App) Init() (err error) {
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
		return
	}

	// TODO: link config

	// create instance
	_, cancel := context.WithCancel(context.Background())
	// TODO: load app config
	this.cancel = cancel
	return
}

func (this *App) Run() (err error) {
	// let health check run
	if err = this.StartHttpServer(); err != nil {
		return
	}
	// TODO: let app run
	time.Local, _ = time.LoadLocation("Asia/Beijing")
	log.Runtime().Infof("center has running")
	return
}

func (this *App) Stop() (err error) {
	this.cancel()
	return
}
