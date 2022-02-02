package util

import (
	"strings"
	"time"
)

type Params map[string]interface{}

func (this Params) GetDuration(key string) (value time.Duration, ok bool) {
	key = strings.ToLower(key)
	v, ok := this.GetString(key)
	if ok {
		var err error
		value, err = time.ParseDuration(v)
		if err != nil {
			ok = false
			// log.Runtime().Errorf("parse duration faild: %s", v)
		}
	}
	return value, ok
}

func (this Params) GetDurationOrDefault(key, defaultValue string) (value time.Duration) {
	key = strings.ToLower(key)
	var ok bool
	if value, ok = this.GetDuration(key); !ok {
		var err error
		value, err = time.ParseDuration(defaultValue)
		if err != nil {
			ok = false
		}
	}
	return
}

func (this Params) GetStringOrDefault(key, defaultValue string) (value string) {
	key = strings.ToLower(key)
	var ok bool
	if value, ok = this.GetString(key); !ok {
		value = defaultValue
	}
	return
}

func (this Params) Merge(p Params) Params {
	for k, v := range p {
		lowK := strings.ToLower(k)
		if _, ok := this[lowK]; !ok {
			this[lowK] = v
		}
	}
	return this
}

func (this Params) GetString(key string) (value string, ok bool) {
	key = strings.ToLower(key)
	intV, ok := this[key]
	if ok {
		value, ok = intV.(string)
	}
	return
}

func (this Params) GetInt(key string) (value int, ok bool) {
	key = strings.ToLower(key)
	intV, ok := this[key]
	if ok {
		switch v := intV.(type) {
		case int:
			return v, ok
		case int64:
			return int(v), ok
		case uint64:
			return int(v), ok
		case int32:
			return int(v), ok
		case uint32:
			return int(v), ok
		}
	}
	return
}

func (this Params) GetIntOrDefault(key string, defaultValue int) (value int) {
	key = strings.ToLower(key)
	var ok bool
	if value, ok = this.GetInt(key); !ok {
		value = defaultValue
	}
	return
}
