package model

import "aurora/internal/request"

func init() {
	ExtantTaskMap["concat"] = &request.Handler{
		Usage: "将一个[]string的元素相拼接, 返回string, error",
		Fn:    Concat,
	}
}

// Concat ...
func Concat(strs ...string) (string, error) {
	var res string
	for _, s := range strs {
		res += s
	}
	return res, nil
}
