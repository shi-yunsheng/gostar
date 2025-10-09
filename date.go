package gostar

import "github.com/shi-yunsheng/gostar/date"

// @en initialize timezone
//
// @zh 初始化日期时区配置
func (g *goStar) initDate() {
	config := g.config.Timezone

	date.SetCurrentTimezone(config)
}
