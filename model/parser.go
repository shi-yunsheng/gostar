package model

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/shi-yunsheng/gostar/utils"

	"github.com/mitchellh/mapstructure"
	"gorm.io/gorm"
)

// @en get database client
// @zh 获取数据库客户端
func getDBClient(dbName ...string) *DBClient {
	if len(dbName) > 0 && dbName[0] != "" {
		return GetDB(dbName[0])
	}
	return GetDB()
}

// @en parse model with cache
// @zh 模型解析并缓存
func parseModelWithCache[T any](dbName ...string) (*T, error) {
	var model *T
	t := reflect.TypeOf(model)

	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	// @en try to get model from cache
	// @zh 尝试从缓存获取
	if cached, ok := modelCache.Load(t); ok {
		return cached.(*T), nil
	}

	// @en get model from db by name
	// @zh 根据名称获取模型
	modelName := utils.CamelToSnake(t.Name())
	dbClient := getDBClient(dbName...)
	if dbClient.models == nil {
		return nil, fmt.Errorf("no models registered, please call AutoMigrate first")
	}
	model, ok := dbClient.models[modelName].(*T)
	if !ok {
		return nil, fmt.Errorf("model [%s] not found", modelName)
	}

	// @en store model in cache
	// @zh 存入缓存
	modelCache.Store(t, model)
	return model, nil
}

// @en parse data
//
// @zh 数据解析
func parseData[T any](data map[string]any) (T, error) {
	var model T

	if len(data) == 0 {
		return model, fmt.Errorf("data cannot be empty")
	}

	// @en use mapstructure for conversion
	// @zh 使用mapstructure进行转换
	config := &mapstructure.DecoderConfig{
		Result:           &model,
		WeaklyTypedInput: true,
		TagName:          "json",
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return model, fmt.Errorf("failed to create decoder: %w", err)
	}

	err = decoder.Decode(data)
	if err != nil {
		return model, fmt.Errorf("failed to parse data: %w", err)
	}

	return model, nil
}

// @en validate field name
//
// @zh 字段验证
func isValidFieldName(field string) bool {
	if len(field) == 0 || len(field) > 64 {
		return false
	}

	// @en validate field name format
	// @zh 验证字段名格式
	for i, char := range field {
		if i == 0 {
			// @en first character must be letter or underscore
			// @zh 首字符必须是字母或下划线
			if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || char == '_') {
				return false
			}
		} else {
			// @en other characters can be letters, digits, underscore or dot
			// @zh 其他字符可以是字母、数字、下划线或点号
			if !((char >= 'a' && char <= 'z') ||
				(char >= 'A' && char <= 'Z') ||
				(char >= '0' && char <= '9') ||
				char == '_' || char == '.') {
				return false
			}
		}
	}
	return true
}

// @en build base query
//
// @zh 构建基础查询
func buildBaseQuery[T any](queryConditions map[string]any, dbName ...string) (*gorm.DB, error) {
	model, err := parseModelWithCache[T](dbName...)
	if err != nil {
		return nil, err
	}

	dbClient := getDBClient(dbName...)
	query, err := parseQueryConditions(queryConditions, dbClient.db)
	if err != nil {
		return nil, err
	}

	return query.Model(model), nil
}

