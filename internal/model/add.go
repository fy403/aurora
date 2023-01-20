package model

import "aurora/internal/request"

func init() {
	ExtantTaskMap["add"] = &request.Handler{
		Usage: "将任意个数int64相加, 返回int64, error",
		Fn:    Add,
	}
}

// Add ...
func Add(args ...int64) (int64, error) {
	sum := int64(0)
	for _, arg := range args {
		sum += arg
	}
	return sum, nil
}
