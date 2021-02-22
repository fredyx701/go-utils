package retry

import (
	"time"

	"github.com/pkg/errors"
)

// Option .
type Option func(*Retry)

// WithBackoff is used to set the backoff function used when retrying Calls
// interval  base time interval
func WithBackoff(fn BackoffFunc, interval time.Duration) Option {
	return func(o *Retry) {
		o.backoff = fn
		o.interval = interval
	}
}

// WithRetry Number of retries when making the request.
// Should this be a Call Option?
func WithRetry(i int) Option {
	return func(o *Retry) {
		o.retries = i
	}
}

// WithCheck sets the retry function to be used when re-trying.
func WithCheck(fn CheckFunc) Option {
	return func(o *Retry) {
		o.check = fn
	}
}

// WithInterval  base time interval  与  Backoff 中的 interval 冲突
func WithInterval(interval time.Duration) Option {
	return func(o *Retry) {
		o.interval = interval
	}
}

// Retry .
type Retry struct {
	retries  int
	check    CheckFunc
	backoff  BackoffFunc
	interval time.Duration
}

// NewRetry .
func NewRetry(opts ...Option) *Retry {
	r := &Retry{
		retries:  3,
		check:    defaultCheckFunc,
		backoff:  FibonacciBackoff,      // 默认斐波那契数列间隔
		interval: time.Millisecond * 10, // 默认 10ms 间隔
	}
	for _, o := range opts {
		o(r)
	}
	return r
}

// Do exec fn
func (r *Retry) Do(fn func() error) error {
	var gerr error
	for i := 0; i <= r.retries; i++ {
		// call backoff first. Someone may want an initial start delay
		t, berr := r.backoff(i, r.interval)
		if berr != nil {
			return errors.Wrap(berr, "retry backoff error")
		}

		// 0 duration not sleep
		if t.Seconds() > 0 {
			time.Sleep(t)
		}

		// exec
		ferr := fn()
		if ferr == nil {
			return nil
		}

		// cehck retry
		retry, cerr := r.check(i, ferr)
		if cerr != nil {
			return errors.Wrap(cerr, "retry check error")
		}
		if !retry {
			return ferr
		}

		// merge error
		if gerr == nil {
			gerr = ferr
		} else {
			gerr = errors.Wrap(gerr, ferr.Error())
		}
	}

	return gerr
}
