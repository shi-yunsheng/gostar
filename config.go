package gostar

import (
	"os"

	"github.com/shi-yunsheng/gostar/model"

	"gopkg.in/yaml.v3"
)

// @en global config
//
// @zh 全局配置
type config struct {
	// @en debug mode
	//
	// @zh 调试模式
	Debug bool `yaml:"debug"`
	// @en allowed origins
	//
	// @zh 允许的来源
	AllowedOrigins []string `yaml:"allowed_origins"`
	// @en bind address and port
	//
	// @zh 绑定地址和端口
	Bind string `yaml:"bind"`
	// @en log config
	//
	// @zh 日志配置
	Log logConfig `yaml:"log"`
	// @en timezone
	//
	// @zh 时区
	Timezone string `yaml:"timezone"`
	// @en lang
	//
	// @zh 语言
	Lang string `yaml:"lang"`
	// @en database config
	//
	// @zh 数据库配置
	Database map[string]model.DBConfig `yaml:"database"`
	// @en redis config
	//
	// @zh Redis配置
	Redis map[string]model.RedisConfig `yaml:"redis"`
}

// @en generate default config
//
// @zh 生成默认配置
func generateDefaultConfig() {
	const defaultConfig = `# 调试模式，开启后会输出详细的调试信息
debug: false

# 服务绑定地址和端口
bind: 0.0.0.0:8000

# 允许的来源
allowed_origins:
  - "*"

# 日志配置
log:
    # 是否启用日志打印到控制台
    enable_print: true
    # 是否启用日志保存到文件
    #enable_save: false
    # 日志保存路径
    #save_path: logs/
    # 是否启用自动删除过期日志
    #enable_auto_delete: false
    # 日志最大保存天数，支持格式如：7d, 30d
    #max_save_days: 7d
    # 单个日志文件最大大小，支持单位：B, KB, MB, GB, TB，如：1KB, 2.5MB, 1GB
    #max_file_size: None

# 时区设置，默认使用亚洲/上海时区
#timezone: Asia/Shanghai

# 语言设置，默认使用中文
#lang: zh-CN

# 数据库配置，可以配置多个数据库连接
# 示例：
# database:
#   default:
#     driver: mysql
#     host: localhost
#     port: 3306
#     user: root
#     password: password
#     database: mydb
database:

# Redis配置，可以配置多个Redis连接
# 示例：
# redis:
#   default:
#     host: localhost
#     port: 6379
#     password: ""
#     db: 0
#     prefix: "app:"
redis:`
	_ = os.WriteFile("config.yaml", []byte(defaultConfig), 0644)
}

// @en get config
//
// @zh 获取配置
func getConfig() *config {
	if fileInfo, err := os.Stat("config.yaml"); os.IsNotExist(err) {
		generateDefaultConfig()
	} else if fileInfo.IsDir() {
		panic("config.yaml is a directory, please check the file path")
	}

	data, err := os.ReadFile("config.yaml")
	if err != nil {
		panic("read config file failed: " + err.Error())
	}

	config := &config{}
	if err := yaml.Unmarshal(data, config); err != nil {
		panic("parse config file failed: " + err.Error())
	}

	return config
}
