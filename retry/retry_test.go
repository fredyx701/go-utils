package retry

import (
	"errors"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRetry(t *testing.T) {
	now := time.Now()

	// fibonacci
	NewRetry(
		WithRetry(5),
		WithInterval(time.Millisecond*10),
	).Do(func() error {
		now2 := time.Now()
		log.Printf("fibonacci ms %v", (now2.UnixNano()-now.UnixNano())/1e6)
		now = now2
		return errors.New("testerror")
	})

	now = time.Now()

	// exponent
	NewRetry(
		WithBackoff(ExponentialBackoff, time.Millisecond),
	).Do(func() error {
		log.Printf("exponent ms %v", (time.Now().UnixNano()-now.UnixNano())/1e6)
		return errors.New("testerror")
	})

	now = time.Now()

	// average
	NewRetry(
		WithBackoff(AverageBackOff, time.Millisecond*10),
	).Do(func() error {
		now2 := time.Now()
		log.Printf("average ms %v", (time.Now().UnixNano()-now.UnixNano())/1e6)
		now = now2
		return errors.New("testerror")
	})

	// increase
	err := NewRetry(
		WithBackoff(IncreaseBackOff, time.Millisecond*10),
	).Do(func() error {
		now2 := time.Now()
		log.Printf("increase ms %v", (time.Now().UnixNano()-now.UnixNano())/1e6)
		now = now2
		return errors.New("testerror")
	})
	log.Println(err)

	// check
	check := func(retryCount int, err error) (bool, error) {
		if retryCount == 1 {
			return false, nil
		}
		return true, nil
	}
	NewRetry(
		WithCheck(check),
		WithBackoff(AverageBackOff, time.Millisecond*10),
	).Do(func() error {
		now2 := time.Now()
		log.Printf("check ms %v", (time.Now().UnixNano()-now.UnixNano())/1e6)
		now = now2
		return errors.New("testerror")
	})
}

func TestRetryCheck(t *testing.T) {
	// success
	count := 0
	success, err := NewPollingRetry(
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
	success, err = NewPollingRetry(
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

	// failed
	count = 0
	success, err = NewPollingRetry(
		WithRetry(10),
		WithInterval(time.Millisecond*10),
	).Polling(func() (bool, error) {
		count++
		return false, errors.New("testerror")
	})
	assert.Equal(t, success, false)
	assert.Equal(t, count, 1) // 直接失败
	assert.Equal(t, err != nil, true)

	// check
	count = 0
	check := func(retryCount int, err error) (bool, error) {
		if retryCount == 2 {
			return false, nil
		}
		return true, nil
	}
	success, err = NewPollingRetry(
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

}
