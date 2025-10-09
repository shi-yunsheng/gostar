package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shi-yunsheng/gostar/date"
	"github.com/shi-yunsheng/gostar/utils"
)

var (
	enablePrint = true
	enableSave  = false
	savePath    = func() string {
		execPath, _ := os.Executable()
		execDir := filepath.Dir(execPath)
		return execDir + "/logs/"
	}()
	enableAutoDelete = false
	maxSaveDays      = "7d"
	dateFormat       = date.FORMAT_DATETIME
	logFormat        = "[%s] [%s] %s"
)

// @en enable log print
//
// @zh 启用日志打印
func DisablePrint() {
	enablePrint = false
	if !enableSave {
		fmt.Println("Log printing is disabled and log saving is not enabled. Some important information may be missed.")
	}
}

// @en enable log save
//
// @zh 启用日志保存
func EnableSave() {
	enableSave = true
}

// @en set log save path
//
// @zh 设置日志保存路径
func SetSavePath(path string) {
	savePath = path
}

// @en enable auto delete
//
// @zh 启用自动删除
func EnableAutoDelete() {
	enableAutoDelete = true
	go autoDeleteLogs()
}

// @en max logs save days, only support days, like: 1d, 2d, 3d, default 7d
//
// @zh 日志保存天数，仅支持天数，如：1d, 2d, 3d，默认7d
func SetMaxSaveDays(days string) {
	maxSaveDays = days
}

// @en max single log file size, support units: B, KB, MB, GB, TB, like: 1KB, 2.5MB, 1GB, default None
//
// @zh 单个日志文件最大大小，支持单位：B, KB, MB, GB, TB，如：1KB, 2.5MB, 1GB，默认不限制
func SetMaxSingleLogFileSize(size string) {
	if strings.TrimSpace(size) != "None" && strings.TrimSpace(size) != "" {
		size, err := utils.ParseSize(size)
		if err != nil {
			panic(err)
		}
		maxLogSize = size
	}
}

// @en set channel buffer size
//
// @zh 设置通道缓冲区大小
func SetChannelBufferSize(size int) {
	chanBufferSize = size
}

// @en set date format, default: 2006-01-02 15:04:05
//
// @zh 设置日期格式，默认2006-01-02 15:04:05
func SetDateFormat(format date.DateFormat) {
	dateFormat = format
}

// @en set log format, default: [%s] %s: %s, first %s is time, second %s is level, third %s is message
//
// @zh 设置日志格式，默认：[%s] %s: %s，第一个 %s 是时间，第二个 %s 是级别，第三个 %s 是消息
func SetLogFormat(format string) {
	logFormat = format
}
