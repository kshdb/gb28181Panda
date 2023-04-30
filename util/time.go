package util

import "time"

const TIME_LAYOUT = "2006-01-02T15:04:05"
const TIME_LAYOUT_With_SPACE = "2006-01-02 15:04:05"

func GetCurrenTimeNowFormat() string {
	return time.Now().Format(TIME_LAYOUT)
}
func GetCurrenTimeNow() string {
	return time.Now().Format(TIME_LAYOUT_With_SPACE)
}
