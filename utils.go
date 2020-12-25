package utils

import (
	"log"

	"github.com/pkg/errors"
)

// MergeError 合并 error
func MergeError(errs ...error) error {
	var resErr error
	for _, err := range errs {
		if err != nil {
			if resErr == nil {
				resErr = err
			} else {
				resErr = errors.Wrap(resErr, err.Error())
			}
		}
	}
	return resErr
}

// Protect panic protect
func Protect(g func()) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[Panic] catch panic: %v", r)
		}
	}()
	g()
}
