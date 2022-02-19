package log

import (
	"aurora/internal/utils"
	"os"
	"path"

	"github.com/BurntSushi/toml"
	"github.com/mae-pax/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LogConfig struct {
	Runtime logger.LogOptions
	Event   logger.LogOptions
}

type Loggers struct {
	cfg     *LogConfig
	runtime *logger.Log
	event   *logger.Log
}

var defaultLoggers *Loggers

var stdoutLogger *logger.Log

func init() {
	var stdLogOption = logger.New()
	stdLogOption.SetCaller(true, 1)
	stdoutLogger = stdLogOption.InitLogger("time", "level", true, true)
}

func Stdout() *logger.Log {
	return stdoutLogger
}

func Runtime() *logger.Log {
	if defaultLoggers == nil || defaultLoggers.runtime == nil {
		return stdoutLogger
	}
	return defaultLoggers.runtime
}

func EventLog(event string, params utils.Params, value interface{}) {
	var fileds []zapcore.Field
	// if name, ok := params.GetString("StoreName"); ok {
	// 	fileds = append(fileds, zap.Any("StoreName", name))
	// }
	// if name, ok := params.GetString("GatherName"); ok {
	// 	fileds = append(fileds, zap.Any("GatherName", name))
	// }
	// if name, ok := params.GetString("JobName"); ok {
	// 	fileds = append(fileds, zap.Any("JobName", name))
	// }
	fileds = append(fileds, zap.Any("obj", value))
	Event().Info("gather informer event", fileds...)
}

func Event() *logger.Log {
	if defaultLoggers == nil || defaultLoggers.event == nil {
		return stdoutLogger
	}
	return defaultLoggers.event
}

func InitLog(file string) error {
	var cfg LogConfig
	var _, err = toml.DecodeFile(file, &cfg)
	if err != nil {
		return err
	}
	if cfg.Runtime.InfoFilename != "" {
		os.Mkdir(path.Dir(cfg.Runtime.InfoFilename), 0755)
	}
	if cfg.Runtime.ErrorFilename != "" {
		os.Mkdir(path.Dir(cfg.Runtime.ErrorFilename), 0755)
	}
	if cfg.Event.InfoFilename != "" {
		os.Mkdir(path.Dir(cfg.Event.InfoFilename), 0755)
	}
	if cfg.Event.ErrorFilename != "" {
		os.Mkdir(path.Dir(cfg.Event.ErrorFilename), 0755)
	}
	cfg.Event.CloseConsoleDisplay()
	cfg.Runtime.SetCaller(true, 1)
	defaultLoggers = &Loggers{
		cfg:     &cfg,
		runtime: cfg.Runtime.InitLogger("time", "level", true, true),
		event:   cfg.Event.InitLogger("time", "level", true, true),
	}
	return nil
}
