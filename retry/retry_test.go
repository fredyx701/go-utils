package retry

import (
	"errors"
	"log"
	"testing"
	"time"
)

func TestRetry(t *testing.T) {

	now := time.Now()

	// fibonacci
	NewRetry(Options{
		Retries:  5,
		Interval: time.Millisecond * 10,
	}).Do(func() error {
		now2 := time.Now()
		log.Printf("fibonacci ms %v", (now2.UnixNano()-now.UnixNano())/1e6)
		now = now2
		return errors.New("testerror")
	})

	now = time.Now()

	// exponent
	NewRetry(Options{
		BackOff:  ExponentialBackoff,
		Interval: time.Millisecond,
	}).Do(func() error {
		log.Printf("exponent ms %v", (time.Now().UnixNano()-now.UnixNano())/1e6)
		return errors.New("testerror")
	})

	now = time.Now()

	// average
	NewRetry(Options{
		BackOff:  AverageBackOff,
		Interval: time.Millisecond * 10,
	}).Do(func() error {
		now2 := time.Now()
		log.Printf("average ms %v", (time.Now().UnixNano()-now.UnixNano())/1e6)
		now = now2
		return errors.New("testerror")
	})

	// increase
	err := NewRetry(Options{
		BackOff:  IncreaseBackOff,
		Interval: time.Millisecond * 10,
	}).Do(func() error {
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
	NewRetry(Options{
		Check:    check,
		BackOff:  AverageBackOff,
		Interval: time.Millisecond * 10,
	}).Do(func() error {
		now2 := time.Now()
		log.Printf("check ms %v", (time.Now().UnixNano()-now.UnixNano())/1e6)
		now = now2
		return errors.New("testerror")
	})

}
