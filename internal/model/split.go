package model

import (
	"strings"
)

func init() {
	ExtantTaskMap["split"] = Split
}

// Split ...
func Split(str string) ([]string, error) {
	return strings.Split(str, ""), nil
}
