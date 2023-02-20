package model

import (
	"aurora/internal/log"
	"aurora/internal/request"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

var ExtantTaskMap map[string]*request.Handler = make(map[string]*request.Handler)

func init() {
	pathName := "./bin"
	// 监控bin目录，二进制句柄注册
	// linux,windows
	sysType := runtime.GOOS
	var files []string
	// 扫描pathName下的所有子文件
	err := filepath.Walk(pathName, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if sysType == "windows" && filepath.Ext(path) == ".exe" {
			files = append(files, path)
		} else if sysType == "linux" && filepath.Ext(path) == "" {
			files = append(files, path)
		}
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
		ExtantTaskMap[filepath.Base(file)] = &request.Handler{
			Usage: "binary句柄: in[0]为string; 返回string, error",
			Fn:    fn,
		}
	}
}
