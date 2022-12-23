package retry

import (
	"log"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

// Option .
type Option func(*retry)

// WithBackoff is used to set the backoff function used when retrying Calls
func WithBackoff(fn BackoffFunc) Option {
	return func(o *retry) {
		o.backoff = fn
	}
}

// WithRetry Number of retries when making the request.
// Should this be a Call Option?
func WithRetry(i int) Option {
	return func(o *retry) {
		o.retries = i
	}
}

// WithCheck sets the retry function to be used when re-trying.
func WithCheck(fn CheckFunc) Option {
	return func(o *retry) {
		o.check = fn
	}
}

// WithInterval  base time interval
func WithInterval(interval time.Duration) Option {
	return func(o *retry) {
		o.interval = interval
	}
}

// WithMaxInterval  max time interval
func WithMaxInterval(maxInterval time.Duration) Option {
	return func(o *retry) {
		o.maxInterval = maxInterval
	}
}

// WithExpiredDuration  expired duration
func WithExpiredDuration(expiredDuration time.Duration) Option {
	return func(o *retry) {
		o.expiredDuration = expiredDuration
	}
}

// WithLogMode 是否输出错误日志
func WithLogMode(logMode bool) Option {
	return func(o *retry) {
		o.logMode = logMode
	}
}

// WithDelay 是否延迟执行
func WithDelay(delay bool) Option {
	return func(o *retry) {
		if delay {
			o.delay = 1
		} else {
			o.delay = 0
		}
	}
}

type Retry interface {
	Do(fn func() error) error
}

type Polling interface {
	Polling(fn func() (bool, error)) (bool, error)
}

// Retry .
type retry struct {
	retries         int           // 重试次数，不包含首次执行。 总执行次数 = 1 + retry
	check           CheckFunc     // 错误检查策略
	backoff         BackoffFunc   // 重试间隔策略
	interval        time.Duration // base interval
	maxInterval     time.Duration // 最大重试间隔
	expiredDuration time.Duration // 最大重试持续时间
	delay           int           // 是否延迟执行;  0 立即执行  1 一个周期后执行.   默认立即执行
	logMode         bool          // 是否输出错误日志
}

// NewRetry .
func NewRetry(opts ...Option) Retry {
	r := &retry{
		retries:         3,
		check:           defaultCheckFunc,
		backoff:         FibonacciBackoff,      // 默认斐波那契数列间隔
		interval:        time.Millisecond * 10, // 默认 10ms 间隔
		expiredDuration: 0,                     // 默认不设置最大重试持续时间
		delay:           0,
		logMode:         true, // 默认输出错误日志
	}
	for _, o := range opts {
		o(r)
	}
	return r
}

// Do exec fn
func (r *retry) Do(fn func() error) error {
	var gerr error
	startAt := time.Now()
	for i := 0; i <= r.retries; i++ {
		// call backoff first. Someone may want an initial start delay
		t, berr := r.backoff(i+r.delay, r.interval)
		if berr != nil {
			return errors.Wrap(berr, "retry backoff error")
		}

		// 0 duration not sleep
		if t > 0 {
			if r.maxInterval > 0 && r.maxInterval < t { // 判断最大超时时间
				time.Sleep(r.maxInterval)
			} else {
				time.Sleep(t)
			}
		}

		// exec
		ferr := fn()
		if ferr == nil {
			return nil // no error, continue
		}

		// check func retry
		retry, cerr := r.check(i, ferr)
		if cerr != nil {
			return errors.Wrap(cerr, "retry check func error")
		}
		if !retry {
			return ferr // reject retry and return func error
		}

		// 判断最大重试持续时间
		// 超时则退出     start_at + expired_duration < now
		if r.expiredDuration > 0 && startAt.Add(r.expiredDuration).Before(time.Now()) {
			gerr = multierror.Append(gerr, ferr)
			break
		}

		if r.logMode {
			log.Printf("exec retry do with error: %v", ferr)
		}

		// merge func error. 最多包装 30 个 error
		if i <= 30 {
			gerr = multierror.Append(gerr, ferr)
		}
	}

	gerr = multierror.Append(gerr, errors.New("retry timeout"))

	return gerr
}

// NewPolling .
// 与 NewRetry 默认参数不同
func NewPolling(opts ...Option) Polling {
	r := &retry{
		retries:         60, // 默认 60 次轮询, 60 * 10 = 10 分钟
		check:           defaultPollingCheckFunc,
		backoff:         AverageBackOff,   // 平均间隔
		interval:        time.Second * 10, // 默认 10 s 间隔
		expiredDuration: 0,                // 默认不设置最大重试持续时间
		delay:           0,
		logMode:         true, // 默认输出错误日志
	}
	for _, o := range opts {
		o(r)
	}
	return r
}

// 轮询 check 的场景
func (r *retry) Polling(fn func() (bool, error)) (bool, error) {
	var gerr error
	startAt := time.Now()
	for i := 0; i <= r.retries; i++ {
		// call backoff first. Someone may want an initial start delay
		t, berr := r.backoff(i+r.delay, r.interval)
		if berr != nil {
			return false, errors.Wrap(berr, "retry backoff error")
		}

		// 0 duration not sleep
		if t > 0 {
			if r.maxInterval > 0 && r.maxInterval < t { // 判断最大超时时间
				time.Sleep(r.maxInterval)
			} else {
				time.Sleep(t)
			}
		}

		// exec
		success, ferr := fn()
		if success {
			return success, ferr
		}
		if ferr == nil {
			continue // no error, continue
		}

		// check func error
		retry, cerr := r.check(i, ferr)
		if cerr != nil {
			return false, errors.Wrap(cerr, "retry check func error")
		}
		if !retry {
			return false, ferr // reject retry and return func error
		}

		// 判断最大重试持续时间
		// 超时则退出     start_at + expired_duration < now
		if r.expiredDuration > 0 && startAt.Add(r.expiredDuration).Before(time.Now()) {
			gerr = multierror.Append(gerr, ferr)
			break
		}

		if r.logMode {
			log.Printf("exec retry polling with error: %v", ferr)
		}

		// merge func error. 最多包装 30 个 error
		if i <= 30 {
			gerr = multierror.Append(gerr, ferr)
		}
	}

	gerr = multierror.Append(gerr, errors.New("retry timeout"))

	return false, gerr
}
