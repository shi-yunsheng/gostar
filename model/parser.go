package model

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/shi-yunsheng/gostar/utils"

	"github.com/mitchellh/mapstructure"
	"gorm.io/gorm"
)

// 获取数据库客户端
func getDBClient(dbName ...string) *DBClient {
	if len(dbName) > 0 && dbName[0] != "" {
		return GetDB(dbName[0])
	}
	return GetDB()
}

// 模型解析并缓存
func parseModelWithCache[T any](dbName ...string) (*T, error) {
	var model *T
	t := reflect.TypeOf(model)

	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	// 尝试从缓存获取
	if cached, ok := modelCache.Load(t); ok {
		return cached.(*T), nil
	}
	// 根据名称获取模型
	modelName := utils.CamelToSnake(t.Name())
	dbClient := getDBClient(dbName...)
	if dbClient.models == nil {
		return nil, fmt.Errorf("no models registered, please call AutoMigrate first")
	}
	model, ok := dbClient.models[modelName].(*T)
	if !ok {
		return nil, fmt.Errorf("model [%s] not found", modelName)
	}
	// 存入缓存
	modelCache.Store(t, model)
	return model, nil
}

// 数据解析
func parseData[T any](data map[string]any) (T, error) {
	var model T

	if len(data) == 0 {
		return model, fmt.Errorf("data cannot be empty")
	}
	// 使用mapstructure进行转换
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

// 字段验证
func isValidFieldName(field string) bool {
	if len(field) == 0 || len(field) > 64 {
		return false
	}
	// 验证字段名格式
	for i, char := range field {
		if i == 0 {
			// 首字符必须是字母或下划线
			if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || char == '_') {
				return false
			}
		} else {
			// 其他字符可以是字母、数字、下划线、点号或星号（用于 table.*）
			if !((char >= 'a' && char <= 'z') ||
				(char >= 'A' && char <= 'Z') ||
				(char >= '0' && char <= '9') ||
				char == '_' || char == '.' || char == '*') {
				return false
			}
		}
	}
	return true
}

// 格式化字段名
func formatFieldName(field string) string {
	// 如果包含点号，则分别包装表名和字段名
	if strings.Contains(field, ".") {
		parts := strings.Split(field, ".")
		if len(parts) == 2 {
			return "`" + parts[0] + "`.`" + parts[1] + "`"
		}
		// 如果包含多个点号，只处理第一个点号（表名.字段名）
		firstDot := strings.Index(field, ".")
		return "`" + field[:firstDot] + "`.`" + field[firstDot+1:] + "`"
	}
	// 不包含点号，直接包装
	return "`" + field + "`"
}

// 格式化查询字段
func formatSelectField(field string, tablePrefix string) string {
	// 如果包含点号，处理 table.field 或 table.* 格式
	if strings.Contains(field, ".") {
		parts := strings.Split(field, ".")
		if len(parts) == 2 {
			tableName := utils.CamelToSnake(parts[0])
			fieldName := parts[1]
			// 如果是 table.* 格式
			if fieldName == "*" {
				return tablePrefix + tableName + ".*"
			}
			// 如果是 table.field 格式
			return tablePrefix + tableName + "." + utils.CamelToSnake(fieldName)
		}
		// 如果包含多个点号，只处理第一个点号
		firstDot := strings.Index(field, ".")
		tableName := utils.CamelToSnake(field[:firstDot])
		fieldName := field[firstDot+1:]
		if fieldName == "*" {
			return tablePrefix + tableName + ".*"
		}
		return tablePrefix + tableName + "." + utils.CamelToSnake(fieldName)
	}
	// 不包含点号，直接转换并添加前缀
	return tablePrefix + utils.CamelToSnake(field)
}

// 构建基础查询
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

