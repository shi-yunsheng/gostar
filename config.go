package gostar

import (
	"os"
	"reflect"
	"strings"

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

	// @en custom config
	//
	// @zh 自定义配置
	Custom map[string]any
}

// @en generate default config
//
// @zh 生成默认配置
func generateDefaultConfig(configName string) {
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
redis:

# 自定义配置（用户可以在此添加任何自定义配置项）
# 支持嵌套结构，可通过应用上下文获取 GetConfig() 配置
# 示例：
# app_name: "我的应用"
# upload:
#   maxsize: 10485760
#   allow_types:
#     - "image/jpeg"
#     - "image/png"
# features:
#   enable_cache: true`
	_ = os.WriteFile(configName, []byte(defaultConfig), 0644)
}

// @en get config
//
// @zh 获取配置
func getConfig(name ...string) *config {
	configName := "config.yaml"
	if len(name) > 0 {
		configName = name[0]
		// @en if config name not end with .yaml, add it
		// @zh 如果配置名不以.yaml结尾，添加它
		if !strings.HasSuffix(configName, ".yaml") {
			configName = configName + ".yaml"
		}
	}

	if fileInfo, err := os.Stat(configName); os.IsNotExist(err) {
		generateDefaultConfig(configName)
	} else if fileInfo.IsDir() {
		panic(configName + " is a directory, please check the file path")
	}

	data, err := os.ReadFile(configName)
	if err != nil {
		panic("read config file failed: " + err.Error())
	}

	// @en first parse: parse to struct (framework config)
	// @zh 第一次解析：解析到结构体（框架配置）
	config := &config{}
	if err := yaml.Unmarshal(data, config); err != nil {
		panic("parse config file failed: " + err.Error())
	}

	// @en second parse: parse to map (all config)
	// @zh 第二次解析：解析到map（所有配置）
	var allConfig map[string]any
	if err := yaml.Unmarshal(data, &allConfig); err != nil {
		panic("parse config file failed: " + err.Error())
	}

	// @en get framework field names using reflection
	// @zh 通过反射获取框架字段名
	knownFields := getStructYamlTags(config)

	// @en extract custom config (remove known framework fields)
	// @zh 提取自定义配置（删除框架已知字段）
	config.Custom = make(map[string]any)
	for key, value := range allConfig {
		if !containsString(knownFields, key) {
			config.Custom[key] = value
		}
	}

	// @en if user uses custom field, merge it
	// @zh 如果用户使用了custom字段，合并进去
	if customFromYaml, ok := allConfig["custom"].(map[string]any); ok {
		for key, value := range customFromYaml {
			config.Custom[key] = value
		}
	}

	return config
}

// @en get struct yaml tags using reflection
//
// @zh 使用反射获取结构体的yaml标签
func getStructYamlTags(v any) []string {
	var tags []string
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		yamlTag := field.Tag.Get("yaml")
		if yamlTag != "" {
			// @en extract tag name (remove options like omitempty)
			// @zh 提取标签名（去掉选项，如omitempty）
			if idx := strings.Index(yamlTag, ","); idx != -1 {
				yamlTag = yamlTag[:idx]
			}
			if yamlTag != "" && yamlTag != "-" {
				tags = append(tags, yamlTag)
			}
		}
	}
	return tags
}

// @en check if string slice contains item
//
// @zh 检查字符串切片是否包含项
func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// @en get custom config value by path (supports nested keys like "upload.maxsize")
//
// @zh 通过路径获取自定义配置值（支持嵌套键，如 "upload.maxsize"）
func getConfigValue(custom map[string]any, path string, defaultVal ...any) any {
	if custom == nil {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return nil
	}

	keys := strings.Split(path, ".")
	var current any = custom

	for _, key := range keys {
		if m, ok := current.(map[string]any); ok {
			if val, exists := m[key]; exists {
				current = val
			} else {
				if len(defaultVal) > 0 {
					return defaultVal[0]
				}
				return nil
			}
		} else {
			if len(defaultVal) > 0 {
				return defaultVal[0]
			}
			return nil
		}
	}

	return current
}

// @en get custom config by path (supports nested keys like "upload.maxsize")
//
// @zh 通过路径获取自定义配置（支持嵌套键，如 "upload.maxsize"）
func (g *goStar) GetConfig(path string, defaultVal ...any) any {
	if g.config == nil {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return nil
	}
	return getConfigValue(g.config.Custom, path, defaultVal...)
}

// @en get custom config string by path
//
// @zh 通过路径获取自定义配置字符串
func (g *goStar) GetConfigString(path string, defaultVal ...string) string {
	val := g.GetConfig(path)
	if val == nil {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return ""
	}
	if str, ok := val.(string); ok {
		return str
	}
	if len(defaultVal) > 0 {
		return defaultVal[0]
	}
	return ""
}

// @en get custom config int by path
//
// @zh 通过路径获取自定义配置整数
func (g *goStar) GetConfigInt(path string, defaultVal ...int) int {
	val := g.GetConfig(path)
	if val == nil {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return 0
	}
	if num, ok := val.(int); ok {
		return num
	}
	if len(defaultVal) > 0 {
		return defaultVal[0]
	}
	return 0
}

// @en get custom config bool by path
//
// @zh 通过路径获取自定义配置布尔值
func (g *goStar) GetConfigBool(path string, defaultVal ...bool) bool {
	val := g.GetConfig(path)
	if val == nil {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return false
	}
	if b, ok := val.(bool); ok {
		return b
	}
	if len(defaultVal) > 0 {
		return defaultVal[0]
	}
	return false
}

// @en get custom config float by path
//
// @zh 通过路径获取自定义配置浮点数
func (g *goStar) GetConfigFloat(path string, defaultVal ...float64) float64 {
	val := g.GetConfig(path)
	if val == nil {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return 0
	}
	// @en YAML may parse as float64 or int
	// @zh YAML可能解析为float64或int
	switch v := val.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case float32:
		return float64(v)
	}
	if len(defaultVal) > 0 {
		return defaultVal[0]
	}
	return 0
}

// @en get custom config map by path
//
// @zh 通过路径获取自定义配置映射
func (g *goStar) GetConfigMap(path string) map[string]any {
	val := g.GetConfig(path)
	if val == nil {
		return nil
	}
	if m, ok := val.(map[string]any); ok {
		return m
	}
	return nil
}

// @en get custom config slice by path
//
// @zh 通过路径获取自定义配置切片
func (g *goStar) GetConfigSlice(path string) []any {
	val := g.GetConfig(path)
	if val == nil {
		return nil
	}
	if s, ok := val.([]any); ok {
		return s
	}
	return nil
}

// @en get custom config string slice by path
//
// @zh 通过路径获取自定义配置字符串切片
func (g *goStar) GetConfigStringSlice(path string) []string {
	val := g.GetConfig(path)
	if val == nil {
		return nil
	}
	if s, ok := val.([]any); ok {
		result := make([]string, 0, len(s))
		for _, item := range s {
			if str, ok := item.(string); ok {
				result = append(result, str)
			}
		}
		return result
	}
	return nil
}
