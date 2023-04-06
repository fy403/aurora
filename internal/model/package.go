package model

import (
	"aurora/internal/log"
	"aurora/internal/request"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/fsnotify/fsnotify"
)

var ExtantTaskMap map[string]*request.Handler = make(map[string]*request.Handler)

func init() {
	rootF := "./bin"
	// 遍历整个目录树，并注册句柄
	initBinary(rootF)
	// 监控bin目录，二进制句柄注册
	watchBinaryPath(rootF)
}

func initBinary(pathName string) {
	var files []string
	// 扫描pathName下的所有子文件
	err := filepath.Walk(pathName, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if !isWinExe(path) && !isLinuxExe(path) {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		log.Runtime().Error(err.Error())
		return
	}
	for _, file := range files {
		absFile, err := filepath.Abs(file)
		if err != nil {
			continue
		}
		ExtantTaskMap[filepath.Base(file)] = &request.Handler{
			Usage: "binary句柄: in[0]为string; 返回string, error",
			Fn:    createFn(absFile),
		}
	}
}

func watchBinaryPath(pathName string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Runtime().Errorf("NewWatcher failed: ", err)
		return
	}
	// defer watcher.Close()

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				log.Runtime().Debugf("watchBinaryPath: %+v", event)
				if !ok {
					log.Runtime().Errorf("watchBinaryPath Events channel closed")
					return
				}
				if event.Op != fsnotify.Remove && event.Op != fsnotify.Rename && !isFile(event.Name) {
					break
				}
				if !isWinExe(event.Name) && !isLinuxExe(event.Name) {
					break
				}
				switch event.Op {
				case fsnotify.Create:
					absFile, err := filepath.Abs(event.Name)
					if err != nil {
						continue
					}
					ExtantTaskMap[filepath.Base(event.Name)] = &request.Handler{
						Usage: "binary句柄: in[0]为string; 返回string, error",
						Fn:    createFn(absFile),
					}
				case fsnotify.Remove, fsnotify.Rename:
					delete(ExtantTaskMap, filepath.Base(event.Name))
				default:
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					log.Runtime().Errorf("watchBinaryPath Errors channel closed")
					return
				}
				log.Runtime().Errorf("NewWatcher failed: ", err)
			}
		}
	}()
	watcher.Add(pathName)
	filepath.Walk(pathName, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			err = watcher.Add(path)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func createFn(absFile string) func(string) (string, error) {
	fn := func(args string) (string, error) {
		cmd := exec.Command(absFile)
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return "", err
		}
		go func() {
			defer stdin.Close()
			io.WriteString(stdin, args)
		}()

		data, err := cmd.CombinedOutput()
		if err != nil {
			return "", err
		}
		return string(data), nil
	}
	return fn
}

func isFile(f string) bool {
	fi, e := os.Stat(f)
	if e != nil {
		return false
	}
	return !fi.IsDir()
}

func isWinExe(pathName string) bool {
	// linux,windows
	sysType := runtime.GOOS
	if sysType == "windows" && filepath.Ext(pathName) == ".exe" {
		return true
	}
	return false
}

func isLinuxExe(pathName string) bool {
	// linux,windows
	sysType := runtime.GOOS
	if sysType == "linux" && filepath.Ext(pathName) == "" {
		return true
	}
	return false
}
