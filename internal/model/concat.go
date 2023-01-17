package model

func init() {
	ExtantTaskMap["concat"] = Concat
}

// Concat ...
func Concat(strs []string) (string, error) {
	var res string
	for _, s := range strs {
		res += s
	}
	return res, nil
}
