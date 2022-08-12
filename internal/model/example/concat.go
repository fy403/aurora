package example

import "aurora/internal/model"

func init() {
	model.ExtantTaskMap["concat"] = Concat
}

// Concat ...
func Concat(strs []string) (string, error) {
	var res string
	for _, s := range strs {
		res += s
	}
	return res, nil
}
