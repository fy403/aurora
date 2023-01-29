package utils

import (
	"hash"
	"hash/fnv"
	"os"
	"path/filepath"
	"strings"
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

func ExtractParams(output string) map[string]string {
	strs := strings.Split(output, "\n")
	for idx := 0; idx < len(strs); idx++ {
		strs[idx] = strings.Replace(strs[idx], " ", "", -1)
		strs[idx] = strings.Replace(strs[idx], "\n", "", -1)
		strs[idx] = strings.Replace(strs[idx], "\t", "", -1)
	}
	params := map[string]string{}
	for _, str := range strs {
		splitVal := strings.SplitN(str, ":", 2)
		if len(splitVal) >= 2 {
			key := splitVal[0]
			val := splitVal[1]
			params[key] = val
		}
	}
	return params
}
