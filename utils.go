package utils

import (
	"math/rand"
	"time"
)

//Rand  自定义 rand
var Rand *rand.Rand

func init() {
	seed := (time.Now().Unix() + int64(rand.Int31())) / 2
	Rand = rand.New(rand.NewSource(seed))
}

// RandString 生成指定位数的随机字符串
func RandString(length int) string {
	str := "0123456789abcdefghijklmnopqrstuvwyxzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	strLen := len(str)
	randStr := make([]byte, length)

	for i := 0; i < length; i++ {
		randStr[i] = str[Rand.Intn(strLen)]
	}
	return string(randStr)
}
