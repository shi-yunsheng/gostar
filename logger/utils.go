package logger

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/shi-yunsheng/gostar/date"
	"github.com/shi-yunsheng/gostar/utils"
)

// 检查文件名是否为日志文件
func isLogFile(fileName string) bool {
	// 移除 .log 扩展名
	nameWithoutExt := strings.TrimSuffix(fileName, ".log")
	// single file format: YYYY-MM-DD and chunked format: YYYY-MM-DD.number
	// 使用正则表达式匹配日志文件名格式
	// 单文件格式：YYYY-MM-DD 和分片格式：YYYY-MM-DD.数字
	matched, _ := regexp.MatchString(`^\d{4}-\d{2}-\d{2}(\.\d+)?$`, nameWithoutExt)

	return matched
}

// 最后一个日志文件索引
var lastLogFileIndex int = -1

// 保存日志
func saveLog(message string) {
	if !enableSave {
		return
	}

	currentDate := date.GetToday(date.FORMAT_DATE)
	// 分片的情况下，需额外创建日期目录
	dateDir := ""
	if maxLogSize > 0 {
		dateDir = currentDate
	}
	// 检查日志目录是否存在
	if dirInfo, err := os.Stat(utils.JoinPath(savePath, dateDir)); os.IsNotExist(err) {
		os.MkdirAll(utils.JoinPath(savePath, dateDir), os.ModePerm)
		// 确保每次新目录创建时，lastLogFileIndex 为 -1
		lastLogFileIndex = -1
	} else if !dirInfo.IsDir() {
		panic("log save path is not a directory")
	}

	var logFilename string
	if maxLogSize > 0 {
		if lastLogFileIndex == -1 {
			files, _ := os.ReadDir(utils.JoinPath(savePath, dateDir))

			logFileCount := 0
			for _, file := range files {
				if isLogFile(file.Name()) {
					logFileCount++
				}
			}

			lastLogFileIndex = logFileCount
		}
		// 如果文件大小超过最大大小，则创建新的日志分片
		fileInfo, err := os.Stat(utils.JoinPath(savePath, dateDir, fmt.Sprintf("%s.%d.log", currentDate, lastLogFileIndex)))
		if err != nil || fileInfo.Size() > maxLogSize {
			lastLogFileIndex++
		}

		logFilename = utils.JoinPath(savePath, dateDir, fmt.Sprintf("%s.%d.log", currentDate, lastLogFileIndex))
	} else {
		logFilename = utils.JoinPath(savePath, dateDir, currentDate+".log")
	}
	// 打开日志文件，如果文件不存在，则创建文件
	file, err := os.OpenFile(logFilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer file.Close()

	file.WriteString(message + "\n")
}
