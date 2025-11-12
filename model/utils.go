package model

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/shi-yunsheng/gostar/utils"
)

// 雪花算法相关常量
const (
	// 时间戳占用位数
	timestampBits = 41
	// 机器ID占用位数
	machineIDBits = 10
	// 序列号占用位数
	sequenceBits = 12
	// 最大值
	maxMachineID = -1 ^ (-1 << machineIDBits)
	maxSequence  = -1 ^ (-1 << sequenceBits)
	maxTimestamp = -1 ^ (-1 << timestampBits)
	// 位移
	machineIDShift = sequenceBits
	timestampShift = sequenceBits + machineIDBits
	// 起始时间戳 (2020-01-01 00:00:00 UTC)
	epoch = 1577836800000
)

// 雪花算法生成器
type SnowflakeGenerator struct {
	mutex     sync.Mutex
	machineID int64
	sequence  int64
	lastTime  int64
}

var (
	snowflakeGen *SnowflakeGenerator
	once         sync.Once
)

// 获取雪花算法生成器实例
func getSnowflakeGenerator() *SnowflakeGenerator {
	once.Do(func() {
		// 使用机器ID的哈希值作为机器ID，确保分布式环境下的唯一性
		machineID := int64(utils.StringHash("snowflake") & maxMachineID)
		snowflakeGen = &SnowflakeGenerator{
			machineID: machineID,
		}
	})
	return snowflakeGen
}

// 生成雪花ID
func GenerateSnowflakeID(prefix ...string) (string, error) {
	gen := getSnowflakeGenerator()
	gen.mutex.Lock()
	defer gen.mutex.Unlock()

	now := time.Now().UnixMilli()
	// 如果当前时间小于上次生成ID的时间，说明系统时钟回退
	if now < gen.lastTime {
		return "", fmt.Errorf("system clock rolled back, cannot generate ID")
	}
	// 检查时间戳是否超出范围
	timeDiff := now - epoch
	if timeDiff < 0 {
		return "", fmt.Errorf("timestamp cannot be less than start time")
	}
	if timeDiff > maxTimestamp {
		return "", fmt.Errorf("timestamp out of range, snowflake algorithm has expired")
	}
	// 如果是同一毫秒内，序列号递增
	if now == gen.lastTime {
		gen.sequence = (gen.sequence + 1) & maxSequence
		// 序列号溢出，等待下一毫秒
		if gen.sequence == 0 {
			for now <= gen.lastTime {
				now = time.Now().UnixMilli()
			}
		}
	} else {
		// 新的毫秒，序列号重置
		gen.sequence = 0
	}

	gen.lastTime = now
	// 生成ID
	id := ((now - epoch) << timestampShift) | (gen.machineID << machineIDShift) | gen.sequence
	// 添加前缀
	if len(prefix) > 0 && prefix[0] != "" {
		return fmt.Sprintf("%s%d", prefix[0], id), nil
	}
	return fmt.Sprintf("%d", id), nil
}

// 生成UUID
func GenerateUUID(prefix ...string) string {
	id := uuid.New().String()
	// 添加前缀
	if len(prefix) > 0 && prefix[0] != "" {
		return fmt.Sprintf("%s%s", prefix[0], id)
	}
	return id
}

// 生成雪花ID，失败时使用UUID作为备用方案
func GenerateSnowflakeIDSafe(prefix ...string) string {
	id, err := GenerateSnowflakeID(prefix...)
	if err != nil {
		// 如果雪花算法失败，使用UUID作为备用方案
		return GenerateUUID(prefix...)
	}
	return id
}
