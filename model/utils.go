package model

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/shi-yunsheng/gostar/utils"
)

// @en snowflake algorithm related constants
//
// @zh 雪花算法相关常量
const (
	// @en timestamp bits
	// @zh 时间戳占用位数
	timestampBits = 41
	// @en machine ID bits
	// @zh 机器ID占用位数
	machineIDBits = 10
	// @en sequence bits
	// @zh 序列号占用位数
	sequenceBits = 12

	// @en maximum values
	// @zh 最大值
	maxMachineID = -1 ^ (-1 << machineIDBits)
	maxSequence  = -1 ^ (-1 << sequenceBits)
	maxTimestamp = -1 ^ (-1 << timestampBits)

	// @en bit shifts
	// @zh 位移
	machineIDShift = sequenceBits
	timestampShift = sequenceBits + machineIDBits

	// @en start timestamp (2020-01-01 00:00:00 UTC)
	// @zh 起始时间戳 (2020-01-01 00:00:00 UTC)
	epoch = 1577836800000
)

// @en snowflake algorithm generator
//
// @zh 雪花算法生成器
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

// @en get snowflake generator instance
//
// @zh 获取雪花算法生成器实例
func getSnowflakeGenerator() *SnowflakeGenerator {
	once.Do(func() {
		// @en use hash value of machine ID as machine ID to ensure uniqueness in distributed environment
		// @zh 使用机器ID的哈希值作为机器ID，确保分布式环境下的唯一性
		machineID := int64(utils.StringHash("snowflake") & maxMachineID)
		snowflakeGen = &SnowflakeGenerator{
			machineID: machineID,
		}
	})
	return snowflakeGen
}

// @en generate snowflake ID
//
// @zh 生成雪花ID
func GenerateSnowflakeID(prefix ...string) (string, error) {
	gen := getSnowflakeGenerator()
	gen.mutex.Lock()
	defer gen.mutex.Unlock()

	now := time.Now().UnixMilli()

	// @en if current time is less than last ID generation time, system clock has rolled back
	// @zh 如果当前时间小于上次生成ID的时间，说明系统时钟回退
	if now < gen.lastTime {
		return "", fmt.Errorf("system clock rolled back, cannot generate ID")
	}

	// @en check if timestamp is out of range
	// @zh 检查时间戳是否超出范围
	timeDiff := now - epoch
	if timeDiff < 0 {
		return "", fmt.Errorf("timestamp cannot be less than start time")
	}
	if timeDiff > maxTimestamp {
		return "", fmt.Errorf("timestamp out of range, snowflake algorithm has expired")
	}

	// @en if in the same millisecond, increment sequence number
	// @zh 如果是同一毫秒内，序列号递增
	if now == gen.lastTime {
		gen.sequence = (gen.sequence + 1) & maxSequence
		// @en sequence number overflow, wait for next millisecond
		// @zh 序列号溢出，等待下一毫秒
		if gen.sequence == 0 {
			for now <= gen.lastTime {
				now = time.Now().UnixMilli()
			}
		}
	} else {
		// @en new millisecond, reset sequence number
		// @zh 新的毫秒，序列号重置
		gen.sequence = 0
	}

	gen.lastTime = now

	// @en generate ID
	// @zh 生成ID
	id := ((now - epoch) << timestampShift) | (gen.machineID << machineIDShift) | gen.sequence

	// @en add prefix
	// @zh 添加前缀
	if len(prefix) > 0 && prefix[0] != "" {
		return fmt.Sprintf("%s%d", prefix[0], id), nil
	}
	return fmt.Sprintf("%d", id), nil
}

// @en generate UUID
//
// @zh 生成UUID
func GenerateUUID(prefix ...string) string {
	id := uuid.New().String()

	// @en add prefix
	// @zh 添加前缀
	if len(prefix) > 0 && prefix[0] != "" {
		return fmt.Sprintf("%s%s", prefix[0], id)
	}
	return id
}

// @en generate snowflake ID with fallback to UUID
//
// @zh 生成雪花ID，失败时使用UUID作为备用方案
func GenerateSnowflakeIDSafe(prefix ...string) string {
	id, err := GenerateSnowflakeID(prefix...)
	if err != nil {
		// @en fallback to UUID if snowflake fails
		// @zh 如果雪花算法失败，使用UUID作为备用方案
		return GenerateUUID(prefix...)
	}
	return id
}
