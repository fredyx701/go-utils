package testdata

import (
	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis/v8"
)

func NewTestRedis() *redis.Client {
	mr, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	return redisClient
}
