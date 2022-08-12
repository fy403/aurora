package example

import "aurora/internal/model"

func init() {
	model.ExtantTaskMap["add"] = Add
}

// Add ...
func Add(args ...int64) (int64, error) {
	sum := int64(0)
	for _, arg := range args {
		sum += arg
	}
	return sum, nil
}
