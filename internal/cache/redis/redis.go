package redis

import (
	"aurora/internal/cache/iface"
	"aurora/internal/config"
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type Redis struct {
	rclient    redis.UniversalClient
	expiration time.Duration
}

func New(cnf *config.Config) iface.Cache {
	ropt := &redis.UniversalOptions{
		Addrs:    cnf.Redis.Addr,
		DB:       cnf.Redis.DB,
		Password: cnf.Redis.Password,
	}
	if cnf.Redis != nil {
		ropt.MasterName = cnf.Redis.MasterName
	}
	return &Redis{
		rclient:    redis.NewUniversalClient(ropt),
		expiration: time.Second * 30,
	}
}

func (r *Redis) Add(key string, value interface{}) (err error) {
	ctx := context.Background()
	// redis只能存储string类型,需提前编码为string赋值给value
	v := r.rclient.Set(ctx, key, value, r.expiration)
	return v.Err()
}

func (r *Redis) Get(key string) (interface{}, error) {
	ctx := context.Background()
	v, err := r.rclient.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	return v, nil
}
func (r *Redis) Del(key string) error {
	ctx := context.Background()
	v := r.rclient.Del(ctx, key)
	return v.Err()
}

func (r *Redis) Keys(pattern string) ([]interface{}, error) {
	ctx := context.Background()
	v, err := r.rclient.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}
	var results = make([]interface{}, 0, len(v))
	for _, str := range v {
		results = append(results, str)
	}
	return results, nil
}
