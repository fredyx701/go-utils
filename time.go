package utils

import (
	"errors"
	"regexp"
	"strconv"
	"time"
)

// TimeFormatWithLayout 使用 time layout 格式化 时间
// offset 时区秒偏移  不传表示 local  e.g.  UTC+8 = 3600 * 8
func TimeFormatWithLayout(date time.Time, layout string, offset ...int) string {
	if len(offset) > 0 {
		date = date.In(time.FixedZone("TimeFormatWithLayout", offset[0]))
	}
	return date.Format(layout)
}

// TimeFormat 格式化时间
// opts[0]  format  yyyyMMddHHmmss  默认  yyyy-MM-dd HH:mm:ss
// opts[1]  timezone  时区  不传表示 local   支持的格式 e.g.  +08:00, 0800, +08, -07:00, -0700, -07
func TimeFormat(date time.Time, opts ...string) (string, error) {

	// 默认 local 时区
	if len(opts) >= 2 {
		timezone := opts[1]
		offset, err := convertTimezoneOffset(timezone)
		if err != nil {
			return "", err
		}
		date = date.In(time.FixedZone(timezone, offset))
	}

	year := zeroPad(date.Year(), 4)
	month := zeroPad(int(date.Month()), 2)
	day := zeroPad(date.Day(), 2)
	hour := zeroPad(date.Hour(), 2)
	minute := zeroPad(date.Minute(), 2)
	second := zeroPad(date.Second(), 2)

	var output string
	if len(opts) < 1 {
		output = year + "-" + month + "-" + day + " " + hour + ":" + minute + ":" + second
	} else {
		reg := regexp.MustCompile(`yyyy|MM|dd|HH|mm|ss`)
		output = reg.ReplaceAllStringFunc(opts[0], func(str string) string {
			switch str {
			case "yyyy":
				return year
			case "MM":
				return month
			case "dd":
				return day
			case "HH":
				return hour
			case "mm":
				return minute
			case "ss":
				return second
			default:
				return ""
			}
		})
	}
	return output, nil
}

// MustTimeFormat TimeFormat with panic
func MustTimeFormat(date time.Time, opts ...string) string {
	formatStr, err := TimeFormat(date, opts...)
	if err != nil {
		panic("utils TimeFormat Panic: " + err.Error())
	}
	return formatStr
}

// 填充0
func zeroPad(num int, length int) string {
	str := strconv.Itoa(num)
	for len(str) < length {
		str = "0" + str
	}
	return str
}

// 时区转换
// tz 'Z', +08:00, -08:00, +HH:MM or -HH:MM
// 返回 相对于UTC的秒数
func convertTimezoneOffset(tz string) (int, error) {
	if tz == "Z" {
		return 0, nil
	}
	reg := regexp.MustCompile(`([\+\-])?(\d\d):?(\d\d)?`)
	m := reg.FindStringSubmatch(tz)
	if m == nil {
		return 0, errors.New("invalid timezone string")
	}

	offset := 1
	if m[1] == "-" {
		offset = -1
	}
	hour, err := strconv.Atoi(m[2])
	minute := 0
	if m[3] != "" {
		minute, err = strconv.Atoi(m[3])
	}
	return offset * (hour*60 + minute) * 60, err
}

// IncrTimeWithClock 获取增量精确时间
// clock  (0, 86400]  目标时间点
// 目标时间小于 clock, 顺延一天
func IncrTimeWithClock(source time.Time, duration int64, clock int64, loc *time.Location) (clockTime time.Time) {

	if clock == 0 {
		return time.Unix(source.Unix()+duration, 0)
	}

	hour := clock / 3600
	minute := clock % 3600 / 60
	second := clock % 60

	time2 := time.Unix(source.Unix()+duration, 0).In(loc)
	year, month, day := time2.Date()
	time3 := time.Date(year, month, day, int(hour), int(minute), int(second), 0, loc)

	if time3.Before(time2) {
		clockTime = time3.Add(time.Hour * 24)
	} else {
		clockTime = time3
	}
	return
}

// IncrTimeWithClockUTC8 获取增量精确时间  东八区
func IncrTimeWithClockUTC8(source time.Time, duration int64, clock int64) time.Time {
	loc := time.FixedZone("UTC+8", 8*60*60)
	return IncrTimeWithClock(source, duration, clock, loc)
}
