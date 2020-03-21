package utils

import (
	"crypto/md5"
	"encoding/hex"
	"math/rand"
	"strconv"
	"time"
)

//Rand custom rand
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

// MD5 generate md5 digest
func MD5(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}

// MD5WithString generate md5 digest with string
func MD5WithString(text string) string {
	return MD5([]byte(text))
}

// CreateRandDigest  create random digest string
func CreateRandDigest(opts ...string) string {
	now := time.Now()
	str := RandString(20) + strconv.Itoa(int(now.Unix())) + strconv.Itoa(now.Nanosecond())
	for _, v := range opts {
		str += v
	}
	return MD5WithString(str)
}
