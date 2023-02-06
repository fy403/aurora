package redis

import (
	"aurora/internal/cache/iface"
	"aurora/internal/config"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var cnf = &config.Config{
	Redis: &config.RedisConfig{
		DB:       0,
		Addr:     []string{":6379"},
		Password: "123456",
		Retries:  3,
	},
}

func TestNew(t *testing.T) {
	redis := New(cnf)
	assert.Implements(t, (*iface.Cache)(nil), redis)
}

func TestGet(t *testing.T) {
	redis := New(cnf)
	val, err := redis.Get("Redis cache")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Redis cache: %s", val.(string))
}

func TestAdd(t *testing.T) {
	redis := New(cnf)
	err := redis.Add("Redis cache", "ok")
	if err != nil {
		t.Fatal(err)
	}
}
func TestAddObject(t *testing.T) {
	redis := New(cnf)
	obj := struct {
		Name string `json:"name"`
	}{Name: "ok"}
	b, err := json.Marshal(obj)
	if err != nil {
		t.Fatal(err)
	}
	err = redis.Add("Redis cache", b)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetObject(t *testing.T) {
	redis := New(cnf)
	val, err := redis.Get("Redis cache")
	if err != nil {
		t.Fatal(err)
	}
	obj := struct {
		Name string `json:"name"`
	}{}
	json.Unmarshal([]byte(val.(string)), &obj)
	t.Logf("Redis cache: %#v", obj)
}

func TestDel(t *testing.T) {
	redis := New(cnf)
	err := redis.Del("Redis cache")
	if err != nil {
		t.Fatal(err)
	}
}

func TestKeys(t *testing.T) {
	redis := New(cnf)
	for i := 0; i < 20; i++ {
		redis.Add(fmt.Sprintf("Redis:%d", i), i)
	}
	keys, err := redis.Keys("Redis:*")
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range keys {
		val, err := redis.Get(key.(string))
		if err != nil {
			continue
		}
		t.Logf("%s %s\t", key, val)
	}
}
