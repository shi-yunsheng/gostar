package logger

import (
	"os"
	"path/filepath"
	"time"

	"github.com/shi-yunsheng/gostar/date"
)

// 自动删除日志
func autoDeleteLogs() {
	if !enableAutoDelete {
		return
	}

	duration, err := date.ParseTimeDuration(maxSaveDays)
	if err != nil {
		panic(err)
	}

	for {
		// 获取过期时间
		expiredTime := time.Now().Add(-duration)

		filepath.Walk(savePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}

			if !info.IsDir() && filepath.Ext(path) == ".log" {
				fileName := filepath.Base(path)
				if isLogFile(fileName) {
					// 检查文件是否过期
					if info.ModTime().Before(expiredTime) {
						// 删除过期日志文件，忽略错误
						os.Remove(path)
					}
				}
			}

			return nil
		})

		time.Sleep(duration)
	}
}
