package utils

import (
	"hash"
	"hash/fnv"
	"os"
	"path/filepath"
)

const (
	LockKeyPrefix = "aurora_lock_"
)

func GetLockName(name, spec string) string {
	return LockKeyPrefix + filepath.Base(os.Args[0]) + name + spec
}

func Hash32WithMap(src map[string]string) uint32 {
	var h32 hash.Hash32 = fnv.New32a()
	for k, v := range src {
		h32.Write([]byte(k + v))
	}
	return h32.Sum32()
}
