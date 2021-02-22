package retry

import (
	"time"

	"github.com/pkg/errors"
)

// Options .
type Options struct {
	Retries  int           // Number of retries when making the request.
	Check    CheckFunc     // set retry function to be used when re-trying.
	BackOff  BackoffFunc   // set the backoff function used when retrying Calls
	Interval time.Duration // base time interval
}

// Retry .
type Retry struct {
	retries  int
	check    CheckFunc
	backoff  BackoffFunc
	interval time.Duration
}

// NewRetry .
func NewRetry(opts ...Options) *Retry {
	r := &Retry{
		retries:  3,
		check:    defaultCheckFunc,
		backoff:  FibonacciBackoff,      // 默认斐波那契数列间隔
		interval: time.Millisecond * 10, // 默认 10ms 间隔
	}
	if len(opts) > 0 {
		o := opts[0]
		if o.Retries > 0 {
			r.retries = o.Retries
		}
		if o.Check != nil {
			r.check = o.Check
		}
		if o.BackOff != nil {
			r.backoff = o.BackOff
		}
		if o.Interval > 0 {
			r.interval = o.Interval
		}
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
