package model

import (
	"fmt"
	"reflect"

	"github.com/shi-yunsheng/gostar/utils"
)

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

// 联合查询
func JoinQuery(params JoinParams, dbName ...string) (any, error) {
	if len(params.Models) == 0 {
		return nil, fmt.Errorf("models list cannot be empty")
	}

	if len(params.JoinConditions) > 0 && (len(params.Models) != len(params.JoinConditions)+1) {
		return nil, fmt.Errorf("join conditions should match the number of models excluding the main table")
	}

	db := getDBClient(dbName...)
	query := db.db
	tablePrefix := db.prefix
	joinConditions := parseJoinConditions(params.JoinConditions, dbName...)
	// 将所有模型转换为结构体类型并获取表名
	for i, model := range params.Models {
		modelType := reflect.TypeOf(model)
		if modelType.Kind() == reflect.Pointer {
			modelType = modelType.Elem()
		}
		modelName := utils.CamelToSnake(modelType.Name())
		registeredModel, ok := db.models[modelName]
		if !ok {
			return nil, fmt.Errorf("model [%s] not found", modelName)
		}
		// 获取表名（优先使用 TableName() 方法）
		tableName := getTableName(registeredModel)
		// 设置主表
		if i == 0 {
			query = query.Model(registeredModel)
			continue
		}
		// 连接表，如果没有连接条件，则使用模型定义的外键
		var sql string
		if len(joinConditions) > 0 && i-1 < len(joinConditions) {
			joinCond := joinConditions[i-1]
			// 使用解析出的连接类型和 ON 子句
			sql = fmt.Sprintf("%s JOIN `%s` ON %s", joinCond.JoinType, tablePrefix+tableName, joinCond.OnClause)
		} else {
			// 没有连接条件时，默认使用 INNER JOIN
			sql = fmt.Sprintf("INNER JOIN `%s`", tablePrefix+tableName)
		}
		query = query.Joins(sql)
	}
	// 如果指定查询字段
	if len(params.SelectFields) > 0 {
		for i, field := range params.SelectFields {
			// 验证字段名
			if !isValidFieldName(field) {
				return nil, fmt.Errorf("query field [%s] not found", field)
			}

			params.SelectFields[i] = formatSelectField(field, tablePrefix)
		}
		query = query.Select(params.SelectFields)
	}
	// 解析查询条件
	query, err := parseQueryConditions(params.QueryConditions, query)
	if err != nil {
		return nil, err
	}
	// 如果有分页参数，则应用分页
	if params.PageParams != nil {
		return parsePager[map[string]any](params.PageParams, query)
	}

	var result []map[string]any
	err = query.Find(&result).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}
