package model

import "aurora/internal/request"

func init() {
	ExtantTaskMap["multiply"] = &request.Handler{
		Usage: "将任意个数int64相乘, 返回int64, error",
		Fn:    Multiply,
	}
}

// Multiply ...
func Multiply(args ...int64) (int64, error) {
	sum := int64(1)
	for _, arg := range args {
		sum *= arg
	}
	return sum, nil
}
