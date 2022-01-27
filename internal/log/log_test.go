package log

import (
	"math/rand"
	"testing"
)

func TestRuntime(t *testing.T) {
	Runtime().Debugf("rand:%d", rand.Int())
	Runtime().Warnf("rand:%d", rand.Int())
	Runtime().Warnf("rand:%d", rand.Int())
}
