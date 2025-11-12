package model

import (
	"fmt"
	"reflect"

	"github.com/shi-yunsheng/gostar/utils"
)

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
		if modelType.Kind() == reflect.Ptr {
			modelType = modelType.Elem()
		}
		modelName := utils.CamelToSnake(modelType.Name())
		registeredModel, ok := db.models[modelName]
		if !ok {
			return nil, fmt.Errorf("model [%s] not found", modelName)
		}
		// 设置主表
		if i == 0 {
			query = query.Model(registeredModel)
			continue
		}
		// 连接表，如果没有连接条件，则使用模型定义的外键
		sql := ""
		if len(joinConditions) > 0 {
			sql = fmt.Sprintf("JOIN %s ON %s", tablePrefix+modelName, joinConditions[i-1])
		} else {
			sql = tablePrefix + modelName
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

			params.SelectFields[i] = tablePrefix + utils.CamelToSnake(field)
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
