package api

import (
	"aurora/internal/auth"
	"aurora/internal/center"
	"aurora/internal/config"
	"aurora/internal/log"
	"aurora/internal/opentracing/tracers"
	"aurora/internal/request"
	"fmt"
	"net/http"
	"strings"

	// "github.com/prometheus/client_golang/prometheus/promhttp"
	mongobackend "aurora/internal/backends/mongo"
	amqpbroker "aurora/internal/brokers/amqp"
	eagercache "aurora/internal/cache/eager"
	cachesiface "aurora/internal/cache/iface"
	rediscache "aurora/internal/cache/redis"
	eagerlock "aurora/internal/locks/eager"
	locksiface "aurora/internal/locks/iface"
	redislock "aurora/internal/locks/redis"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/yddeng/utils/task"
)

type Api struct {
	cfg       *config.AppConfig
	app       *gin.Engine
	server    *center.Server
	taskQueue *task.TaskPool
}

func NewApi() *Api {
	return &Api{}
}

func (api *Api) GetConfig() *config.AppConfig {
	return api.cfg
}
func (api *Api) GetApp() *gin.Engine {
	return api.app
}

func (api *Api) GetServer() *center.Server {
	return api.server
}

func (api *Api) GetTaskQueue() *task.TaskPool {
	return api.taskQueue
}

func (api *Api) Init() (err error) {
	// load config
	if err = config.AppInitConfig(); err != nil {
		log.Runtime().Errorf("config init error: %s", err.Error())
		return err
	}
	api.cfg = config.GetAppConfig()

	// init logs
	if err = api.initLogs(); err != nil {
		log.Runtime().Errorf("logs init error: %s", err.Error())
	}

	// init metrics
	if err = api.initMetrics(); err != nil {
		log.Runtime().Errorf("metrics init error: %s", err.Error())
	}

	// init auth
	if err = api.initAuth(); err != nil {
		log.Runtime().Errorf("auth init error: %s", err.Error())
	}

	// Only Load Gateway Config
	var cfg = api.GetConfig().Gateway
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
	api.server = center.NewServer(cfg, broker, backend, cache, lock)

	// Register faas instance
	err = api.GetServer().RegisterFaas(api.cfg.Faas)
	if err != nil {
		log.Runtime().Fatalf("RegisterFaas process error:", err)
		return
	}

	// 本地任务队列：控制连接的并发数
	api.taskQueue = task.NewTaskPool(1, 1024)
	api.app = gin.New()
	return
}

func (api *Api) Stop() (err error) {
	return
}

func (api *Api) Run() (err error) {
	app := api.GetApp()
	cfg := api.GetConfig()
	app.Use(gin.Logger(), gin.Recovery())
	// 跨域
	app.Use(func(ctx *gin.Context) {
		ctx.Header("Access-Control-Allow-Origin", "*")
		ctx.Header("Access-Control-Allow-Headers", "*")
		ctx.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE, PUT, PATCH")
		ctx.Header("Access-Control-Allow-Credentials", "true")
		ctx.Header("Access-Control-Expose-Headers", "*")
		if ctx.Request.Method == "OPTIONS" {
			// 处理浏览器的options请求时，返回200状态即可
			ctx.JSON(http.StatusOK, "")
			ctx.Abort()
			return
		}

		ctx.Next()

	})

	// 静态资源浏览
	if cfg.Gateway.Web.StaticFS {
		app.StaticFS("/static", gin.Dir(cfg.Gateway.Web.FilePath, true))
	}

	// 前端
	if cfg.Gateway.Web.WebIndex != "" {
		app.Use(static.Serve("/", static.LocalFile(cfg.Gateway.Web.WebIndex, false)))
		app.NoRoute(func(ctx *gin.Context) {
			ctx.File(cfg.Gateway.Web.WebIndex + "/index.html")
		})
	}
	// Setup opentracing
	opentracingCfg := api.GetConfig().Opentracing
	serviceName := "aurora_gateway"
	if opentracingCfg.ServiceName != "" {
		serviceName = opentracingCfg.ServiceName
	}
	cleanup, err := tracers.SetupTracer(serviceName, opentracingCfg.CollectorEndpoint, opentracingCfg.LogSpans)
	if err != nil {
		log.Runtime().Fatalf("Unable to instantiate a tracer:", err)
	}
	defer cleanup()

	api.initHandler(app)

	port := strings.Split(cfg.Gateway.Web.WebAddr, ":")[1]
	webAddr := fmt.Sprintf("0.0.0.0:%s", port)

	log.Runtime().Infof("start web service on %s", cfg.Gateway.Web.WebAddr)

	if err = app.Run(webAddr); err != nil {
		log.Runtime().Error(err.Error())
	}
	return
}

