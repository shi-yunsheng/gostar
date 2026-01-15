package model

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/shi-yunsheng/gostar/utils"
)

// UnionTableConfig 定义 UNION 查询中每个表的配置
type UnionTableConfig struct {
	Model           any            // 模型实例
	QueryConditions map[string]any // 查询条件，字段名使用数据库字段名（snake_case），如: {"source_sec_uid": "123", "created_at": []string{">=2024-01-01", "<2024-12-31"}}
	SelectFields    []string       // 需要查询的字段列表，使用模型字段名（驼峰），如: []string{"Nickname", "Avatar", "Level", "SourceSecUID"}
	ExtraFields     []any          // 额外的固定字段值，按照顺序添加到 SelectFields 之后，用于填充结果，如: []any{false} 表示 IsAnonymous 字段为 false
}

// UnionParams UNION 查询参数
type UnionParams struct {
	Tables     []UnionTableConfig // UNION 的多个表配置
	SelectAs   []string           // 最终 SELECT 的字段别名，顺序必须与 SelectFields 对应，如: []string{"nickname", "avatar", "level", "is_anonymous", "source_sec_uid"}
	OrderBy    []string           // 排序字段，使用 SelectAs 中定义的别名
	Limit      int                // 限制数量
	Offset     int                // 偏移量
	PageParams *PageParams        // 分页参数（如果设置了分页，Limit 和 Offset 将被忽略）
}

// UnionQuery 执行 UNION 查询，合并多个表的数据
// T 为返回结果的类型，通常是一个 DTO 结构
func UnionQuery[T any](params UnionParams, dbName ...string) (any, error) {
	if len(params.Tables) < 2 {
		return nil, fmt.Errorf("union query requires at least 2 tables")
	}

	if len(params.Tables) == 0 || len(params.Tables[0].SelectFields) == 0 {
		return nil, fmt.Errorf("select fields cannot be empty")
	}

	// 验证所有表的字段数量是否一致（包括 SelectFields 和 ExtraFields）
	firstTableTotalFields := len(params.Tables[0].SelectFields) + len(params.Tables[0].ExtraFields)
	for i, table := range params.Tables {
		totalFields := len(table.SelectFields) + len(table.ExtraFields)
		if totalFields != firstTableTotalFields {
			return nil, fmt.Errorf("table %d total fields count (%d = %d select + %d extra) doesn't match first table (%d)",
				i+1, totalFields, len(table.SelectFields), len(table.ExtraFields), firstTableTotalFields)
		}
	}

	db := getDBClient(dbName...)
	tablePrefix := db.prefix

	// 构建 UNION SQL
	var unionParts []string
	var args []any
	argIndex := 1

	for tableIdx, table := range params.Tables {
		// 获取模型信息
		modelType := reflect.TypeOf(table.Model)
		if modelType.Kind() == reflect.Pointer {
			modelType = modelType.Elem()
		}
		modelName := utils.CamelToSnake(modelType.Name())

		if _, ok := db.models[modelName]; !ok {
			return nil, fmt.Errorf("model [%s] not found", modelName)
		}

		tableName, ok := db.tableNameMap[modelName]
		if !ok {
			return nil, fmt.Errorf("table name for model [%s] not found in tableNameMap", modelName)
		}
		fullTableName := tablePrefix + tableName

		// 构建 SELECT 字段列表
		var selectFields []string
		for i, field := range table.SelectFields {
			// 将驼峰字段名转换为数据库字段名
			snakeField := utils.CamelToSnake(field)
			// 使用 AS 别名，确保与 DTO 字段匹配
			selectExpr := fmt.Sprintf("`%s`.`%s`", fullTableName, snakeField)
			// 如果提供了 SelectAs，使用它作为别名，否则使用字段名的小写形式
			if len(params.SelectAs) > i {
				selectExpr += fmt.Sprintf(" AS `%s`", params.SelectAs[i])
			} else {
				selectExpr += fmt.Sprintf(" AS `%s`", strings.ToLower(snakeField))
			}
			selectFields = append(selectFields, selectExpr)
		}

		// 添加额外字段（固定值）
		for i, fieldValue := range table.ExtraFields {
			extraIndex := len(table.SelectFields) + i
			var selectExpr string
			// 根据值的类型生成 SQL 字面量或使用占位符
			switch v := fieldValue.(type) {
			case bool:
				if v {
					selectExpr = "1"
				} else {
					selectExpr = "0"
				}
			case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
				selectExpr = fmt.Sprintf("%v", v)
			case float32, float64:
				selectExpr = fmt.Sprintf("%v", v)
			case string:
				// 字符串使用占位符，避免 SQL 注入
				selectExpr = "?"
				args = append(args, v)
				argIndex++
			case nil:
				selectExpr = "NULL"
			default:
				// 其他类型使用占位符
				selectExpr = "?"
				args = append(args, v)
				argIndex++
			}
			// 为额外字段添加别名（必须提供，否则 GORM 无法映射）
			if len(params.SelectAs) > extraIndex {
				selectExpr += fmt.Sprintf(" AS `%s`", params.SelectAs[extraIndex])
			} else {
				// 如果没有提供 SelectAs，使用默认别名（这种情况应该避免）
				selectExpr += fmt.Sprintf(" AS `extra_field_%d`", i)
			}
			selectFields = append(selectFields, selectExpr)
		}

		// 构建 SELECT 语句
		selectSQL := fmt.Sprintf("SELECT %s FROM `%s`", strings.Join(selectFields, ", "), fullTableName)

		// 添加 WHERE 条件
		if len(table.QueryConditions) > 0 {
			whereClause, whereArgs, err := buildUnionWhereClause(table.QueryConditions, modelName, db.tableNameMap, tablePrefix)
			if err != nil {
				return nil, fmt.Errorf("failed to build where clause for table %d: %w", tableIdx+1, err)
			}
			if whereClause != "" {
				selectSQL += " WHERE " + whereClause
				args = append(args, whereArgs...)
				argIndex += len(whereArgs)
			}
		}

		unionParts = append(unionParts, selectSQL)
	}

	// 组合 UNION SQL
	unionSQL := strings.Join(unionParts, " UNION ALL ")

	// 如果需要排序
	if len(params.OrderBy) > 0 {
		orderFields := make([]string, 0, len(params.OrderBy))
		for _, field := range params.OrderBy {
			orderFields = append(orderFields, fmt.Sprintf("`%s`", field))
		}
		unionSQL += " ORDER BY " + strings.Join(orderFields, ", ")
	}

	// 如果有分页参数，使用分页
	if params.PageParams != nil {
		if params.PageParams.PageSize <= 0 {
			params.PageParams.PageSize = defaultPageSize
		}
		if params.PageParams.PageNo <= 0 {
			params.PageParams.PageNo = defaultPageNo
		}
		offset := (params.PageParams.PageNo - 1) * params.PageParams.PageSize
		unionSQL += fmt.Sprintf(" LIMIT %d OFFSET %d", params.PageParams.PageSize, offset)
	} else {
		// 使用 Limit 和 Offset
		if params.Limit > 0 {
			unionSQL += fmt.Sprintf(" LIMIT %d", params.Limit)
		}
		if params.Offset > 0 {
			unionSQL += fmt.Sprintf(" OFFSET %d", params.Offset)
		}
	}

	// 执行查询
	var result []T = make([]T, 0)
	err := db.db.Raw(unionSQL, args...).Scan(&result).Error
	if err != nil {
		return nil, fmt.Errorf("union query failed: %w", err)
	}

	// 如果有分页参数，需要查询总数（用于分页结果）
	if params.PageParams != nil {
		// 构建计数查询（去掉 ORDER BY 和 LIMIT）
		countSQL := "SELECT COUNT(*) FROM (" + strings.Join(unionParts, " UNION ALL ") + ") AS union_result"
		var count int64
		err = db.db.Raw(countSQL, args...).Scan(&count).Error
		if err != nil {
			return nil, fmt.Errorf("union count query failed: %w", err)
		}

		pages := int(count) / params.PageParams.PageSize
		if int(count)%params.PageParams.PageSize > 0 {
			pages++
		}

		return PageResult[T]{
			Count:    count,
			List:     result,
			PageNo:   params.PageParams.PageNo,
			PageSize: params.PageParams.PageSize,
			Pages:    pages,
		}, nil
	}

	return result, nil
}

