package model

import (
	"strconv"
	"strings"
	"sync"

	"github.com/shi-yunsheng/gostar/date"

	"github.com/go-redis/redis"
)

// @en Redis config
//
// @zh Redis配置
type RedisConfig struct {
	// @en redis dsn, if exists, other redis configs are invalid
	// @zh Redis连接字符串，如果存在，则Redis其他配置无效
	DSN string `yaml:"dsn"`
	// @en redis host
	// @zh Redis主机
	Host string `yaml:"host"`
	// @en redis port
	// @zh Redis端口
	Port int `yaml:"port"`
	// @en redis password
	// @zh Redis密码
	Password string `yaml:"password"`
	// @en redis db
	// @zh Redis
	DB int `yaml:"db"`
	// @en redis key prefix
	// @zh Redis键前缀
	Prefix string `yaml:"prefix"`
}

// @en redis client
//
// @zh Redis客户端
type RedisClient struct {
	client *redis.Client
	prefix string
}

var (
	redisMap     map[string]*RedisClient
	redisMapLock sync.RWMutex
)

// @en initialize redis
//
// @zh 初始化Redis
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

// @en get redis
//
// @zh 获取Redis
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

// @en get value from redis
//
// @zh 从Redis中获取值
func (r *RedisClient) Get(key string) (string, error) {
	return r.client.Get(r.prefix + key).Result()
}

// @en set value to redis
//
// @zh 设置值到Redis
func (r *RedisClient) Set(key string, value any, expiration string) error {
	duration, err := date.ParseTimeDuration(expiration)
	if err != nil {
		return err
	}
	return r.client.Set(r.prefix+key, value, duration).Err()
}

// @en delete value from redis
//
// @zh 从Redis中删除值
func (r *RedisClient) Del(key string) error {
	return r.client.Del(r.prefix + key).Err()
}

// @en check if value exists
//
// @zh 检查值是否存在
func (r *RedisClient) Exists(key string) (int64, error) {
	return r.client.Exists(r.prefix + key).Result()
}
