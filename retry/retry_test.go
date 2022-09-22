package retry

import (
	"errors"
	"log"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRetry(t *testing.T) {
	now := time.Now()
	arr := []int64{}

	// fibonacci
	NewRetry(
		WithRetry(5),
		WithInterval(time.Millisecond*10),
	).Do(func() error {
		now2 := time.Now()
		delta := (now2.UnixNano() - now.UnixNano()) / 1e6
		log.Printf("fibonacci ms %v", delta)
		now = now2
		arr = append(arr, delta/10)
		return errors.New("testerror")
	})
	assert.EqualValues(t, arr, []int64{0, 1, 2, 3, 5, 8})

	now = time.Now()
	arr = []int64{}
	count := 0

	// exponent
	NewRetry(
		WithInterval(time.Millisecond),
		WithBackoff(ExponentialBackoff),
	).Do(func() error {
		delta := (time.Now().UnixNano() - now.UnixNano()) / 1e6
		log.Printf("exponent ms %v", delta)
		arr = append(arr, int64(float64(delta)/math.Pow(10, float64(count)-1))) // 0/0.1, 1+/1,  10+/10, 100+/100
		count++
		return errors.New("testerror")
	})
	assert.EqualValues(t, arr, []int64{0, 1, 1, 1})

	now = time.Now()
	arr = []int64{}

	// average
	NewRetry(
		WithInterval(time.Millisecond*10),
		WithBackoff(AverageBackOff),
	).Do(func() error {
		now2 := time.Now()
		delta := (now2.UnixNano() - now.UnixNano()) / 1e6
		log.Printf("average ms %v", delta)
		now = now2
		arr = append(arr, delta/10)
		return errors.New("testerror")
	})
	assert.EqualValues(t, arr, []int64{0, 1, 1, 1})

	now = time.Now()
	arr = []int64{}

	// increase
	err := NewRetry(
		WithInterval(time.Millisecond*10),
		WithBackoff(IncreaseBackOff),
	).Do(func() error {
		now2 := time.Now()
		delta := (now2.UnixNano() - now.UnixNano()) / 1e6
		log.Printf("increase ms %v", delta)
		now = now2
		arr = append(arr, delta/10)
		return errors.New("testerror")
	})
	log.Println(err)
	assert.EqualValues(t, arr, []int64{0, 1, 2, 3})

	now = time.Now()
	arr = []int64{}

	// check
	check := func(retryCount int, err error) (bool, error) {
		if retryCount == 1 {
			return false, nil
		}
		return true, nil
	}
	NewRetry(
		WithCheck(check),
		WithInterval(time.Millisecond*10),
		WithBackoff(AverageBackOff),
	).Do(func() error {
		now2 := time.Now()
		delta := (now2.UnixNano() - now.UnixNano()) / 1e6
		log.Printf("check ms %v", delta)
		now = now2
		arr = append(arr, delta/10)
		return errors.New("testerror")
	})
	assert.EqualValues(t, arr, []int64{0, 1})

	now = time.Now()
	arr = []int64{}

	// max interval
	err = NewRetry(
		WithRetry(5),
		WithMaxInterval(time.Millisecond*30),
		WithInterval(time.Millisecond*10),
		WithBackoff(IncreaseBackOff),
	).Do(func() error {
		now2 := time.Now()
		delta := (now2.UnixNano() - now.UnixNano()) / 1e6
		log.Printf("increase ms %v", delta)
		now = now2
		arr = append(arr, delta/10)
		return errors.New("testerror")
	})
	log.Println(err)
	assert.EqualValues(t, arr, []int64{0, 1, 2, 3, 3, 3}) // 最后 2 次，触发 maxInterval
}

func TestRetryCheck(t *testing.T) {
	// success
	count := 0
	success, err := NewPolling(
		WithRetry(10),
		WithInterval(time.Millisecond*10),
	).Polling(func() (bool, error) {
		count++
		if count == 3 {
			return true, nil
		}
		return false, nil
	})
	assert.Equal(t, success, true)
	assert.NoError(t, err)

	// timeout
	count = 0
	success, err = NewPolling(
		WithRetry(10),
		WithInterval(time.Millisecond*10),
	).Polling(func() (bool, error) {
		count++
		return false, nil
	})
	assert.Equal(t, success, false)
	assert.Equal(t, count, 11)
	assert.Equal(t, err != nil, true)
	log.Println("get timeout error: ", err)

	// func error
	count = 0
	success, err = NewPolling(
		WithRetry(2),
		WithInterval(time.Millisecond*10),
	).Polling(func() (bool, error) {
		count++
		return false, errors.New("testerror")
	})
	assert.Equal(t, success, false)
	assert.Equal(t, count, 1) // func error 一次发生即不再重试
	assert.Equal(t, err != nil, true)
	log.Println("get max failed error: ", err)

	// check
	count = 0
	check := func(retryCount int, err error) (bool, error) {
		if retryCount == 2 {
			return false, nil
		}
		return true, nil
	}
	success, err = NewPolling(
		WithCheck(check),
		WithRetry(10),
		WithInterval(time.Millisecond*10),
	).Polling(func() (bool, error) {
		count++
		return false, errors.New("testerror")
	})
	assert.Equal(t, success, false)
	assert.Equal(t, count, 3) // 执行 3 次
	assert.Equal(t, err != nil, true)

	// delay
	count = 0
	backoff := func(attempt int, interval time.Duration) (time.Duration, error) {
		count += attempt
		if attempt == 0 {
			return time.Duration(0), nil
		}
		return interval, nil
	}
	_, err = NewPolling(
		WithInterval(time.Millisecond*10),
		WithBackoff(backoff),
		WithRetry(3), // 默认立即执行
	).Polling(func() (bool, error) {
		return false, nil
	})
	assert.Equal(t, success, false)
	assert.Equal(t, count, 6) // index 从 0 开始, 0+1+2+3 = 6
	assert.Equal(t, err != nil, true)

	count = 0
	_, err = NewPolling(
		WithInterval(time.Millisecond*10),
		WithBackoff(backoff),
		WithRetry(3),
		WithDelay(true), // 一个周期后执行
	).Polling(func() (bool, error) {
		return false, nil
	})
	assert.Equal(t, success, false)
	assert.Equal(t, count, 10) // index 从 1 开始, 1+2+3+4 = 10
	assert.Equal(t, err != nil, true)
}
