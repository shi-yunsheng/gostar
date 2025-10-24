package model

import (
	"regexp"
	"sync"
	"time"

	"gorm.io/gorm"
)

// @en default pagination constants
//
// @zh 默认列表相关
const (
	defaultPageSize = 20
	defaultPageNo   = 1
	maxPageSize     = 1000
	defaultOrderBy  = "`updated_at` DESC, `id` DESC"
)

var (
	// @en model cache
	// @zh 模型缓存
	modelCache = sync.Map{}
	// @en operator regex
	// @zh 操作符正则
	operatorRegex = regexp.MustCompile(`^(<>|>=|<=|!=|>|<|=)(.*)$`)
)

// @en query condition struct
//
// @zh 查询条件结构
type QueryCondition struct {
	// @en field
	// @zh 字段
	Field string `json:"field"`
	// @en operator
	// @zh 操作符
	Operator string `json:"operator"`
	// @en value
	// @zh 值
	Value any `json:"value"`
	// @en is or
	// @zh 是否是或条件
	IsOr bool `json:"is_or"`
}

// @en page params struct
//
// @zh 分页参数结构
type PageParams struct {
	// @en page no
	// @zh 当前页码
	PageNo int `json:"page_no,string"` // 当前页码
	// @en page size
	// @zh 每页条数
	PageSize int `json:"page_size,string"` // 每页条数
	// @en filter
	// @zh 过滤条件
	Filter map[string]any `json:"filter"` // 过滤条件
	// @en order by
	// @zh 排序字段
	OrderBy []string `json:"order_by"` // 排序字段
	// @en group by
	// @zh 分组字段
	GroupBy []string `json:"group_by"` // 分组字段
	// @en having
	// @zh 分组条件
	Having map[string]any `json:"having"` // 分组条件
}

// @en page result struct
//
// @zh 分页结果结构
type PageResult[T any] struct {
	// @en count
	// @zh 总条数
	Count int64 `json:"count"` // 总条数
	// @en list
	// @zh 列表
	List []T `json:"list"` // 列表
	// @en page no
	// @zh 当前页码
	PageNo int `json:"page_no"` // 当前页码
	// @en page size
	// @zh 每页条数
	PageSize int `json:"page_size"` // 每页条数
	// @en pages
	// @zh 总页数
	Pages int `json:"pages"` // 总页数
}

// @en transaction config struct
//
// @zh 事务配置结构
type TxConfig struct {
	// @en isolation level
	// @zh 隔离级别
	IsolationLevel string
	// @en timeout
	// @zh 超时时间
	Timeout time.Duration
}

// @en query builder struct
//
// @zh 查询构建器结构
type QueryBuilder[T any] struct {
	db         *gorm.DB
	model      *T
	conditions []QueryCondition
	orderBy    []string
	groupBy    []string
	having     map[string]any
	limit      int
	offset     int
}

// @en join params struct
//
// @zh 联合查询参数结构
type JoinParams struct {
	// @en models
	// @zh 需要联合查询的模型切片，顺序决定了主表和关联表
	Models []any `json:"models"`
	// @en query conditions
	// @zh 查询条件，如果是多表查询，则需要指定表名，如: "table_name.id": 1
	QueryConditions map[string]any `json:"query_conditions"`
	// @en join conditions
	// @zh 连接条件，条件与除去主表外的models对应，如: Models: [User, LoginLog]，那么连接条件就则是["User.account = LoginLog.account"]
	JoinConditions []string `json:"join_conditions"`
	// @en select fields
	// @zh 查询字段，如: []string{"table1_name.field1", "table2_name.field2"}
	SelectFields []string `json:"select_fields"`
	// @en page params
	// @zh 分页参数，如果需要分页查询，则需要指定
	PageParams *PageParams `json:"page_params"`
}

// @en Base model struct. Contains ID, created time, updated time, and deleted time.
//
// @zh 基础模型结构，包含ID、创建时间、更新时间和删除时间
type BaseModel struct {
	ID        string `gorm:"primarykey" json:"id"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// @en before create hook, if ID is empty, generate snowflake ID
//
// @zh 创建前钩子，如果ID为空，则生成雪花ID
func (b *BaseModel) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = GenerateSnowflakeIDSafe()
	}
	return nil
}
