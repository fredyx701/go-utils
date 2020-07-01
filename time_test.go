package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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
