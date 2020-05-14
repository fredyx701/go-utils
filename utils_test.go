package utils

import (
	"log"
	"testing"
	"time"

	"github.com/FredyXue/go-utils/mock"
	"github.com/stretchr/testify/assert"
)

func TestUtils(t *testing.T) {
	str1 := RandString(10)
	str2 := CreateRandDigest("test")
	log.Println(str1, str2)

	hash1 := MD5(nil)
	hash2 := MD5WithString("test")
	assert.Equal(t, hash1, "d41d8cd98f00b204e9800998ecf8427e")
	assert.Equal(t, hash2, "098f6bcd4621d373cade4e832627b4f6")
}

func TestTimeFormat(t *testing.T) {
	date, _ := time.Parse(time.RFC3339, "2020-02-07T17:34:10+08:00")
	timeStr1, err := TimeFormat(date, "yyyy-MM-dd HH:mm:ss")
	assert.NoError(t, err)
	assert.Equal(t, timeStr1, TimeFormatWithLayout(date, "2006-01-02 15:04:05"))
	assert.Equal(t, timeStr1, "2020-02-07 17:34:10")

	timeStr2, err := TimeFormat(date, "yyyy-MM-dd HH:mm:ss", "Z")
	assert.NoError(t, err)
	assert.Equal(t, timeStr2, TimeFormatWithLayout(date, "2006-01-02 15:04:05", 0))
	assert.Equal(t, timeStr2, "2020-02-07 09:34:10")

	timeStr3 := MustTimeFormat(date, "yyyy-MM-dd HH:mm:ss", "+07:00")
	assert.Equal(t, timeStr3, TimeFormatWithLayout(date, "2006-01-02 15:04:05", 7*60*60))
	assert.Equal(t, timeStr3, "2020-02-07 16:34:10")

	timeStr4 := MustTimeFormat(date, "yyyy-MM-dd HH:mm:ss", "-0100")
	assert.Equal(t, timeStr4, TimeFormatWithLayout(date, "2006-01-02 15:04:05", -60*60))
	assert.Equal(t, timeStr4, "2020-02-07 08:34:10")

	timeStr5 := MustTimeFormat(date, "yyyy-MM-dd HH:mm:ss", "-01")
	assert.Equal(t, timeStr5, TimeFormatWithLayout(date, "2006-01-02 15:04:05", -60*60))
	assert.Equal(t, timeStr5, "2020-02-07 08:34:10")

	timeStr6 := MustTimeFormat(date, "yyyy-MM-dd HH:mm:ss", "01")
	assert.Equal(t, timeStr6, TimeFormatWithLayout(date, "2006-01-02 15:04:05", 60*60))
	assert.Equal(t, timeStr6, "2020-02-07 10:34:10")
}

func TestIncrTimeWithClock(t *testing.T) {
	source, _ := time.Parse(time.RFC3339, "2020-03-01T17:34:10+08:00")

	// after
	time1 := IncrTimeWithClockUTC8(source, 30*24*3600, 18*3600+10*60+36)
	time1T, _ := time.Parse(time.RFC3339, "2020-03-31T18:10:36+08:00")

	// before
	time2 := IncrTimeWithClockUTC8(source, 30*24*3600, 15*3600+21*60+15)
	time2T, _ := time.Parse(time.RFC3339, "2020-04-01T15:21:15+08:00")

	// equal
	time3 := IncrTimeWithClockUTC8(source, 30*24*3600, 17*3600+34*60+10)
	time3T, _ := time.Parse(time.RFC3339, "2020-03-31T17:34:10+08:00")

	// clock = 0
	time4 := IncrTimeWithClockUTC8(source, 30*24*3600, 0)
	time4T, _ := time.Parse(time.RFC3339, "2020-03-31T17:34:10+08:00")

	assert.Equal(t, time1.Unix(), time1T.Unix())
	assert.Equal(t, time2.Unix(), time2T.Unix())
	assert.Equal(t, time3.Unix(), time3T.Unix())
	assert.Equal(t, time4.Unix(), time4T.Unix())
}

func TestSet(t *testing.T) {
	sets := NewSet(&mock.SetSource{}, time.Second, time.Second*2)

	assert.Equal(t, true, sets.Has(1))
	assert.Equal(t, false, sets.Has(10))
	assert.Equal(t, 5, sets.Size())

	time.Sleep(time.Second * 3)
	assert.Equal(t, 0, sets.Size())

	assert.Equal(t, true, sets.Has(1))
	assert.Equal(t, 5, sets.Size())

	arr := []interface{}{4, 5, 6, 7, 8}
	res1 := sets.Intersect(arr)
	res2 := sets.Union(arr)
	res3 := sets.Diff(arr)
	assert.ElementsMatch(t, []interface{}{4, 5}, res1)
	assert.ElementsMatch(t, []interface{}{1, 2, 3, 4, 5, 6, 7, 8}, res2)
	assert.ElementsMatch(t, []interface{}{1, 2, 3}, res3)
}

func TestMap(t *testing.T) {
	maps := NewMap(&mock.MapSource{}, time.Second, time.Second*2)
	assert.Equal(t, 1, maps.GetInt("1"))
	assert.Equal(t, int64(2), maps.GetInt64("2"))
	assert.Equal(t, "3", maps.GetString("3"))
	assert.Equal(t, true, maps.GetBool("4"))
	assert.Equal(t, 5.0, maps.GetFloat64("5"))

	v1, has := maps.Get(10)
	assert.Equal(t, false, has)
	assert.Equal(t, nil, v1)
	assert.Equal(t, 5, maps.Size())

	time.Sleep(time.Second * 3)
	assert.Equal(t, 0, maps.Size())

	assert.Equal(t, 1, maps.GetInt("1"))
	assert.Equal(t, 5, maps.Size())
}

func TestHidden(t *testing.T) {
	assert.Equal(t, HiddenName("蒙奇·D·路飞"), "***·路飞")
	assert.Equal(t, HiddenName("张三"), "张*")
	assert.Equal(t, HiddenName("李二四"), "李*四")
	assert.Equal(t, HiddenName("钱二三四"), "钱**四")
	assert.Equal(t, HiddenName("王二三四五六"), "王**四五六")
	assert.Equal(t, HiddenName("赵二三四五六七"), "赵二****七")

	assert.Equal(t, HiddenPhoneNumber("86-13712341234"), "137****1234")
	assert.Equal(t, HiddenPhoneNumber("13712341234"), "137****1234")
}
