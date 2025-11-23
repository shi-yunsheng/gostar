package model

import (
	"fmt"
	"reflect"

	"github.com/shi-yunsheng/gostar/utils"
)

// 联合查询
func JoinQuery[T any](params JoinParams, dbName ...string) (any, error) {
	if len(params.Models) == 0 {
		return nil, fmt.Errorf("models list cannot be empty")
	}

	if len(params.JoinConditions) > 0 && (len(params.Models) != len(params.JoinConditions)+1) {
		return nil, fmt.Errorf("join conditions should match the number of models excluding the main table")
	}

	db := getDBClient(dbName...)
	query := db.db
	tablePrefix := db.prefix
	joinConditions := parseJoinConditions(params.JoinConditions, db.tableNameMap, dbName...)
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
		tableName, ok := db.tableNameMap[modelName]
		if !ok {
			return nil, fmt.Errorf("table name for model [%s] not found in tableNameMap", modelName)
		}
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
			// 提取字段部分（去除 AS 别名）进行验证
			fieldPart := extractFieldPart(field)
			if !isValidFieldName(fieldPart) {
				return nil, fmt.Errorf("query field [%s] not found", field)
			}

			params.SelectFields[i] = formatSelectField(field, db.tableNameMap, tablePrefix)
		}
		query = query.Select(params.SelectFields)
	}
	// 解析查询条件
	query, err := parseQueryConditions(params.QueryConditions, db.tableNameMap, query)
	if err != nil {
		return nil, err
	}
	// 如果有分页参数，则应用分页
	if params.PageParams != nil {
		return parsePager[T](params.PageParams, query, db.tableNameMap)
	}

	var result []T
	err = query.Find(&result).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}
