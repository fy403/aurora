package model

import "aurora/internal/request"

func init() {
	ExtantTaskMap["add"] = &request.Handler{
		Usage: "将任意个数int64累加, 返回int64, error",
		Fn:    Add,
	}
}

// SumInts ...
func SumInts(numbers ...int64) (int64, error) {
	var sum int64
	for _, num := range numbers {
		sum += num
	}
	return sum, nil
}
