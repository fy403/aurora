package model

func init() {
	ExtantTaskMap["sum_floats"] = SumFloats
}

// SumFloats ...
func SumFloats(numbers ...float64) (float64, error) {
	var sum float64
	for _, num := range numbers {
		sum += num
	}
	return sum, nil
}