// buildUnionWhereClause 构建 UNION 查询的 WHERE 子句
func buildUnionWhereClause(conditions map[string]any, modelName string, tableNameMap map[string]string, tablePrefix string) (string, []any, error) {
	if len(conditions) == 0 {
		return "", nil, nil
	}

	var whereParts []string
	var args []any

	tableName := getTableNameFromMap(modelName, tableNameMap)
	fullTableName := tablePrefix + tableName

	for fieldName, value := range conditions {
		// fieldName 已经是 snake_case 格式，直接使用
		fieldExpr := fmt.Sprintf("`%s`.`%s`", fullTableName, fieldName)

		switch v := value.(type) {
		case []string:
			// 支持范围查询，如: []string{">=2024-01-01", "<2024-12-31"}
			for _, condition := range v {
				operator, val, err := parseSingleCondition(condition)
				if err != nil {
					return "", nil, err
				}
				switch operator {
				case "LIKE":
					whereParts = append(whereParts, fieldExpr+" LIKE ?")
					args = append(args, val)
				case "IS_NULL_OR_EMPTY":
					whereParts = append(whereParts, "("+fieldExpr+" IS NULL OR "+fieldExpr+" = '')")
				case "IS_NOT_NULL_AND_NOT_EMPTY":
					whereParts = append(whereParts, "("+fieldExpr+" IS NOT NULL AND "+fieldExpr+" <> '')")
				default:
					whereParts = append(whereParts, fieldExpr+" "+operator+" ?")
					args = append(args, val)
				}
			}
		default:
			whereParts = append(whereParts, fieldExpr+" = ?")
			args = append(args, value)
		}
	}

	return strings.Join(whereParts, " AND "), args, nil
}
