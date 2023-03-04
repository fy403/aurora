package redisson

import (
	"errors"
	"time"

	"aurora/internal/config"

	"github.com/MaricoHan/redisson"
	"github.com/go-redis/redis/v8"
)

var (
	ErrRedissonLockFailed = errors.New("redisson lock: failed to acquire lock")
)

type Lock struct {
	rclient        *redis.Client
	lockSessionMap map[string]*redisson.Redisson
	retries        int
	interval       time.Duration
}

func New(cnf *config.Config) Lock {
	lock := Lock{
		retries:  cnf.Redis.Retries,
		interval: 2 * time.Second,
	}

	ropt := &redis.Options{
		Addr:     cnf.Redis.Addr[0],
		DB:       cnf.Redis.DB,
		Password: cnf.Redis.Password,
	}

	lock.rclient = redis.NewClient(ropt)

	return lock
}

func (r Lock) LockWithRetries(key string, clientId int64) error {
	for hasTries := 0; hasTries != r.retries; hasTries++ {
		err := r.Lock(key, clientId)
		if err == nil {
			return nil
		}
		time.Sleep(r.interval)
	}
	return ErrRedissonLockFailed
}

func (r Lock) Lock(key string, clientId int64) error {
	name := key + string(clientId)
	session := redisson.New(r.rclient)
	r.lockSessionMap[name] = session
	session.NewMutex(name)
	err := mutex.Lock()
	if err != nil {
		return err
	}
	return nil
}

func (r Lock) UnLock(key string) error {
	err = mutex.Unlock()
	if err != nil {
		return err
	}

	return nil
}
