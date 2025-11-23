package model

import (
	"fmt"
	"strings"

	"github.com/shi-yunsheng/gostar/utils"

	"gorm.io/gorm"
)

// 分页解析
func parsePager[T any](params *PageParams, query *gorm.DB, tableNameMap map[string]string) (PageResult[T], error) {
	if params == nil {
		return PageResult[T]{}, nil
	}
	// 参数验证和默认值设置
	if params.PageNo <= 0 {
		params.PageNo = defaultPageNo
	}
	if params.PageSize <= 0 {
		params.PageSize = defaultPageSize
	}
	if params.PageSize > maxPageSize {
		params.PageSize = maxPageSize
	}
	// 解析过滤条件
	if len(params.Filter) > 0 {
		var err error
		// 传递 tableNameMap
		query, err = parseQueryConditions(params.Filter, tableNameMap, query)
		if err != nil {
			return PageResult[T]{}, err
		}
	}
	// 应用分组
	if len(params.GroupBy) > 0 {
		validGroupFields := make([]string, 0, len(params.GroupBy))
		for _, field := range params.GroupBy {
			if isValidFieldName(field) {
				// 格式化分组字段
				formattedField := formatGroupByField(field, tableNameMap)
				validGroupFields = append(validGroupFields, formattedField)
			}
		}
		if len(validGroupFields) > 0 {
			query = query.Group(strings.Join(validGroupFields, ", "))
		}
	}
	// 应用HAVING条件
	if len(params.Having) > 0 {
		// 传递 tableNameMap
		havingQuery, err := parseQueryConditions(params.Having, tableNameMap, query.Session(&gorm.Session{}))
		if err != nil {
			return PageResult[T]{}, fmt.Errorf("failed to parse HAVING conditions: %w", err)
		}
		// 提取HAVING条件的WHERE子句
		query = query.Having(havingQuery.Statement.SQL.String(), havingQuery.Statement.Vars...)
	}
	// 计算总记录数
	var totalCount int64
	countQuery := query.Session(&gorm.Session{})
	err := countQuery.Count(&totalCount).Error
	if err != nil {
		return PageResult[T]{}, fmt.Errorf("failed to count records: %w", err)
	}
	// 应用排序
	if len(params.OrderBy) > 0 {
		for _, orderField := range params.OrderBy {
			parts := strings.Fields(strings.TrimSpace(orderField))
			if len(parts) == 0 {
				continue
			}

			fieldName := parts[0]
			direction := "ASC"

			if len(parts) > 1 {
				dir := strings.ToUpper(parts[1])
				if dir == "DESC" || dir == "ASC" {
					direction = dir
				}
			}

			if isValidFieldName(fieldName) {
				// 格式化排序字段（支持 table.field 格式）
				formattedField := formatOrderByField(fieldName, tableNameMap)
				query = query.Order(formattedField + " " + direction)
			}
		}
	} else {
		query = query.Order(defaultOrderBy)
	}
	// 计算分页
	offset := (params.PageNo - 1) * params.PageSize
	pages := int((totalCount + int64(params.PageSize) - 1) / int64(params.PageSize))
	// 查询数据
	var result []T
	err = query.Limit(params.PageSize).Offset(offset).Find(&result).Error
	if err != nil {
		return PageResult[T]{}, fmt.Errorf("failed to query data: %w", err)
	}

	return PageResult[T]{
		Count:    totalCount,
		List:     result,
		PageNo:   params.PageNo,
		PageSize: params.PageSize,
		Pages:    pages,
	}, nil
}

// 格式化排序字段
func formatOrderByField(field string, tableNameMap map[string]string) string {
	field = strings.TrimSpace(field)
	// 如果包含点号，说明是 table.field 格式
	if strings.Contains(field, ".") {
		parts := strings.Split(field, ".")
		if len(parts) == 2 {
			modelName := strings.TrimSpace(parts[0])
			fieldName := strings.TrimSpace(parts[1])
			tableName := getTableNameFromMap(modelName, tableNameMap)
			// 格式化字段名
			formattedFieldName := utils.CamelToSnake(fieldName)
			return fmt.Sprintf("`%s`.`%s`", tableName, formattedFieldName)
		}
		// 如果包含多个点号，只处理第一个点号
		firstDot := strings.Index(field, ".")
		modelName := strings.TrimSpace(field[:firstDot])
		fieldName := strings.TrimSpace(field[firstDot+1:])
		tableName := getTableNameFromMap(modelName, tableNameMap)
		// 格式化字段名
		formattedFieldName := utils.CamelToSnake(fieldName)
		return fmt.Sprintf("`%s`.`%s`", tableName, formattedFieldName)
	}
	// 不包含点号，直接包装
	return "`" + field + "`"
}

// 格式化分组字段，支持 table.field 格式
func formatGroupByField(field string, tableNameMap map[string]string) string {
	field = strings.TrimSpace(field)
	// 如果包含点号，说明是 table.field 格式
	if strings.Contains(field, ".") {
		parts := strings.Split(field, ".")
		if len(parts) == 2 {
			modelName := strings.TrimSpace(parts[0])
			fieldName := strings.TrimSpace(parts[1])
			// 从 tableNameMap 获取实际表名
			tableName := getTableNameFromMap(modelName, tableNameMap)
			// 格式化字段名
			formattedFieldName := utils.CamelToSnake(fieldName)
			return fmt.Sprintf("`%s`.`%s`", tableName, formattedFieldName)
		}
		// 如果包含多个点号，只处理第一个点号
		firstDot := strings.Index(field, ".")
		modelName := strings.TrimSpace(field[:firstDot])
		fieldName := strings.TrimSpace(field[firstDot+1:])
		// 从 tableNameMap 获取实际表名
		tableName := getTableNameFromMap(modelName, tableNameMap)
		// 格式化字段名
		formattedFieldName := utils.CamelToSnake(fieldName)
		return fmt.Sprintf("`%s`.`%s`", tableName, formattedFieldName)
	}
	// 不包含点号，直接包装
	return "`" + field + "`"
}

// 分页查询
func Pagination[T any](params PageParams, dbName ...string) (PageResult[T], error) {
	model, err := parseModelWithCache[T](dbName...)
	if err != nil {
		return PageResult[T]{}, err
	}

	db := getDBClient(dbName...)
	query := db.db.Model(model)

	return parsePager[T](&params, query, nil)
}
