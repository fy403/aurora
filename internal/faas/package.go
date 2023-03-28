package faas

import (
	"aurora/internal/faas/iface"
)

var ExtantFaasMap = make(map[string]iface.Faas)
