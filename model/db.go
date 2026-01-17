package model

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/shi-yunsheng/gostar/utils"

	"github.com/glebarez/sqlite"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// DB配置
type DBConfig struct {
	// 数据库驱动
	Driver string `yaml:"driver"`
	// 数据库连接字符串，如果存在，则数据库其他配置无效
	DSN string `yaml:"dsn"`
	// 数据库表前缀
	TablePrefix string `yaml:"table_prefix"`
	// 数据库主机
	Host string `yaml:"host"`
	// 数据库端口
	Port int `yaml:"port"`
	// 数据库用户
	User string `yaml:"user"`
	// 数据库密码
	Password string `yaml:"password"`
	// 数据库
	Database string `yaml:"database"`
}

// 数据库客户端
type DBClient struct {
	db           *gorm.DB
	mongo        *mongo.Database
	prefix       string
	models       map[string]any    // 模型类型名 -> 模型实例
	tableNameMap map[string]string // 模型类型名 -> 实际表名
}

var (
	dbMap      map[string]*DBClient
	dbMapLock  sync.RWMutex
	modelCache = sync.Map{} // 模型缓存
)

var debug = false

// 开启调试模式
func EnableDebug() {
	debug = true
}

// 初始化数据库
func InitDB(config map[string]DBConfig) {
	if len(config) == 0 {
		return
	}

	dbMapLock.Lock()
	defer dbMapLock.Unlock()

	dbMap = make(map[string]*DBClient)

	for name, config := range config {
		switch config.Driver {
		case "mysql", "postgres", "sqlite":
			// 如果DSN未设置，则生成DSN
			if strings.TrimSpace(config.DSN) == "" {
				config.DSN = generateDSN(config)
			}

			var db *gorm.DB
			var err error

			conf := &gorm.Config{}
			if debug {
				conf.Logger = logger.Default.LogMode(logger.Info)
			} else {
				conf.Logger = logger.Default.LogMode(logger.Silent)
			}

			switch config.Driver {
			case "mysql":
				db, err = gorm.Open(mysql.Open(config.DSN), conf)
			case "postgres":
				db, err = gorm.Open(postgres.Open(config.DSN), conf)
			case "sqlite":
				db, err = gorm.Open(sqlite.Open(config.DSN), conf)
			}

			if err != nil {
				panic("failed to connect to database " + name + ": " + err.Error())
			}
			// 设置表前缀
			if config.TablePrefix != "" {
				db.NamingStrategy = schema.NamingStrategy{
					TablePrefix: config.TablePrefix,
				}
			}

			dbMap[name] = &DBClient{
				db:           db,
				prefix:       config.TablePrefix,
				models:       make(map[string]any),
				tableNameMap: make(map[string]string),
			}

		case "mongo":
			if strings.TrimSpace(config.DSN) == "" {
				config.DSN = generateMongoDSN(config)
			}

			client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(config.DSN))
			if err != nil {
				panic("failed to connect to MongoDB " + name + ": " + err.Error())
			}
			// 测试连接
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			err = client.Ping(ctx, nil)
			if err != nil {
				panic("failed to ping MongoDB " + name + ": " + err.Error())
			}
			// 获取数据库
			databaseName := config.Database
			if databaseName == "" {
				databaseName = fmt.Sprintf("db%s", config.Database)
			}
			dbMap[name] = &DBClient{
				mongo:        client.Database(databaseName),
				prefix:       config.TablePrefix,
				models:       make(map[string]any),
				tableNameMap: make(map[string]string),
			}

		default:
			panic("unsupported database driver: " + config.Driver)
		}
	}
}

// 根据数据库类型生成对应的DSN
func generateDSN(config DBConfig) string {
	switch config.Driver {
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&multiStatements=true",
			config.User, config.Password, config.Host, config.Port, config.Database)
	case "postgres":
		return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
			config.Host, config.User, config.Password, config.Database, config.Port)
	case "sqlite":
		return fmt.Sprintf("file:%s.db", config.Database)
	default:
		return ""
	}
}

// 生成MongoDB连接字符串
func generateMongoDSN(config DBConfig) string {
	if config.User != "" && config.Password != "" {
		return fmt.Sprintf("mongodb://%s:%s@%s:%d/%s",
			config.User, config.Password, config.Host, config.Port, config.Database)
	}
	return fmt.Sprintf("mongodb://%s:%d/%s",
		config.Host, config.Port, config.Database)
}

// 获取数据库
func GetDB(name ...string) *DBClient {
	if dbMap == nil {
		panic("database not initialized")
	}

	dbMapLock.RLock()
	defer dbMapLock.RUnlock()

	if len(name) == 0 {
		return dbMap["default"]
	}

	if _, ok := dbMap[name[0]]; !ok {
		panic("database " + name[0] + " not found")
	}

	return dbMap[name[0]]
}

// 自动迁移数据库
func (d *DBClient) AutoMigrate(models ...any) error {
	if d.models == nil {
		d.models = make(map[string]any)
	}
	if d.tableNameMap == nil {
		d.tableNameMap = make(map[string]string)
	}
	for _, model := range models {
		modelType := reflect.TypeOf(model)
		if modelType.Kind() == reflect.Pointer {
			modelType = modelType.Elem()
		}

		// if tableNameMethod := reflect.ValueOf(model).MethodByName("TableName"); tableNameMethod.IsValid() {
		// 	tableName := tableNameMethod.Call(nil)[0].String()
		// 	d.models[tableName] = model
		// } else {
		// 	d.models[utils.CamelToSnake(modelType.Name())] = model
		// }
		modelName := utils.CamelToSnake(modelType.Name())
		d.models[modelName] = model
		d.tableNameMap[modelName] = getTableName(model)
	}
	return d.db.AutoMigrate(models...)
}

// 获取模型表名
func getTableName(model any) string {
	modelValue := reflect.ValueOf(model)
	modelType := reflect.TypeOf(model)

	// 获取基础类型
	baseType := modelType
	if baseType.Kind() == reflect.Pointer {
		baseType = baseType.Elem()
	}

	// 准备用于查找方法的 Value
	var ptrValue reflect.Value
	var valueValue reflect.Value

	if modelValue.Kind() == reflect.Pointer {
		if modelValue.IsNil() {
			// 如果是 nil 指针，创建新实例
			ptrValue = reflect.New(baseType)
			valueValue = ptrValue.Elem()
		} else {
			ptrValue = modelValue
			valueValue = modelValue.Elem()
		}
	} else {
		// 值类型，创建指针
		ptrValue = modelValue.Addr()
		valueValue = modelValue
	}

	// 先尝试在指针上查找方法
	tableNameMethod := ptrValue.MethodByName("TableName")
	if !tableNameMethod.IsValid() {
		// 如果指针上找不到，尝试在值类型上查找
		tableNameMethod = valueValue.MethodByName("TableName")
	}

	// 如果找到了方法，调用它
	if tableNameMethod.IsValid() {
		results := tableNameMethod.Call(nil)
		if len(results) > 0 {
			return results[0].String()
		}
	}

	// 如果没有 TableName 方法，使用默认命名规则
	return utils.CamelToSnake(baseType.Name())
}
