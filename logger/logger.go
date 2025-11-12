// 一个简单的日志系统
package logger

import (
	"fmt"
	"sync"
)

var (
	logChan        chan *logMessage
	chanBufferSize = 10000
	wg             sync.WaitGroup
	once           sync.Once
	done           chan struct{}
	maxLogSize     int64 = -1
)

// 日志消息结构
type logMessage struct {
	level   string
	message string
}

// 初始化日志系统
func init() {
	initLogger()
}

// 初始化日志系统
func initLogger() {
	once.Do(func() {
		logChan = make(chan *logMessage, chanBufferSize)
		done = make(chan struct{})

		wg.Add(1)
		go consumeLogs()
	})
}

// 从通道消费日志
func consumeLogs() {
	defer wg.Done()

	for {
		select {
		case msg := <-logChan:
			formattedMsg := getLogLevelMessage(msg.level, msg.message)
			saveLog(formattedMsg)

			if enablePrint {
				fmt.Println(formattedMsg)
			}

		case <-done:
			for len(logChan) > 0 {
				msg := <-logChan
				formattedMsg := getLogLevelMessage(msg.level, msg.message)
				saveLog(formattedMsg)

				if enablePrint {
					fmt.Println(formattedMsg)
				}
			}
			return
		}
	}
}

// 基础打印
func basePrint(level string, message string, args ...any) {
	message = fmt.Sprintf(message, args...)

	select {
	case logChan <- &logMessage{level: level, message: message}:
	default:
		fmt.Printf("[WARN] Log channel is full, message dropped: %s\n", message)
	}
}

// 优雅关闭
func Close() {
	close(done)
	wg.Wait()
	close(logChan)
}

// 信息打印
func I(message string, args ...any) {
	basePrint("INFO", message, args...)
}

// 警告打印
func W(message string, args ...any) {
	basePrint("WARN", message, args...)
}

// 错误打印
func E(message string, args ...any) {
	basePrint("ERROR", message, args...)
}

// 成功打印
func S(message string, args ...any) {
	basePrint("SUCCESS", message, args...)
}

// 调试打印
func D(message string, args ...any) {
	basePrint("DEBUG", message, args...)
}

// 普通打印
func P(message string, args ...any) {
	basePrint("PRINT", message, args...)
}
