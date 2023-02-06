package redis

import (
	"context"
	"errors"
	"strconv"
	"time"

	"aurora/internal/config"

	"github.com/go-redis/redis/v8"
)

var (
	ErrRedisLockFailed = errors.New("redis lock: failed to acquire lock")
)

type Lock struct {
	rclient  redis.UniversalClient
	retries  int
	interval time.Duration
}

func New(cnf *config.Config) Lock {
	if cnf.Redis.Retries <= 0 {
		cnf.Redis.Retries = 3
	}
	lock := Lock{
		retries:  cnf.Redis.Retries,
		interval: 2 * time.Second,
	}

	ropt := &redis.UniversalOptions{
		Addrs:    cnf.Redis.Addr,
		DB:       cnf.Redis.DB,
		Password: cnf.Redis.Password,
	}
	if cnf.Redis != nil {
		ropt.MasterName = cnf.Redis.MasterName
	}

	lock.rclient = redis.NewUniversalClient(ropt)

	return lock
}

func (r Lock) LockWithRetries(key string, expiration int64) error {
	for i := 0; i <= r.retries; i++ {
		err := r.Lock(key, expiration)
		if err == nil {
			return nil
		}

		time.Sleep(r.interval)
	}
	return ErrRedisLockFailed
}

func (r Lock) Lock(key string, expiration64 int64) error {
	expiration := time.Duration(expiration64)
	ctx := r.rclient.Context()
	unixTsToExpireNs := time.Now().Add(expiration).UnixNano()

	success, err := r.rclient.SetNX(ctx, key, unixTsToExpireNs, expiration+1).Result()
	if err != nil {
		return err
	}
	// Has locked
	if !success {
		v, err := r.rclient.Get(ctx, key).Result()
		if err != nil {
			return err
		}
		timeout, err := strconv.Atoi(v)
		if err != nil {
			return err
		}
		now := time.Now().UnixNano()
		// lock has expired
		if timeout != 0 && now > int64(timeout) {
			// oldTimeout is old value of key
			unixTsToExpireNs = time.Now().Add(expiration).UnixNano()
			oldTimeout, err := r.rclient.GetSet(ctx, key, unixTsToExpireNs).Result()
			if err != nil {
				return err
			}

			curTimeout, err := strconv.Atoi(oldTimeout)
			if err != nil {
				return err
			}

			if now > int64(curTimeout) {
				// success to acquire lock with get set
				// set the expiration of redis key
				r.rclient.Expire(ctx, key, expiration)
				return nil
			}
			// Others acquire lock with get set faster
			return ErrRedisLockFailed
		}

		return ErrRedisLockFailed
	}

	return nil
}

func (r Lock) UnLock(key string) error {
	ctx := context.Background()
	unixTsToExpireNs := time.Now().UnixNano() - 1
	_, err := r.rclient.GetSet(ctx, key, unixTsToExpireNs).Result()
	if err != nil {
		return err
	}
	return nil
}
