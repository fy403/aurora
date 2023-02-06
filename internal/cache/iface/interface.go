package iface

type Cache interface {
	// String
	Add(key string, value interface{}) error
	Get(key string) (interface{}, error)
	Del(key string) error
	Keys(pattern string) ([]interface{}, error)
}