// 查询条件解析
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
	// 预分配切片容量
	conditions := make([]QueryCondition, 0, len(queryConditions))

	for key, value := range queryConditions {
		condition := QueryCondition{}
		// 检查是否为OR条件
		if strings.HasSuffix(key, "::__OR__") {
			condition.IsOr = true
			condition.Field = utils.CamelToSnake(strings.TrimSuffix(key, "::__OR__"))
		} else {
			condition.IsOr = false
			condition.Field = utils.CamelToSnake(key)
		}
		// 验证字段名
		if !isValidFieldName(condition.Field) {
			return nil, fmt.Errorf("invalid field name: %s", condition.Field)
		}
		// 检测值是否为数组或切片
		valueType := reflect.TypeOf(value)
		valueKind := valueType.Kind()
		isArrayOrSlice := valueKind == reflect.Array || valueKind == reflect.Slice

		// 如果是数组或切片，使用 IN 操作符
		if isArrayOrSlice {
			valueValue := reflect.ValueOf(value)
			if valueValue.Len() == 0 {
				// 空数组，跳过此条件
				continue
			}
			// 将数组/切片转换为 []any
			values := make([]any, valueValue.Len())
			for i := 0; i < valueValue.Len(); i++ {
				values[i] = valueValue.Index(i).Interface()
			}
			condition.Operator = "IN"
			condition.Value = values
		} else {
			// 解析操作符和值
			valueStr := fmt.Sprintf("%v", value)

			if strings.Contains(valueStr, "%") {
				condition.Operator = "LIKE"
				condition.Value = valueStr
			} else if matches := operatorRegex.FindStringSubmatch(valueStr); len(matches) == 3 {
				condition.Operator = matches[1]
				condition.Value = strings.TrimSpace(matches[2])
				// 处理__EMPTY__标记
				switch condition.Value {
				case "__EMPTY__":
					switch condition.Operator {
					case "=":
						// 查询空值：IS NULL OR = ''
						condition.Operator = "IS_NULL_OR_EMPTY"
						condition.Value = ""
					case "!=", "<>":
						// 查询非空值：IS NOT NULL AND <> ''
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
		}

		conditions = append(conditions, condition)
	}
	// 分离AND和OR条件
	andConditions := make([]QueryCondition, 0)
	orConditions := make([]QueryCondition, 0)

	for _, condition := range conditions {
		if condition.IsOr {
			orConditions = append(orConditions, condition)
		} else {
			andConditions = append(andConditions, condition)
		}
	}
	// 处理AND条件
	for _, condition := range andConditions {
		fieldName := formatFieldName(condition.Field)
		// 处理特殊操作符
		switch condition.Operator {
		case "IS_NULL_OR_EMPTY":
			// 查询空值：IS NULL OR = ''
			db = db.Where(fmt.Sprintf("(%s IS NULL OR %s = '')", fieldName, fieldName))
		case "IS_NOT_NULL_AND_NOT_EMPTY":
			// 查询非空值：IS NOT NULL AND <> ''
			db = db.Where(fmt.Sprintf("(%s IS NOT NULL AND %s <> '')", fieldName, fieldName))
		case "IN":
			// 处理 IN 操作符（数组值）
			if values, ok := condition.Value.([]any); ok && len(values) > 0 {
				placeholders := strings.Repeat("?,", len(values))
				placeholders = placeholders[:len(placeholders)-1] // 移除最后一个逗号
				queryStr := fmt.Sprintf("%s IN (%s)", fieldName, placeholders)
				db = db.Where(queryStr, values...)
			}
		default:
			queryStr := fmt.Sprintf("%s %s ?", fieldName, condition.Operator)
			db = db.Where(queryStr, condition.Value)
		}
	}
	// 处理OR条件
	if len(orConditions) > 0 {
		orQueries := make([]string, 0, len(orConditions))
		orArgs := make([]any, 0, len(orConditions))

		for _, condition := range orConditions {
			fieldName := formatFieldName(condition.Field)
			// 处理特殊操作符
			switch condition.Operator {
			case "IS_NULL_OR_EMPTY":
				// 查询空值：IS NULL OR = ''
				orQueries = append(orQueries, fmt.Sprintf("(%s IS NULL OR %s = '')", fieldName, fieldName))
			case "IS_NOT_NULL_AND_NOT_EMPTY":
				// 查询非空值：IS NOT NULL AND <> ''
				orQueries = append(orQueries, fmt.Sprintf("(%s IS NOT NULL AND %s <> '')", fieldName, fieldName))
			case "IN":
				// 处理 IN 操作符（数组值）
				if values, ok := condition.Value.([]any); ok && len(values) > 0 {
					placeholders := strings.Repeat("?,", len(values))
					placeholders = placeholders[:len(placeholders)-1] // 移除最后一个逗号
					queryStr := fmt.Sprintf("%s IN (%s)", fieldName, placeholders)
					orQueries = append(orQueries, queryStr)
					orArgs = append(orArgs, values...)
				}
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

// 解析连接条件
func parseJoinConditions(joinConditions []string, dbName ...string) []string {
	newJoinConditions := []string{}
	// 获取表前缀
	tablePrefix := getDBClient(dbName...).prefix

	for _, condition := range joinConditions {
		// 解析连接条件
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
