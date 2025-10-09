package model

import (
	"strings"
)

// @en create new query builder
//
// @zh 查询构建器模式
func NewQueryBuilder[T any](dbName ...string) (*QueryBuilder[T], error) {
	model, err := parseModelWithCache[T](dbName...)
	if err != nil {
		return nil, err
	}

	db := getDBClient(dbName...)
	return &QueryBuilder[T]{
		db:         db.db.Model(model),
		model:      model,
		conditions: make([]QueryCondition, 0),
		orderBy:    make([]string, 0),
		groupBy:    make([]string, 0),
		having:     make(map[string]any),
	}, nil
}

// @en add where conditions
//
// @zh 添加查询条件
func (qb *QueryBuilder[T]) Where(conditions map[string]any) *QueryBuilder[T] {
	query, err := parseQueryConditions(conditions, qb.db)
	if err == nil {
		qb.db = query
	}
	return qb
}

// @en add order by
//
// @zh 添加排序
func (qb *QueryBuilder[T]) OrderBy(fields ...string) *QueryBuilder[T] {
	qb.orderBy = append(qb.orderBy, fields...)
	return qb
}

// @en add group by
//
// @zh 添加分组
func (qb *QueryBuilder[T]) GroupBy(fields ...string) *QueryBuilder[T] {
	qb.groupBy = append(qb.groupBy, fields...)
	return qb
}

// @en set limit
//
// @zh 设置限制数量
func (qb *QueryBuilder[T]) Limit(limit int) *QueryBuilder[T] {
	qb.limit = limit
	return qb
}

// @en set offset
//
// @zh 设置偏移量
func (qb *QueryBuilder[T]) Offset(offset int) *QueryBuilder[T] {
	qb.offset = offset
	return qb
}

// @en execute query and return results
//
// @zh 执行查询并返回结果
func (qb *QueryBuilder[T]) Find() ([]T, error) {
	// @en apply order by
	// @zh 应用排序
	for _, order := range qb.orderBy {
		qb.db = qb.db.Order(order)
	}

	// @en apply group by
	// @zh 应用分组
	if len(qb.groupBy) > 0 {
		qb.db = qb.db.Group(strings.Join(qb.groupBy, ", "))
	}

	// @en apply limit and offset
	// @zh 应用限制和偏移量
	if qb.limit > 0 {
		qb.db = qb.db.Limit(qb.limit)
	}
	if qb.offset > 0 {
		qb.db = qb.db.Offset(qb.offset)
	}

	var result []T
	err := qb.db.Find(&result).Error
	return result, err
}

// @en execute query and return first result
//
// @zh 执行查询并返回第一条结果
func (qb *QueryBuilder[T]) First() (T, error) {
	var result T
	err := qb.db.First(&result).Error
	return result, err
}

// @en execute query and return count
//
// @zh 执行查询并返回数量
func (qb *QueryBuilder[T]) Count() (int64, error) {
	var count int64
	err := qb.db.Count(&count).Error
	return count, err
}
