package utils

import (
	"io"
	"time"
)

// @en throttled reader for limiting read speed
//
// @zh 限速读取器，用于限制读取速度
type ThrottledReader struct {
	// @en underlying reader
	//
	// @zh 底层读取器
	reader io.ReadSeeker
	// @en bytes per second limit
	//
	// @zh 每秒字节数限制
	bytesPerSecond int64
	// @en start time for speed calculation
	//
	// @zh 速度计算的开始时间
	timeStart time.Time
	// @en total bytes read
	//
	// @zh 已读取的总字节数
	bytesRead int64
}

// @en create a new throttled reader
//
// @zh 创建新的限速读取器
func NewThrottledReader(reader io.ReadSeeker, bytesPerSecond int64) *ThrottledReader {
	return &ThrottledReader{
		reader:         reader,
		bytesPerSecond: bytesPerSecond,
		timeStart:      time.Now(),
		bytesRead:      0,
	}
}

// @en read data with speed limit
//
// @zh 以限速方式读取数据
func (r *ThrottledReader) Read(p []byte) (n int, err error) {
	n, err = r.reader.Read(p)
	r.bytesRead += int64(n)

	// @en calculate expected time based on bytes read and speed limit
	// @zh 根据已读取字节数和速度限制计算期望时间
	expectedTime := time.Duration(float64(r.bytesRead) / float64(r.bytesPerSecond) * float64(time.Second))
	realTime := time.Since(r.timeStart)

	// @en if actual time is less than expected time, wait
	// @zh 如果实际时间小于期望时间，则等待
	if realTime < expectedTime {
		time.Sleep(expectedTime - realTime)
	}

	return n, err
}

// @en seek to a specific position and reset speed calculation
//
// @zh 定位到指定位置并重置速度计算
func (r *ThrottledReader) Seek(offset int64, whence int) (int64, error) {
	r.timeStart = time.Now()
	r.bytesRead = 0
	return r.reader.Seek(offset, whence)
}
