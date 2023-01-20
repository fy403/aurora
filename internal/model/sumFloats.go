package model

import "aurora/internal/request"

func init() {
	ExtantTaskMap["sum_floats"] = &request.Handler{
		Usage: "将任意个数float64累加, 返回float64, error",
		Fn:    SumFloats,
	}
}

// SumFloats ...
func SumFloats(numbers ...float64) (float64, error) {
	var sum float64
	for _, num := range numbers {
		sum += num
	}
	return sum, nil
}
