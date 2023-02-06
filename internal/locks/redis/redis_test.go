package redis

import (
	"aurora/internal/config"
	lockiface "aurora/internal/locks/iface"
	"testing"
	"time"

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

func TestLock_Lock(t *testing.T) {

	lock := New(cnf)
	keyName := "Distributed lock"

	go func() {
		err := lock.Lock(keyName, int64(10*time.Second))
		assert.NoError(t, err)
	}()
	time.Sleep(1 * time.Second)
	err := lock.Lock(keyName, int64(4*time.Second))
	assert.Error(t, err)
	assert.EqualError(t, err, ErrRedisLockFailed.Error())
}

func TestLock_LockWithRetries(t *testing.T) {
	lock := New(cnf)
	keyName := "Distributed lock"
	expiration := 20 * time.Second
	go func() {
		lockedTime := time.Now().UnixNano()
		defer func() {
			unLockedTime := time.Now().UnixNano()
			// 当前客户但持有锁
			if unLockedTime-lockedTime <= int64(expiration) {
				lock.UnLock(keyName)
			}
		}()
		err := lock.LockWithRetries(keyName, int64(expiration))
		assert.NoError(t, err)
	}()
	time.Sleep(1 * time.Second)
	err := lock.LockWithRetries(keyName, int64(expiration))
	assert.NoError(t, err)
}

func TestNew(t *testing.T) {
	lock := New(cnf)
	assert.Implements(t, (*lockiface.Lock)(nil), lock)
}
