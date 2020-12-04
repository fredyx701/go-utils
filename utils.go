package utils

import (
	"log"
)

// MergeError 合并 error
func MergeError(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
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
