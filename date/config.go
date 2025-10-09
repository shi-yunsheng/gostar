package date

import "time"

var (
	currentTimezone = "Asia/Shanghai"
)

// @en Set current timezone, default Asia/Shanghai
//
// @zh 设置当前时区，默认Asia/Shanghai
func SetCurrentTimezone(timezone string) {
	if timezone == "" {
		timezone = "Asia/Shanghai"
	}
	currentTimezone = timezone
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		panic(err)
	}
	time.Local = loc
}

// @en Get current timezone, default Asia/Shanghai
//
// @zh 获取当前时区，默认Asia/Shanghai
func GetCurrentTimezone() string {
	return currentTimezone
}
