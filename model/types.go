package model

import (
	"regexp"
	"time"

	"gorm.io/gorm"
)

// 默认列表相关
const (
	defaultPageSize = 20
	defaultPageNo   = 1
	maxPageSize     = 1000
	defaultOrderBy  = "`updated_at` DESC, `id` DESC"
)

var (
	// 操作符正则
	operatorRegex = regexp.MustCompile(`^(<>|>=|<=|!=|>|<|=)(.*)$`)
)

// 查询条件结构
type QueryCondition struct {
	Field    string `json:"field"`    // 字段
	Operator string `json:"operator"` // 操作符
	Value    any    `json:"value"`    // 值
	IsOr     bool   `json:"is_or"`    // 是否是或条件
}

// 分页参数结构
type PageParams struct {
	PageNo   int            `json:"page_no"`   // 当前页码
	PageSize int            `json:"page_size"` // 每页条数
	Filter   map[string]any `json:"filter"`    // 过滤条件
	OrderBy  []string       `json:"order_by"`  // 排序字段
	GroupBy  []string       `json:"group_by"`  // 分组字段
	Having   map[string]any `json:"having"`    // 分组条件
}

// 分页结果结构
type PageResult[T any] struct {
	Count    int64 `json:"count"`     // 总条数
	List     []T   `json:"list"`      // 列表
	PageNo   int   `json:"page_no"`   // 当前页码
	PageSize int   `json:"page_size"` // 每页条数
	Pages    int   `json:"pages"`     // 总页数
}

// 事务配置结构
type TxConfig struct {
	// 隔离级别
	IsolationLevel string
	// 超时时间
	Timeout time.Duration
}

// 查询构建器结构
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

// 连接条件结构
type JoinCondition struct {
	JoinType string // 连接类型: LEFT, RIGHT, INNER, OUTER，默认为 INNER
	OnClause string // ON 子句，如: "User.id = LoginLog.user_id"
}

// 联合查询参数结构
type JoinParams struct {
	// 需要联合查询的模型切片，顺序决定了主表和关联表
	Models []any `json:"models"`
	// 查询条件，如果是多表查询，则需要指定表名，如: "table_name.id": 1
	QueryConditions map[string]any `json:"query_conditions"`
	// JoinConditions 定义表连接条件。
	//
	// 格式说明：
	//   - 每个连接条件对应 Models 中除第一个（主表）外的其他表
	//   - 基本格式: "表名.字段 = 表名.字段" 或 "连接类型 表名.字段 = 表名.字段"
	//   - 支持连接类型: LEFT, RIGHT, INNER, OUTER（必须在条件最开始，不写则默认为 INNER）
	//   - 支持多个条件，使用"AND", "OR"连接
	//
	// 示例：
	//   Models: [User, LoginLog, Profile, Department, Order]
	//   JoinConditions: [
	//     "LEFT User.id = LoginLog.user_id",      // 左连接 LoginLog
	//     "RIGHT User.id = Profile.user_id",      // 右连接 Profile
	//     "INNER User.dept_id = Department.id",   // 内连接 Department
	//     "User.id = Order.user_id AND User.name = '张三'",               // 默认 INNER 连接
	//   ]
	//
	// 注意：如未指定连接类型，默认使用 INNER JOIN
	JoinConditions []string `json:"join_conditions"`
	// 查询字段，如: []string{"table1_name.field1", "table2_name.field2"}
	// 也可以是: []string{"table1_name.*", "table2_name.field2"}，表示查询table1表的所有字段和table2表的field2字段
	// 也可以是: []string{"table1_name.field1 AS field1_alias", "table2_name.field2 AS field2_alias"}，表示查询table1表的field1字段并别名为field1_alias，查询table2表的field2字段并别名为field2_alias
	SelectFields []string `json:"select_fields"`
	// 分页参数，如果需要分页查询，则需要指定
	PageParams *PageParams `json:"page_params"`
}

// 基础模型结构，包含ID、创建时间、更新时间和删除时间
type BaseModel struct {
	ID        string `gorm:"primarykey" json:"id"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// 创建前钩子，如果ID为空，则生成雪花ID
func (b *BaseModel) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = GenerateSnowflakeIDSafe()
	}
	return nil
}
