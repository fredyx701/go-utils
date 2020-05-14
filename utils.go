package utils

import (
	"crypto/md5"
	"encoding/hex"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
)

//Rand custom rand
var Rand *rand.Rand

func init() {
	seed := (time.Now().Unix() + int64(rand.Int31())) / 2
	Rand = rand.New(rand.NewSource(seed))
}

// RandString 生成指定位数的随机字符串
// template 自定字符串合集
func RandString(length int, template ...string) string {
	str := "0123456789abcdefghijklmnopqrstuvwyxzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	if len(template) > 0 {
		str = template[0]
	}
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

// HiddenName 名称脱敏
// 张三  张*
// 李二四  李*四
func HiddenName(name string) string {
	name = strings.TrimSpace(name)
	if strings.Contains(name, "·") { //  AAA·BBB
		names := strings.Split(name, "·")
		return "***·" + names[len(names)-1]
	}
	nameChar := []rune(name)
	length := len(nameChar)
	if length <= 1 { // A
		return name
	}
	if length <= 2 { // AA
		return string([]rune{nameChar[0], '*'})
	}
	if length <= 3 { // AAA
		return string([]rune{nameChar[0], '*', nameChar[2]})
	}
	if length <= 6 {
		return string(nameChar[0:1]) + "**" + string(nameChar[3:])
	}
	return string(nameChar[0:2]) + "****" + string(nameChar[6:])
}

// HiddenPhoneNumber 手机号 脱敏
// 86-13912341234  139****1234
func HiddenPhoneNumber(phone string) string {
	reg := regexp.MustCompile(`(\d{3})\d{4}(\d{4})`)
	phone = reg.FindString(phone)
	return reg.ReplaceAllString(phone, "$1****$2")
}
