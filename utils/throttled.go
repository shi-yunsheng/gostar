package utils

import (
	"io"
	"time"
)

// 限速读取器，用于限制读取速度
type ThrottledReader struct {
	// 底层读取器
	reader io.ReadSeeker
	// 每秒字节数限制
	bytesPerSecond int64
	// 速度计算的开始时间
	timeStart time.Time
	// 已读取的总字节数
	bytesRead int64
}

// 创建新的限速读取器
func NewThrottledReader(reader io.ReadSeeker, bytesPerSecond int64) *ThrottledReader {
	return &ThrottledReader{
		reader:         reader,
		bytesPerSecond: bytesPerSecond,
		timeStart:      time.Now(),
		bytesRead:      0,
	}
}

// 以限速方式读取数据
func (r *ThrottledReader) Read(p []byte) (n int, err error) {
	n, err = r.reader.Read(p)
	r.bytesRead += int64(n)
	// 根据已读取字节数和速度限制计算期望时间
	expectedTime := time.Duration(float64(r.bytesRead) / float64(r.bytesPerSecond) * float64(time.Second))
	realTime := time.Since(r.timeStart)
	// 如果实际时间小于期望时间，则等待
	if realTime < expectedTime {
		time.Sleep(expectedTime - realTime)
	}

	return n, err
}

// 定位到指定位置并重置速度计算
func (r *ThrottledReader) Seek(offset int64, whence int) (int64, error) {
	r.timeStart = time.Now()
	r.bytesRead = 0
	return r.reader.Seek(offset, whence)
}
