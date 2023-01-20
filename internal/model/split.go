package model

import (
	"aurora/internal/request"
	"strings"
)

func init() {
	ExtantTaskMap["split"] = &request.Handler{
		Usage: "将一个string按字母分割, 返回[]string, error",
		Fn:    Split,
	}
}

// Split ...
func Split(str string) ([]string, error) {
	return strings.Split(str, ""), nil
}
