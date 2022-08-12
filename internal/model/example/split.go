package example

import (
	"aurora/internal/model"
	"strings"
)

func init() {
	model.ExtantTaskMap["split"] = Split
}

// Split ...
func Split(str string) ([]string, error) {
	return strings.Split(str, ""), nil
}