// @en parse query conditions
//
// @zh 查询条件解析
func parseQueryConditions(queryConditions map[string]any, query ...*gorm.DB) (*gorm.DB, error) {
	var db *gorm.DB
	if len(query) > 0 {
		db = query[0]
	} else {
		db = getDBClient().db
	}

	if len(queryConditions) == 0 {
		return db, nil
	}

	// @en pre-allocate slice capacity
	// @zh 预分配切片容量
	conditions := make([]QueryCondition, 0, len(queryConditions))

	for key, value := range queryConditions {
		condition := QueryCondition{}

		// @en check if it's an OR condition
		// @zh 检查是否为OR条件
		if strings.HasSuffix(key, "::__OR__") {
			condition.IsOr = true
			condition.Field = utils.CamelToSnake(strings.TrimSuffix(key, "::__OR__"))
		} else {
			condition.IsOr = false
			condition.Field = utils.CamelToSnake(key)
		}

		// @en validate field name
		// @zh 验证字段名
		if !isValidFieldName(condition.Field) {
			return nil, fmt.Errorf("invalid field name: %s", condition.Field)
		}

		// @en parse operator and value
		// @zh 解析操作符和值
		valueStr := fmt.Sprintf("%v", value)

		if strings.Contains(valueStr, "%") {
			condition.Operator = "LIKE"
			condition.Value = valueStr
		} else if matches := operatorRegex.FindStringSubmatch(valueStr); len(matches) == 3 {
			condition.Operator = matches[1]
			condition.Value = strings.TrimSpace(matches[2])

			// @en handle __EMPTY__ marker
			// @zh 处理__EMPTY__标记
			switch condition.Value {
			case "__EMPTY__":
				switch condition.Operator {
				case "=":
					// @en query empty values: IS NULL OR = ''
					// @zh 查询空值：IS NULL OR = ''
					condition.Operator = "IS_NULL_OR_EMPTY"
					condition.Value = ""
				case "!=", "<>":
					// @en query non-empty values: IS NOT NULL AND <> ''
					// @zh 查询非空值：IS NOT NULL AND <> ''
					condition.Operator = "IS_NOT_NULL_AND_NOT_EMPTY"
					condition.Value = ""
				}
			case "":
				return nil, fmt.Errorf("operator format error for field [%s]: %s", condition.Field, valueStr)
			}
		} else {
			condition.Operator = "="
			condition.Value = value
		}

		conditions = append(conditions, condition)
	}

	// @en separate AND and OR conditions
	// @zh 分离AND和OR条件
	andConditions := make([]QueryCondition, 0)
	orConditions := make([]QueryCondition, 0)

	for _, condition := range conditions {
		if condition.IsOr {
			orConditions = append(orConditions, condition)
		} else {
			andConditions = append(andConditions, condition)
		}
	}

	// @en process AND conditions
	// @zh 处理AND条件
	for _, condition := range andConditions {
		fieldName := "`" + condition.Field + "`"

		// @en handle special operators
		// @zh 处理特殊操作符
		switch condition.Operator {
		case "IS_NULL_OR_EMPTY":
			// @en query empty values: IS NULL OR = ''
			// @zh 查询空值：IS NULL OR = ''
			db = db.Where(fmt.Sprintf("(%s IS NULL OR %s = '')", fieldName, fieldName))
		case "IS_NOT_NULL_AND_NOT_EMPTY":
			// @en query non-empty values: IS NOT NULL AND <> ''
			// @zh 查询非空值：IS NOT NULL AND <> ''
			db = db.Where(fmt.Sprintf("(%s IS NOT NULL AND %s <> '')", fieldName, fieldName))
		default:
			queryStr := fmt.Sprintf("%s %s ?", fieldName, condition.Operator)
			db = db.Where(queryStr, condition.Value)
		}
	}

	// @en process OR conditions
	// @zh 处理OR条件
	if len(orConditions) > 0 {
		orQueries := make([]string, 0, len(orConditions))
		orArgs := make([]any, 0, len(orConditions))

		for _, condition := range orConditions {
			fieldName := "`" + condition.Field + "`"

			// @en handle special operators
			// @zh 处理特殊操作符
			switch condition.Operator {
			case "IS_NULL_OR_EMPTY":
				// @en query empty values: IS NULL OR = ''
				// @zh 查询空值：IS NULL OR = ''
				orQueries = append(orQueries, fmt.Sprintf("(%s IS NULL OR %s = '')", fieldName, fieldName))
			case "IS_NOT_NULL_AND_NOT_EMPTY":
				// @en query non-empty values: IS NOT NULL AND <> ''
				// @zh 查询非空值：IS NOT NULL AND <> ''
				orQueries = append(orQueries, fmt.Sprintf("(%s IS NOT NULL AND %s <> '')", fieldName, fieldName))
			default:
				queryStr := fmt.Sprintf("%s %s ?", fieldName, condition.Operator)
				orQueries = append(orQueries, queryStr)
				orArgs = append(orArgs, condition.Value)
			}
		}

		orClause := "(" + strings.Join(orQueries, " OR ") + ")"
		db = db.Where(orClause, orArgs...)
	}

	return db, nil
}

// @en parse join conditions
//
// @zh 解析连接条件
func parseJoinConditions(joinConditions []string, dbName ...string) []string {
	newJoinConditions := []string{}

	// @en get table prefix
	// @zh 获取表前缀
	tablePrefix := getDBClient(dbName...).prefix

	for _, condition := range joinConditions {
		// @en parse join condition
		// @zh 解析连接条件
		parts := strings.Split(condition, "=")
		if len(parts) != 2 {
			continue
		}
		leftField := tablePrefix + utils.CamelToSnake(strings.TrimSpace(parts[0]))
		rightField := tablePrefix + utils.CamelToSnake(strings.TrimSpace(parts[1]))

		newJoinConditions = append(newJoinConditions, fmt.Sprintf("%s = %s", leftField, rightField))
	}

	return newJoinConditions
}
