package example

import "strings"

// Split ...
func Split(str string) ([]string, error) {
	return strings.Split(str, ""), nil
}
