package eager

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrEagerLockFailed = errors.New("eager lock: failed to acquire lock")
)

type Lock struct {
	retries  int
	interval time.Duration
	register struct {
		sync.RWMutex
		m map[string]int64
	}
}

func New() *Lock {
	return &Lock{
		retries:  3,
		interval: 2 * time.Second,
		register: struct {
			sync.RWMutex
			m map[string]int64
		}{m: make(map[string]int64)},
	}
}

func (e *Lock) LockWithRetries(key string, expiration int64) error {
	for i := 0; i <= e.retries; i++ {
		err := e.Lock(key, expiration)
		if err == nil {
			//成功拿到锁，返回
			return nil
		}

		time.Sleep(e.interval)
	}
	return ErrEagerLockFailed
}

func (e *Lock) Lock(key string, expiration64 int64) error {
	e.register.Lock()
	defer e.register.Unlock()
	expiration := time.Duration(expiration64)
	unixTsToExpireNs := time.Now().Add(expiration).UnixNano()
	timeout, exist := e.register.m[key]
	if !exist || time.Now().UnixNano() > timeout {
		e.register.m[key] = unixTsToExpireNs
		return nil
	}
	return ErrEagerLockFailed
}

func (e *Lock) UnLock(key string) error {
	e.register.Lock()
	defer e.register.Unlock()
	_, exist := e.register.m[key]
	if exist {
		e.register.m[key] = -1
	}
	return nil
}
