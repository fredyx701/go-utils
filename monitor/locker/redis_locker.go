package locker

import (
	"context"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/FredyXue/go-utils"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

// 自动锁续约，直到最大超时时间 或者进程崩溃上下文丢失。
type Locker interface {
	Lock(key string) (success bool, err error)
	Unlock()                                              // 解锁当前 key
	UnlockForce(key string) (owner bool, err error)       // 强制删除 key, 可以删除其他 key
	Check(key string) (exist bool, owner bool, err error) // 判断 key 是否存在，可以查询其他 key
}

// Option
type Option func(*RedisLocker)

func WithLockTime(lockTime time.Duration) Option {
	return func(r *RedisLocker) {
		r.lockTime = lockTime
	}
}

func WithRefreshTime(refreshTime time.Duration) Option {
	return func(r *RedisLocker) {
		r.refreshTime = refreshTime
	}
}

func WithExpiredTime(expiredTime time.Duration) Option {
	return func(r *RedisLocker) {
		r.expiredTime = expiredTime
	}
}

func WithContext(ctx context.Context) Option {
	return func(r *RedisLocker) {
		r.ctx = ctx
	}
}

// RedisLocker .
type RedisLocker struct {
	ctx         context.Context // 业务 ctx
	cancelCtx   context.Context
	cancel      context.CancelFunc
	cli         *redis.Client
	lockTime    time.Duration // 加锁时长，每次续约的时长
	refreshTime time.Duration // 锁续约的周期
	expiredTime time.Duration // 最大时长
	initTime    time.Time     // 首次加锁时间点
	locked      int32         // 0 未加锁 1 加锁
	key         string
	value       string
	refreshCmd  string // expire, pexpire
	refreshDur  int64  // s/ms
}

// NewRedisLocker
// 一个 RedisLocker 对象一次只能管理一个 key
// 当 key 解锁后，可以再次管理一个新的 key
func NewRedisLocker(redisCli *redis.Client, opts ...Option) Locker {
	r := &RedisLocker{
		cli:         redisCli,
		lockTime:    time.Minute * 3,  // 默认加锁 3 分钟
		refreshTime: time.Minute,      // 默认 1 分钟续约
		expiredTime: time.Minute * 30, // 默认最大时长 30 分钟
		ctx:         context.TODO(),
	}
	for _, o := range opts {
		o(r)
	}

	// 计算刷新时间，每次续约 lockTime
	r.refreshCmd = "expire"
	r.refreshDur = int64(r.lockTime / time.Second)
	if r.lockTime < time.Second || r.lockTime%time.Second != 0 {
		r.refreshCmd = "pexpire"
		r.refreshDur = int64(r.lockTime / time.Millisecond)
	}

	return r
}

func (r *RedisLocker) run(cancelCtx context.Context) {
	ticker := time.NewTicker(r.refreshTime)
	defer ticker.Stop()

	// lock
lockerLabel:
	for {
		select {
		case now := <-ticker.C:
			// 超过最大时长，解锁
			if r.initTime.Add(r.expiredTime).Before(now) {
				r.Unlock()
				break
			}
			r.refresh() // refresh
		case <-cancelCtx.Done():
			break lockerLabel // ctx canceled 结束循环
		}
	}
}

func (r *RedisLocker) Lock(key string) (success bool, err error) {
	if r.cancelCtx != nil && r.cancelCtx.Err() == nil {
		return
	}

	// CAS 加锁，针对连续两次 lock 且 key 不同的场景
	if !atomic.CompareAndSwapInt32(&r.locked, 0, 1) {
		return
	}

	value := fmt.Sprintf("%d-%s", time.Now().Unix(), utils.RandString(10)) // 确保 value 唯一
	success, err = r.cli.SetNX(r.ctx, key, value, r.lockTime).Result()
	if err != nil {
		err = errors.Wrap(err, "RedisLocker Lock Error")
		r.locked = 0
		return
	}
	if !success {
		r.locked = 0 // 还原标记
		return
	}

	// 加锁成功
	r.value = value
	r.key = key
	r.initTime = time.Now()

	r.cancelCtx, r.cancel = context.WithCancel(context.TODO()) // 建立 cancel ctx

	go utils.Protect(func() { r.run(r.cancelCtx) }) // 开启续约，cancelCtx 作为入参，避免因为并发被替换
	return
}

// Unlock 解锁当前 key
func (r *RedisLocker) Unlock() {
	if r.cancelCtx == nil || r.cancelCtx.Err() != nil {
		return
	}

	lua := `
	if redis.call("get", KEYS[1]) == ARGV[1] then
		return redis.call("del", KEYS[1])
	else
		return 0
	end
	`
	err := r.cli.Eval(r.ctx, lua, []string{r.key}, r.value).Err()
	if err != nil {
		log.Println("RedisLocker UnLock Error", err)
	}
	r.cancel()   // 无论 redis 解锁是否成功都直接结束循环。若解锁失败，则等到锁自动过期
	r.locked = 0 // 标记解锁
	r.key = ""
}

// UnlockForce  强制删除 key, 可以删除其他 key
// 提供一个入口，为了可以使用相同的存储方式去删除 key
func (r *RedisLocker) UnlockForce(key string) (owner bool, err error) {
	// 与当前 key 不相同直接删除
	if r.key != key {
		err = r.cli.Del(r.ctx, key).Err()
		err = errors.Wrap(err, "RedisLocker UnlockForce Error")
		return
	}

	// 与当前 key 相同，比对 value
	lua := `
	if redis.call("get", KEYS[1]) == ARGV[1] then
		redis.call("del", KEYS[1])
		return 1 
	else
		redis.call("del", KEYS[1])
		return 0
	end
	`
	rlt, err := r.cli.Eval(r.ctx, lua, []string{key}, r.value).Int()
	if err != nil {
		err = errors.Wrap(err, "RedisLocker UnlockForce Error")
		return
	}
	// value 匹配，为当前 key。解锁对象
	if rlt == 1 {
		owner = true
		r.cancel()   // 结束循环
		r.locked = 0 // 标记解锁
		r.key = ""
	}
	return
}

func (r *RedisLocker) refresh() {
	if r.cancelCtx == nil || r.cancelCtx.Err() != nil {
		return
	}

	lua := fmt.Sprintf(`
	if redis.call("get", KEYS[1]) == ARGV[1] then
	  redis.call("%s", KEYS[1], ARGV[2])
	  return 1
	else
	  return 0
	end
	`, r.refreshCmd)
	rlt, err := r.cli.Eval(r.ctx, lua, []string{r.key}, r.value, r.refreshDur).Int()
	if err != nil {
		log.Println("RedisLocker refresh Error", err) // 报错继续循环
		return
	}
	if rlt == 0 {
		r.cancel()   // 续约失败直接结束循环
		r.locked = 0 // 标记解锁
		r.key = ""
	}
}

// Check 判断 key 是否存在，可以查询其他 key。
// 提供一个入口，为了可以使用相同的存储方式去查询
// owner 表示是否为当前对象的 key
func (r *RedisLocker) Check(key string) (exist bool, owner bool, err error) {
	if r.key != key {
		return
	}
	resp, err := r.cli.Get(r.ctx, r.key).Result()
	if errors.Is(err, redis.Nil) {
		return // key 不存在
	}
	if err != nil {
		err = errors.Wrap(err, "CheckLock Error")
		return
	}

	exist = true
	if r.key == key && r.value == resp { // 值相同判断 owner
		owner = true
	}
	return
}
