package model

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/shi-yunsheng/gostar/date"

	"github.com/go-redis/redis"
)

// Redis配置
type RedisConfig struct {
	// Redis连接字符串，如果存在，则Redis其他配置无效
	DSN string `yaml:"dsn"`
	// Redis主机
	Host string `yaml:"host"`
	// Redis端口
	Port int `yaml:"port"`
	// Redis密码
	Password string `yaml:"password"`
	// Redis
	DB int `yaml:"db"`
	// Redis键前缀
	Prefix string `yaml:"prefix"`
}

// Redis客户端
type RedisClient struct {
	client *redis.Client
	prefix string
}

var (
	redisMap     map[string]*RedisClient
	redisMapLock sync.RWMutex
)

// 初始化Redis
func InitRedis(config map[string]RedisConfig) {
	if len(config) == 0 {
		return
	}

	redisMapLock.Lock()
	defer redisMapLock.Unlock()

	redisMap = make(map[string]*RedisClient)

	for name, config := range config {
		if strings.TrimSpace(config.DSN) != "" {
			opt, err := redis.ParseURL(config.DSN)
			if err != nil {
				panic("parse redis dsn failed: " + err.Error())
			}
			redisMap[name] = &RedisClient{
				client: redis.NewClient(opt),
				prefix: config.Prefix,
			}
		} else {
			redisMap[name] = &RedisClient{
				client: redis.NewClient(&redis.Options{
					Addr:     config.Host + ":" + strconv.Itoa(config.Port),
					Password: config.Password,
					DB:       config.DB,
				}),
				prefix: config.Prefix,
			}
		}
	}
}

// 获取Redis
func GetRedis(name ...string) *RedisClient {
	if redisMap == nil {
		panic("redis not initialized")
	}

	redisMapLock.RLock()
	defer redisMapLock.RUnlock()

	if len(name) == 0 {
		return redisMap["default"]
	}

	if _, ok := redisMap[name[0]]; !ok {
		panic("redis " + name[0] + " not found")
	}

	return redisMap[name[0]]
}

// 从Redis中获取值
func (r *RedisClient) Get(key string) (string, error) {
	return r.client.Get(r.prefix + key).Result()
}

// 设置值到Redis
func (r *RedisClient) Set(key string, value any, expiration string) error {
	duration, err := date.ParseTimeDuration(expiration)
	if err != nil {
		return err
	}
	return r.client.Set(r.prefix+key, value, duration).Err()
}

// 设置值到Redis，如果key存在，则不设置
func (r *RedisClient) SetNX(key string, value any, expiration string) (bool, error) {
	duration, err := date.ParseTimeDuration(expiration)
	if err != nil {
		return false, err
	}
	return r.client.SetNX(r.prefix+key, value, duration).Result()
}

// 从Redis中删除值
func (r *RedisClient) Del(key string) error {
	return r.client.Del(r.prefix + key).Err()
}

// 检查值是否存在
func (r *RedisClient) Exists(key string) (int64, error) {
	return r.client.Exists(r.prefix + key).Result()
}

// 获取生命周期
func (r *RedisClient) GetExpiration(key string) (time.Duration, error) {
	return r.client.TTL(r.prefix + key).Result()
}
