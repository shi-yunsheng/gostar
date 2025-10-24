package model

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/shi-yunsheng/gostar/utils"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// @en DB config
//
// @zh DB配置
type DBConfig struct {
	// @en database driver
	// @zh 数据库驱动
	Driver string `yaml:"driver"`
	// @en database dsn, if exists, other database configs are invalid
	// @zh 数据库连接字符串，如果存在，则数据库其他配置无效
	DSN string `yaml:"dsn"`
	// @en database table prefix
	// @zh 数据库表前缀
	TablePrefix string `yaml:"table_prefix"`
	// @en database host
	// @zh 数据库主机
	Host string `yaml:"host"`
	// @en database port
	// @zh 数据库端口
	Port int `yaml:"port"`
	// @en database user
	// @zh 数据库用户
	User string `yaml:"user"`
	// @en database password
	// @zh 数据库密码
	Password string `yaml:"password"`
	// @en database db
	// @zh 数据库
	Database string `yaml:"database"`
}

// @en database client
//
// @zh 数据库客户端
type DBClient struct {
	db     *gorm.DB
	mongo  *mongo.Database
	prefix string
	models map[string]any
}

var (
	dbMap     map[string]*DBClient
	dbMapLock sync.RWMutex
)

// @en initialize database
//
// @zh 初始化数据库
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
			// @en if DSN is not set, generate DSN
			// @zh 如果DSN未设置，则生成DSN
			if strings.TrimSpace(config.DSN) == "" {
				config.DSN = generateDSN(config)
			}

			var db *gorm.DB
			var err error

			switch config.Driver {
			case "mysql":
				db, err = gorm.Open(mysql.Open(config.DSN), &gorm.Config{})
			case "postgres":
				db, err = gorm.Open(postgres.Open(config.DSN), &gorm.Config{})
			case "sqlite":
				db, err = gorm.Open(sqlite.Open(config.DSN), &gorm.Config{})
			}

			if err != nil {
				panic("failed to connect to database " + name + ": " + err.Error())
			}

			// @en set table prefix
			// @zh 设置表前缀
			if config.TablePrefix != "" {
				db.NamingStrategy = schema.NamingStrategy{
					TablePrefix: config.TablePrefix,
				}
			}

			dbMap[name] = &DBClient{
				db:     db,
				prefix: config.TablePrefix,
				models: make(map[string]any),
			}

		case "mongo":
			if strings.TrimSpace(config.DSN) == "" {
				config.DSN = generateMongoDSN(config)
			}

			client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(config.DSN))
			if err != nil {
				panic("failed to connect to MongoDB " + name + ": " + err.Error())
			}

			// @en test connection
			// @zh 测试连接
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			err = client.Ping(ctx, nil)
			if err != nil {
				panic("failed to ping MongoDB " + name + ": " + err.Error())
			}

			// @en get database
			// @zh 获取数据库
			databaseName := config.Database
			if databaseName == "" {
				databaseName = fmt.Sprintf("db%s", config.Database)
			}
			dbMap[name] = &DBClient{
				mongo:  client.Database(databaseName),
				prefix: config.TablePrefix,
				models: make(map[string]any),
			}

		default:
			panic("unsupported database driver: " + config.Driver)
		}
	}
}

// @en generate DSN
//
// @zh 根据数据库类型生成对应的DSN
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

// @en generate MongoDB DSN
//
// @zh 生成MongoDB连接字符串
func generateMongoDSN(config DBConfig) string {
	if config.User != "" && config.Password != "" {
		return fmt.Sprintf("mongodb://%s:%s@%s:%d/%s",
			config.User, config.Password, config.Host, config.Port, config.Database)
	}
	return fmt.Sprintf("mongodb://%s:%d/%s",
		config.Host, config.Port, config.Database)
}

// @en get database
//
// @zh 获取数据库
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

// @en auto migrate database
//
// @zh 自动迁移数据库
func (d *DBClient) AutoMigrate(models ...any) error {
	if d.models == nil {
		d.models = make(map[string]any)
	}
	for _, model := range models {
		modelType := reflect.TypeOf(model)
		if modelType.Kind() == reflect.Pointer {
			modelType = modelType.Elem()
		}

		if tableNameMethod := reflect.ValueOf(model).MethodByName("TableName"); tableNameMethod.IsValid() {
			tableName := tableNameMethod.Call(nil)[0].String()
			d.models[tableName] = model
		} else {
			d.models[utils.CamelToSnake(modelType.Name())] = model
		}
	}
	return d.db.AutoMigrate(models...)
}
