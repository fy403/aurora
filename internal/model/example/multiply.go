package example

import "aurora/internal/model"

func init() {
	model.ExtantTaskMap["multiply"] = Multiply
}

// Multiply ...
func Multiply(args ...int64) (int64, error) {
	sum := int64(1)
	for _, arg := range args {
		sum *= arg
	}
	return sum, nil
}