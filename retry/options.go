package retry

import (
	"math"
	"time"

	"github.com/FredyXue/go-utils"
)

// CheckFunc note that returning either false or a non-nil error will result in the call not being retried
type CheckFunc func(retryCount int, err error) (bool, error)

// defaultCheckFunc .
func defaultCheckFunc(retryCount int, err error) (bool, error) {
	return true, nil
}

// BackoffFunc backoff function
// attempt start with 0
type BackoffFunc func(attempt int, interval time.Duration) (time.Duration, error)

// ExponentialBackoff   10^(n-1)*interval   指数间隔
// attempt start with 0
func ExponentialBackoff(attempt int, interval time.Duration) (time.Duration, error) {
	if attempt == 0 {
		return time.Duration(0), nil
	}
	return time.Duration(math.Pow(10, float64(attempt-1))) * interval, nil
}

// FibonacciBackoff  An = An-1 + An-2   斐波那契数列间隔
func FibonacciBackoff(attempt int, interval time.Duration) (time.Duration, error) {
	if attempt == 0 {
		return time.Duration(0), nil
	}
	return time.Duration(utils.Fibonacci(attempt)) * interval, nil
}

// AverageBackOff .
func AverageBackOff(attempt int, interval time.Duration) (time.Duration, error) {
	if attempt == 0 {
		return time.Duration(0), nil
	}
	return interval, nil
}

// IncreaseBackOff  n * interval
func IncreaseBackOff(attempt int, interval time.Duration) (time.Duration, error) {
	if attempt == 0 {
		return time.Duration(0), nil
	}
	return time.Duration(attempt) * interval, nil
}
