package model

func init() {
	ExtantTaskMap["add"] = Add
}

// SumInts ...
func SumInts(numbers ...int64) (int64, error) {
	var sum int64
	for _, num := range numbers {
		sum += num
	}
	return sum, nil
}