var (
	// 需要验证token的路由
	routeNeedToken = map[string]struct{}{
		"/api/task/send":   {},
		"/api/task/touch":  {},
		"/api/worker/list": {},
		"/api/faas/list":   {},
		"/api/faas/langs":  {},
		"/api/faas/create": {},
		"/api/faas/write":  {},
		"/api/faas/up":     {},
		"/api/faas/delete": {},
	}
)

func (api *Api) initHandler(app *gin.Engine) {
	authGroup := app.Group("/auth")
	apiGroup := app.Group("/api")
	app.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "ok",
			"data":    "",
		})
	})

	authHandle := new(authHandler)
	authGroup.POST("/login", api.WarpHandle(authHandle.login))
	authGroup.GET("/info", api.WarpHandle(authHandle.info))
	authGroup.DELETE("/logout", api.WarpHandle(authHandle.logout))

	taskGroup := apiGroup.Group("/task")
	taskHandle := new(taskHandler)
	taskGroup.POST("/send", api.WarpHandle(taskHandle.send))
	taskGroup.POST("/touch", api.WarpHandle(taskHandle.touch))

	workerGroup := apiGroup.Group("/worker")
	workerHandler := new(workerHandler)
	workerGroup.GET("/list", api.WarpHandle(workerHandler.list))

	faasGroup := apiGroup.Group("/faas")
	faasHandler := new(faasHandler)
	faasGroup.GET("/list", api.WarpHandle(faasHandler.ListInstance))
}

func (api *Api) checkToken(ctx *gin.Context, route string) bool {
	if _, ok := routeNeedToken[route]; !ok {
		return true
	}
	if len(ctx.Request.Header["Authorization"]) > 0 {
		ctx.Request.Header.Add("Cookie", ctx.Request.Header["Authorization"][0])
	}
	session, err := auth.DefaultStore().Get(ctx.Request, "aurora_session")

	if session.IsNew || err != nil {
		return false
	}
	return true
}

func (api *Api) initMetrics() (err error) {
	// if err = metrics.InitMetrics(global.Region(), config.AppTag, api.cfg.Files.Metrics, ""); err != nil {
	// 	return err
	// }
	return nil
}

func (api *Api) initLogs() (err error) {
	if err = log.InitLog(api.cfg.Files.Log); err != nil {
		return err
	}
	return nil
}

func (api *Api) initAuth() (err error) {
	if err = auth.Init(api.cfg.Auth); err != nil {
		return err
	}
	return nil
}

// Adjust api req selector to every signatures
func (api *Api) LabelSelector(requestOBJ *request.CenterRequest) (err error) {
	results, err := api.GetServer().GetAllWorkersInfo()
	if err != nil {
		return err
	}
	defaultLabelSelecotr := requestOBJ.LabelSelector
	// Purge invalid worker
	for idx, result := range results {
		if isValid := result.IsValid(api.cfg.Gateway.BrokerApi); !isValid {
			results[idx] = nil
			// req := request.WorkerRequest{
			// 	UUID: result.UUID,
			// }
			// api.GetServer().PurgeWorkerInfo(&req)
			continue
		}
	}
	// first match algorithm
	var found bool
	for _, sig := range requestOBJ.Signatures {
		if len(sig.LabelSelector) == 0 {
			sig.LabelSelector = defaultLabelSelecotr
		}
		for _, result := range results {
			if result == nil {
				continue
			}
			if ifMatched := result.MatchLabel(sig.LabelSelector); ifMatched {
				found = true
				sig.RoutingKey = result.SpecQueue
			}
		}
		if (len(defaultLabelSelecotr) != 0 || len(sig.LabelSelector) != 0) && !found {
			err = fmt.Errorf("Not found matched label: %s", requestOBJ.LabelSelector)
			return
		}
		found = false
	}

	return
}
