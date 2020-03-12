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
	log.Println(str1)
}

func TestTime(t *testing.T) {
	date, _ := time.Parse(time.RFC3339, "2020-02-07T17:34:10+08:00")
	timeStr1, err := TimeFormat(date, "yyyy-MM-dd HH:mm:ss")
	assert.NoError(t, err)
	assert.Equal(t, timeStr1, "2020-02-07 17:34:10")

	timeStr2, err := TimeFormat(date, "yyyy-MM-dd HH:mm:ss", "Z")
	assert.NoError(t, err)
	assert.Equal(t, timeStr2, "2020-02-07 09:34:10")

	timeStr3 := MustTimeFormat(date, "yyyy-MM-dd HH:mm:ss", "+07:00")
	assert.Equal(t, timeStr3, "2020-02-07 16:34:10")

	timeStr4 := MustTimeFormat(date, "yyyy-MM-dd HH:mm:ss", "-0100")
	assert.Equal(t, timeStr4, "2020-02-07 08:34:10")

	timeStr5 := MustTimeFormat(date, "yyyy-MM-dd HH:mm:ss", "-01")
	assert.Equal(t, timeStr5, "2020-02-07 08:34:10")

	timeStr6 := MustTimeFormat(date, "yyyy-MM-dd HH:mm:ss", "01")
	assert.Equal(t, timeStr6, "2020-02-07 10:34:10")
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
