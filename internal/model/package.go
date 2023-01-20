package model

import (
	"aurora/internal/request"
)

var ExtantTaskMap map[string]*request.Handler = make(map[string]*request.Handler)
