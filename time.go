package utils

import (
	"errors"
	"log"
	"regexp"
	"strconv"
	"time"
)

// TimeFormat 格式化时间
// opts[0]  format  yyyyMMddHHmmss  默认  yyyy-MM-dd HH:mm:ss
// opts[1]  时区  默认local
func TimeFormat(date time.Time, opts ...string) (string, error) {

	// 默认 local 时区
	if len(opts) >= 2 {
		tz, err := convertTimezone(opts[1])
		log.Println(tz)
		if err != nil {
			return "", err
		}
		if tz != 0 {
			date = time.Unix(date.Unix()+int64(tz*60), 0)
		}
		date = date.UTC()
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
// 返回 相对于UTC的分钟数
func convertTimezone(tz string) (int, error) {
	if tz == "Z" {
		return 0, nil
	}
	reg := regexp.MustCompile(`([\+\-\s])(\d\d):?(\d\d)?`)
	m := reg.FindStringSubmatch(tz)
	if m != nil {
		offset := 1
		if m[0] == "-" {
			offset = -1
		}
		hour, _ := strconv.Atoi(m[1])
		minute, _ := strconv.Atoi(m[2])
		return offset * (hour + minute/60) * 60, nil
	}
	return 0, errors.New("invalid timezone string")
}
