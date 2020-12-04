package utils

import (
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStrings(t *testing.T) {
	str1 := RandString(10)
	str2 := CreateRandDigest("test")
	str3 := RandString(10, "012345")
	log.Println(str1, str2, str3)

	hash1 := MD5(nil)
	hash2 := MD5WithString("test")
	assert.Equal(t, hash1, "d41d8cd98f00b204e9800998ecf8427e")
	assert.Equal(t, hash2, "098f6bcd4621d373cade4e832627b4f6")
}

func TestSlice(t *testing.T) {
	// UniqueArraySubSequences
	arr1 := []interface{}{6, 8, 7, 10, 9, 11, 13, 14, 12, 15, 1, 2, 3, 4, 5, 0}
	arr2 := []interface{}{9, 0, 1, 2, 3, 4, 5, 15, 6, 8, 7, 10, 11, 13, 14, 12}
	assert.Equal(t, GetUniqueArraySubSequences(arr1, arr2), [][]interface{}{{6, 8, 7, 10}, {11, 13, 14, 12}, {1, 2, 3, 4, 5}})

	arr1 = []interface{}{6, 8, 7, 10, 9, 11, 13, 14, 12, 15, 1, 2, 3, 4, 5, 0}
	arr2 = []interface{}{9, 0, 1, 2, 3, 4, 5, 15, 11, 13, 14, 12}
	assert.Equal(t, GetUniqueArraySubSequences(arr1, arr2), [][]interface{}{{11, 13, 14, 12}, {1, 2, 3, 4, 5}})

	arr1 = []interface{}{9, 11, 13, 14, 12, 15, 1, 2, 3, 4, 5, 0}
	arr2 = []interface{}{9, 0, 1, 2, 3, 4, 5, 15, 6, 8, 7, 10, 11, 13, 14, 12}
	assert.Equal(t, GetUniqueArraySubSequences(arr1, arr2), [][]interface{}{{11, 13, 14, 12}, {1, 2, 3, 4, 5}})

	// set
	set := NewSet([]interface{}{6, 8, 7, 1, 2, 3, 4, 5, 0, 0, 1})
	arr := []interface{}{1, 2, 3, 11, 12, 13}
	assert.ElementsMatch(t, set.Intersect(arr), []interface{}{1, 2, 3})
	assert.ElementsMatch(t, set.Union(arr), []interface{}{6, 8, 7, 1, 2, 3, 4, 5, 0, 11, 12, 13})
	assert.ElementsMatch(t, set.Diff(arr), []interface{}{6, 8, 7, 4, 5, 0})
}

func TestHidden(t *testing.T) {
	assert.Equal(t, HiddenName("蒙奇·D·路飞"), "***·路飞")
	assert.Equal(t, HiddenName("张三"), "张*")
	assert.Equal(t, HiddenName("李二四"), "李*四")
	assert.Equal(t, HiddenName("钱二三四"), "钱**四")
	assert.Equal(t, HiddenName("王二三四五六"), "王**四五六")
	assert.Equal(t, HiddenName("赵二三四五六七"), "赵二****七")

	assert.Equal(t, HiddenPhoneNumber("86-13712341234"), "86-137****1234")
	assert.Equal(t, HiddenPhoneNumber("13712341234"), "137****1234")
	assert.Equal(t, HiddenPhoneNumber("65-96123412"), "65-96****12")
	assert.Equal(t, HiddenPhoneNumber("96123412"), "96****12")
	assert.Equal(t, HiddenPhoneNumber("852-94123412"), "852-94****12")
	assert.Equal(t, HiddenPhoneNumber("94123412"), "94****12")
	assert.Equal(t, HiddenPhoneNumber("64-02112341234"), "64-021****1234")
	assert.Equal(t, HiddenPhoneNumber("02112341234"), "021****1234")
	assert.Equal(t, HiddenPhoneNumber("1-9291234123"), "1-929****123")
	assert.Equal(t, HiddenPhoneNumber("9291234123"), "929****123")
}

func TestProtect(t *testing.T) {
	go Protect(func() {
		log.Panic("test panic in protect goroutine")
	})

	time.Sleep(time.Second)
	log.Println("protect end")
}
