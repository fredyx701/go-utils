package locker

import (
	"context"
	"errors"
	"log"
	"testing"
	"time"

	"github.com/FredyXue/go-utils/testdata"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

func TestRedisLocker(t *testing.T) {
	redisClient := testdata.NewTestRedis()
	key := "test_redis_locker"
	key2 := "test_redis_locker2"
	ctx := context.TODO()

	// test for refresh
	locker := NewRedisLocker(
		redisClient,
		WithLockTime(time.Millisecond*30),   // 30ms
		WithRefreshTime(time.Millisecond*5), // 5ms
		WithExpiredTime(time.Minute*30),
	)

	success, err := locker.Lock(key)
	assert.NoError(t, err)
	assert.Equal(t, success, true)
	exist, owner, _ := locker.Check(key)
	val, err := redisClient.Get(ctx, key).Result()
	assert.NoError(t, err)
	log.Println("get val:", val)
	assert.Equal(t, []any{exist, owner, val != ""}, []any{true, true, true})

	time.Sleep(time.Millisecond * 100) // 休眠 100ms, 验证 key 还存活
	exist, owner, _ = locker.Check(key)
	pttl, _ := redisClient.PTTL(ctx, key).Result()
	assert.Equal(t, []any{exist, owner, pttl > time.Millisecond*20}, []any{true, true, true})

	locker.Unlock()
	exist, owner, _ = locker.Check(key)
	val, err = redisClient.Get(ctx, key).Result()
	assert.Equal(t, errors.Is(err, redis.Nil), true)
	assert.Equal(t, []any{exist, owner, val}, []any{false, false, ""})

	// test for expired
	locker = NewRedisLocker(
		redisClient,
		WithLockTime(time.Millisecond*30),    // 30ms
		WithRefreshTime(time.Millisecond*5),  // 5ms
		WithExpiredTime(time.Millisecond*90), // 110ms
	)

	locker.Lock(key)
	exist, owner, _ = locker.Check(key)
	val, _ = redisClient.Get(ctx, key).Result()
	assert.Equal(t, []any{exist, owner, val != ""}, []any{true, true, true})

	time.Sleep(time.Millisecond * 60)
	exist, owner, _ = locker.Check(key)
	val, _ = redisClient.Get(ctx, key).Result()
	assert.Equal(t, []any{exist, owner, val != ""}, []any{true, true, true})

	time.Sleep(time.Millisecond * 60) // 超时
	exist, owner, _ = locker.Check(key)
	val, err = redisClient.Get(ctx, key).Result()
	assert.Equal(t, errors.Is(err, redis.Nil), true)
	assert.Equal(t, []any{exist, owner, val}, []any{false, false, ""})

	// test for unlock force
	redisClient.Set(ctx, key2, "value", 0)
	val, _ = redisClient.Get(ctx, key2).Result()
	assert.Equal(t, val, "value")

	locker.UnlockForce(key2) // 删除其他 key
	val, err = redisClient.Get(ctx, key2).Result()
	assert.Equal(t, []any{val, errors.Is(err, redis.Nil)}, []any{"", true})

	locker.Lock(key)
	exist, owner, _ = locker.Check(key)
	val, _ = redisClient.Get(ctx, key).Result()
	assert.Equal(t, []any{exist, owner, val != ""}, []any{true, true, true})

	locker.UnlockForce(key) // 删除当前 key
	exist, owner, _ = locker.Check(key)
	val, err = redisClient.Get(ctx, key).Result()
	assert.Equal(t, errors.Is(err, redis.Nil), true)
	assert.Equal(t, []any{exist, owner, val}, []any{false, false, ""})
}
