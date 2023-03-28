package redisson

import (
	"errors"
	"github.com/MaricoHan/redisson/mutex"
	"time"

	"aurora/internal/config"

	"github.com/MaricoHan/redisson"
	"github.com/go-redis/redis/v8"
)

var (
	ErrRedissonLockFailed = errors.New("redisson lock: failed to acquire lock")
)

type Lock struct {
	session  *redisson.Redisson
	lockMap  map[string]*mutex.Mutex
	retries  int
	interval time.Duration
}

func New(cnf *config.Config) Lock {
	lock := Lock{
		lockMap:  make(map[string]*mutex.Mutex),
		retries:  cnf.Redis.Retries,
		interval: 2 * time.Second,
	}

	ropt := &redis.Options{
		Addr:     cnf.Redis.Addr[0],
		DB:       cnf.Redis.DB,
		Password: cnf.Redis.Password,
	}
	lock.session = redisson.New(redis.NewClient(ropt))
	return lock
}

func (r Lock) LockWithRetries(key string, value int64) error {
	return r.Lock(key, value)
}

func (r Lock) Lock(key string, _ int64) error {
	mu := r.session.NewMutex(key)
	err := mu.Lock()
	if err != nil {
		return err
	}
	r.lockMap[key] = mu
	return nil
}

func (r Lock) UnLock(key string) error {
	mu, ok := r.lockMap[key]
	if !ok {
		return nil
	}
	err := mu.Unlock()
	delete(r.lockMap, key)
	return err
}
