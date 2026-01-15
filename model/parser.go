package model

import (
	"fmt"
	"reflect"
	"regexp"
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
func formatFieldName(field string, tableNameMap map[string]string) string {
	// 如果包含点号，则分别包装表名和字段名
	if strings.Contains(field, ".") {
		parts := strings.Split(field, ".")
		if len(parts) == 2 {
			modelName := strings.TrimSpace(parts[0])
			fieldName := strings.TrimSpace(parts[1])
			// 从 tableNameMap 获取实际表名
			tableName := getTableNameFromMap(modelName, tableNameMap)
			return fmt.Sprintf("`%s`.`%s`", tableName, fieldName)
		}
		// 如果包含多个点号，只处理第一个点号（表名.字段名）
		firstDot := strings.Index(field, ".")
		modelName := strings.TrimSpace(field[:firstDot])
		fieldName := strings.TrimSpace(field[firstDot+1:])
		// 从 tableNameMap 获取实际表名
		tableName := getTableNameFromMap(modelName, tableNameMap)
		return fmt.Sprintf("`%s`.`%s`", tableName, fieldName)
	}
	// 不包含点号，直接包装
	return "`" + field + "`"
}

// 提取字段部分（去除 AS 别名）
func extractFieldPart(field string) string {
	field = strings.TrimSpace(field)
	upperField := strings.ToUpper(field)
	// 查找 " AS " 的位置（前后都有空格）
	if idx := strings.Index(upperField, " AS "); idx != -1 {
		return strings.TrimSpace(field[:idx])
	}
	return field
}

// 格式化查询字段
func formatSelectField(field string, tableNameMap map[string]string, tablePrefix string) string {
	field = strings.TrimSpace(field)

	// 检查是否包含 AS 别名（不区分大小写）
	asIndex := -1
	upperField := strings.ToUpper(field)
	// 查找 " AS " 的位置（前后都有空格）
	if idx := strings.Index(upperField, " AS "); idx != -1 {
		asIndex = idx
	}

	var fieldPart, aliasPart string
	if asIndex != -1 {
		// 分离字段部分和别名部分
		fieldPart = strings.TrimSpace(field[:asIndex])
		aliasPart = strings.TrimSpace(field[asIndex+4:]) // " AS " 长度为 4
	} else {
		fieldPart = field
		aliasPart = ""
	}

	// 格式化字段部分
	var formattedField string
	if strings.Contains(fieldPart, ".") {
		// 处理 table.field 或 table.* 格式
		parts := strings.Split(fieldPart, ".")
		if len(parts) == 2 {
			modelName := strings.TrimSpace(parts[0])
			fieldName := strings.TrimSpace(parts[1])
			// 从 tableNameMap 获取实际表名
			tableName := getTableNameFromMap(modelName, tableNameMap)
			// 如果是 table.* 格式
			if fieldName == "*" {
				formattedField = tablePrefix + tableName + ".*"
			} else {
				// 如果是 table.field 格式，用反引号包装
				formattedField = fmt.Sprintf("`%s`.`%s`", tablePrefix+tableName, utils.CamelToSnake(fieldName))
			}
		} else {
			// 如果包含多个点号，只处理第一个点号
			firstDot := strings.Index(fieldPart, ".")
			modelName := strings.TrimSpace(fieldPart[:firstDot])
			fieldName := strings.TrimSpace(fieldPart[firstDot+1:])
			// 从 tableNameMap 获取实际表名
			tableName := getTableNameFromMap(modelName, tableNameMap)
			if fieldName == "*" {
				formattedField = tablePrefix + tableName + ".*"
			} else {
				// 用反引号包装表名和字段名
				formattedField = fmt.Sprintf("`%s`.`%s`", tablePrefix+tableName, utils.CamelToSnake(fieldName))
			}
		}
	} else {
		// 不包含点号，直接转换并添加前缀，用反引号包装
		formattedField = fmt.Sprintf("`%s`", tablePrefix+utils.CamelToSnake(fieldPart))
	}

	// 如果有别名，添加 AS 别名
	if aliasPart != "" {
		// 别名需要用反引号包装，以支持特殊字符
		return formattedField + " AS `" + aliasPart + "`"
	}

	return formattedField
}

// 构建基础查询
func buildBaseQuery[T any](queryConditions map[string]any, dbName ...string) (*gorm.DB, error) {
	model, err := parseModelWithCache[T](dbName...)
	if err != nil {
		return nil, err
	}

	dbClient := getDBClient(dbName...)
	// 单表查询不需要表名映射，传递 nil
	query, err := parseQueryConditions(queryConditions, nil, dbClient.db)
	if err != nil {
		return nil, err
	}

	return query.Model(model), nil
}

// 解析单个条件值，返回操作符和值
func parseSingleCondition(value any) (string, any, error) {
	valueStr := fmt.Sprintf("%v", value)

	if strings.Contains(valueStr, "%") {
		return "LIKE", valueStr, nil
	} else if matches := operatorRegex.FindStringSubmatch(valueStr); len(matches) == 3 {
		operator := matches[1]
		val := strings.TrimSpace(matches[2])
		// 处理__EMPTY__标记
		switch val {
		case "__EMPTY__":
			switch operator {
			case "=":
				// 查询空值：IS NULL OR = ''
				return "IS_NULL_OR_EMPTY", "", nil
			case "!=", "<>":
				// 查询非空值：IS NOT NULL AND <> ''
				return "IS_NOT_NULL_AND_NOT_EMPTY", "", nil
			}
		case "":
			return "", nil, fmt.Errorf("operator format error: %s", valueStr)
		}
		return operator, val, nil
	} else {
		return "=", value, nil
	}
}

// 应用单个条件到查询
func applyCondition(db *gorm.DB, fieldName string, operator string, value any) *gorm.DB {
	switch operator {
	case "IS_NULL_OR_EMPTY":
		// 查询空值：IS NULL OR = ''
		return db.Where(fmt.Sprintf("(%s IS NULL OR %s = '')", fieldName, fieldName))
	case "IS_NOT_NULL_AND_NOT_EMPTY":
		// 查询非空值：IS NOT NULL AND <> ''
		return db.Where(fmt.Sprintf("(%s IS NOT NULL AND %s <> '')", fieldName, fieldName))
	case "IN":
		// 处理 IN 操作符（数组值）
		if values, ok := value.([]any); ok && len(values) > 0 {
			placeholders := strings.Repeat("?,", len(values))
			placeholders = placeholders[:len(placeholders)-1] // 移除最后一个逗号
			queryStr := fmt.Sprintf("%s IN (%s)", fieldName, placeholders)
			return db.Where(queryStr, values...)
		}
		return db
	default:
		queryStr := fmt.Sprintf("%s %s ?", fieldName, operator)
		return db.Where(queryStr, value)
	}
}

// 查询条件解析
func parseQueryConditions(queryConditions map[string]any, tableNameMap map[string]string, query ...*gorm.DB) (*gorm.DB, error) {
	var db *gorm.DB
	if len(query) > 0 {
		db = query[0]
	} else {
		db = getDBClient().db
	}

	if len(queryConditions) == 0 {
		return db, nil
	}

	for key, value := range queryConditions {
		// 检测值是否为数组或切片
		valueType := reflect.TypeOf(value)
		valueKind := valueType.Kind()
		isArrayOrSlice := valueKind == reflect.Array || valueKind == reflect.Slice

		// 所有格式都必须是数组/切片
		if !isArrayOrSlice {
			return nil, fmt.Errorf("field [%s]: value must be an array or slice, got %T", key, value)
		}

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

		// 解析字段名和后缀
		var fieldName string
		var conditionType string // "OR", "AND", "IN", "DEFAULT"

		if strings.HasSuffix(key, "::__OR__") {
			fieldName = utils.CamelToSnake(strings.TrimSuffix(key, "::__OR__"))
			conditionType = "OR"
		} else if strings.HasSuffix(key, "::__AND__") {
			fieldName = utils.CamelToSnake(strings.TrimSuffix(key, "::__AND__"))
			conditionType = "AND"
		} else if strings.HasSuffix(key, "::__IN__") {
			fieldName = utils.CamelToSnake(strings.TrimSuffix(key, "::__IN__"))
			conditionType = "IN"
		} else {
			fieldName = utils.CamelToSnake(key)
			conditionType = "DEFAULT"
		}

		// 验证字段名
		if !isValidFieldName(fieldName) {
			return nil, fmt.Errorf("invalid field name: %s", fieldName)
		}

		formattedFieldName := formatFieldName(fieldName, tableNameMap)

		// 根据条件类型处理
		switch conditionType {
		case "IN":
			// 直接使用数组作为 IN 的值
			placeholders := strings.Repeat("?,", len(values))
			placeholders = placeholders[:len(placeholders)-1] // 移除最后一个逗号
			queryStr := fmt.Sprintf("%s IN (%s)", formattedFieldName, placeholders)
			db = db.Where(queryStr, values...)

		case "OR":
			// 遍历数组，每个元素解析为条件，用 OR 连接
			orQueries := make([]string, 0, len(values))
			orArgs := make([]any, 0)

			for _, val := range values {
				operator, parsedValue, err := parseSingleCondition(val)
				if err != nil {
					return nil, fmt.Errorf("field [%s] OR condition parse error: %w", fieldName, err)
				}

				switch operator {
				case "IS_NULL_OR_EMPTY":
					orQueries = append(orQueries, fmt.Sprintf("(%s IS NULL OR %s = '')", formattedFieldName, formattedFieldName))
				case "IS_NOT_NULL_AND_NOT_EMPTY":
					orQueries = append(orQueries, fmt.Sprintf("(%s IS NOT NULL AND %s <> '')", formattedFieldName, formattedFieldName))
				case "IN":
					if inValues, ok := parsedValue.([]any); ok && len(inValues) > 0 {
						placeholders := strings.Repeat("?,", len(inValues))
						placeholders = placeholders[:len(placeholders)-1]
						queryStr := fmt.Sprintf("%s IN (%s)", formattedFieldName, placeholders)
						orQueries = append(orQueries, queryStr)
						orArgs = append(orArgs, inValues...)
					}
				default:
					queryStr := fmt.Sprintf("%s %s ?", formattedFieldName, operator)
					orQueries = append(orQueries, queryStr)
					orArgs = append(orArgs, parsedValue)
				}
			}

			if len(orQueries) > 0 {
				orClause := "(" + strings.Join(orQueries, " OR ") + ")"
				db = db.Where(orClause, orArgs...)
			}

		case "AND", "DEFAULT":
			// 遍历数组，每个元素解析为条件，用 AND 连接（默认也是 AND）
			for _, val := range values {
				operator, parsedValue, err := parseSingleCondition(val)
				if err != nil {
					return nil, fmt.Errorf("field [%s] AND condition parse error: %w", fieldName, err)
				}
				db = applyCondition(db, formattedFieldName, operator, parsedValue)
			}
		}
	}

	return db, nil
}

// 解析连接条件
func parseJoinConditions(joinConditions []string, tableNameMap map[string]string, dbName ...string) []JoinCondition {
	result := make([]JoinCondition, 0, len(joinConditions))
	// 获取表前缀
	tablePrefix := getDBClient(dbName...).prefix

	// 支持的连接类型
	joinTypes := map[string]bool{
		"LEFT":  true,
		"RIGHT": true,
		"INNER": true,
		"OUTER": true,
	}

	for _, condition := range joinConditions {
		condition = strings.TrimSpace(condition)
		if condition == "" {
			continue
		}

		joinCond := JoinCondition{
			JoinType: "INNER", // 默认使用 INNER JOIN
		}

		// 检查开头是否有连接类型（必须在条件最开始）
		parts := strings.Fields(condition)
		if len(parts) > 0 {
			upperFirst := strings.ToUpper(parts[0])
			if joinTypes[upperFirst] {
				joinCond.JoinType = upperFirst
				// 移除连接类型，重新组合条件
				condition = strings.Join(parts[1:], " ")
			}
		}

		// 解析多个条件（支持 AND/OR 连接）
		onClause := parseJoinOnClause(condition, tableNameMap, tablePrefix)
		if onClause == "" {
			continue
		}

		joinCond.OnClause = onClause
		result = append(result, joinCond)
	}

	return result
}

// 解析 JOIN ON 子句，支持多个条件用 AND/OR 连接
func parseJoinOnClause(condition string, tableNameMap map[string]string, tablePrefix string) string {
	// 使用正则表达式分割 AND/OR，同时保留分隔符和原始大小写
	andOrRegex := regexp.MustCompile(`\s+(?i)(AND|OR)\s+`)

	// 找到所有 AND/OR 的位置
	matches := andOrRegex.FindAllStringSubmatchIndex(condition, -1)
	if len(matches) == 0 {
		// 单个条件
		formatted := parseSingleJoinCondition(condition, tableNameMap, tablePrefix)
		return formatted
	}

	// 分割条件，保留操作符
	var conditions []string
	var operators []string
	lastIndex := 0

	for _, match := range matches {
		// 添加条件部分
		condPart := strings.TrimSpace(condition[lastIndex:match[0]])
		if condPart != "" {
			conditions = append(conditions, condPart)
		}
		// 添加操作符（保持原始大小写）
		op := condition[match[2]:match[3]]
		operators = append(operators, strings.ToUpper(op))
		lastIndex = match[1]
	}
	// 添加最后一个条件
	lastCond := strings.TrimSpace(condition[lastIndex:])
	if lastCond != "" {
		conditions = append(conditions, lastCond)
	}

	// 解析每个条件并格式化
	formattedConditions := make([]string, 0, len(conditions))
	for _, cond := range conditions {
		formatted := parseSingleJoinCondition(cond, tableNameMap, tablePrefix)
		if formatted != "" {
			formattedConditions = append(formattedConditions, formatted)
		}
	}

	if len(formattedConditions) == 0 {
		return ""
	}

	if len(formattedConditions) == 1 {
		return formattedConditions[0]
	}

	// 重新组合条件，使用原始的操作符
	var result strings.Builder
	result.WriteString(formattedConditions[0])
	for i := 1; i < len(formattedConditions); i++ {
		if i-1 < len(operators) {
			result.WriteString(" " + operators[i-1] + " ")
		} else {
			result.WriteString(" AND ")
		}
		result.WriteString(formattedConditions[i])
	}

	return result.String()
}

// 解析单个连接条件
func parseSingleJoinCondition(condition string, tableNameMap map[string]string, tablePrefix string) string {
	condition = strings.TrimSpace(condition)
	if condition == "" {
		return ""
	}

	// 支持的比较操作符
	operators := []string{"<>", ">=", "<=", "!=", ">", "<", "="}

	var operator string
	var operatorIndex int = -1

	// 查找第一个匹配的操作符
	for _, op := range operators {
		idx := strings.Index(condition, op)
		if idx != -1 {
			operator = op
			operatorIndex = idx
			break
		}
	}

	if operatorIndex == -1 {
		// 没有找到操作符，返回空
		return ""
	}

	leftPart := strings.TrimSpace(condition[:operatorIndex])
	rightPart := strings.TrimSpace(condition[operatorIndex+len(operator):])

	// 解析左侧字段
	leftField := parseJoinField(leftPart, tableNameMap, tablePrefix)

	// 解析右侧（可能是字段或值）
	rightField := parseJoinFieldOrValue(rightPart, tableNameMap, tablePrefix)

	return fmt.Sprintf("%s %s %s", leftField, operator, rightField)
}

// 解析连接字段或值
func parseJoinFieldOrValue(part string, tableNameMap map[string]string, tablePrefix string) string {
	part = strings.TrimSpace(part)
	if part == "" {
		return ""
	}

	// 如果是字符串值（用单引号或双引号包围）
	if (strings.HasPrefix(part, "'") && strings.HasSuffix(part, "'")) ||
		(strings.HasPrefix(part, "\"") && strings.HasSuffix(part, "\"")) {
		// 直接返回，保持原样
		return part
	}

	// 如果是数字（可能是整数或浮点数）
	if isNumeric(part) {
		return part
	}

	// 否则当作字段处理（table.field 或 field）
	return parseJoinField(part, tableNameMap, tablePrefix)
}

// 判断字符串是否为数字
func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	// 检查是否为数字（包括负号、小数点）
	hasDot := false
	for i, r := range s {
		if i == 0 && r == '-' {
			continue
		}
		if r == '.' {
			if hasDot {
				return false
			}
			hasDot = true
			continue
		}
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// 解析连接字段
func parseJoinField(field string, tableNameMap map[string]string, tablePrefix string) string {
	field = strings.TrimSpace(field)
	// 如果包含点号，说明是 table.field 格式
	if strings.Contains(field, ".") {
		parts := strings.Split(field, ".")
		if len(parts) == 2 {
			modelName := strings.TrimSpace(parts[0])
			fieldName := utils.CamelToSnake(strings.TrimSpace(parts[1]))
			// 从 tableNameMap 获取实际表名
			tableName := getTableNameFromMap(modelName, tableNameMap)
			return fmt.Sprintf("`%s`.`%s`", tablePrefix+tableName, fieldName)
		}
		// 如果包含多个点号，只处理第一个点号
		firstDot := strings.Index(field, ".")
		modelName := strings.TrimSpace(field[:firstDot])
		fieldName := utils.CamelToSnake(strings.TrimSpace(field[firstDot+1:]))
		// 从 tableNameMap 获取实际表名
		tableName := getTableNameFromMap(modelName, tableNameMap)
		return fmt.Sprintf("`%s`.`%s`", tablePrefix+tableName, fieldName)
	}
	// 不包含点号，直接转换
	return fmt.Sprintf("`%s`", tablePrefix+utils.CamelToSnake(field))
}

// 从 tableNameMap 获取表名，如果找不到则使用默认命名规则
func getTableNameFromMap(modelName string, tableNameMap map[string]string) string {
	if tableNameMap == nil {
		// 如果没有映射，使用默认命名规则
		return utils.CamelToSnake(modelName)
	}
	// 先尝试直接查找（可能是模型类型名，如 "Notice"）
	if tableName, ok := tableNameMap[modelName]; ok {
		return tableName
	}
	// 尝试蛇形命名（如 "notice"）
	snakeName := utils.CamelToSnake(modelName)
	if tableName, ok := tableNameMap[snakeName]; ok {
		return tableName
	}
	// 如果都找不到，使用默认命名规则
	return snakeName
}
