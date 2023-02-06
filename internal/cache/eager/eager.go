package eager

import (
	"aurora/internal/cache/iface"
	"fmt"
	"sync"
)

type Cache struct {
	sync.RWMutex
	m map[string]interface{}
}

func New() iface.Cache {
	return &Cache{
		m: make(map[string]interface{}),
	}
}

func (c *Cache) Add(key string, value interface{}) error {
	c.Lock()
	defer c.Unlock()
	c.m[key] = value
	return nil
}

func (c *Cache) Get(key string) (interface{}, error) {
	c.Lock()
	defer c.Unlock()
	if v, ok := c.m[key]; !ok {
		return nil, fmt.Errorf("Not such key: ", key)
	} else {
		return v, nil
	}
}
func (c *Cache) Del(key string) error {
	c.Lock()
	defer c.Unlock()
	delete(c.m, key)
	return nil
}

func (c *Cache) Keys(pattern string) ([]interface{}, error) {
	c.Lock()
	defer c.Unlock()
	var results = make([]interface{}, 0, len(c.m))
	for _, str := range c.m {
		results = append(results, str)
	}
	return results, nil
}
