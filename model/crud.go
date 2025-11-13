package model

import (
	"fmt"

	"gorm.io/gorm"
)

// 插入方法
func Insert[T any](data map[string]any, dbName ...string) error {
	model, err := parseData[T](data)
	if err != nil {
		return err
	}

	db := getDBClient(dbName...)
	return db.db.Create(&model).Error
}

// 更新方法
func Update[T any](queryConditions map[string]any, data map[string]any, dbName ...string) error {
	model, err := parseData[T](data)
	if err != nil {
		return err
	}

	query, err := buildBaseQuery[T](queryConditions, dbName...)
	if err != nil {
		return err
	}

	return query.Updates(model).Error
}

// 删除方法
func Delete[T any](queryConditions map[string]any, isHardDelete bool, dbName ...string) error {
	query, err := buildBaseQuery[T](queryConditions, dbName...)
	if err != nil {
		return err
	}

	var model T
	if isHardDelete {
		err = query.Unscoped().Delete(&model).Error
	} else {
		err = query.Delete(&model).Error
	}

	return err
}

// 查询方法
func Query[T any](queryConditions map[string]any, dbName ...string) ([]T, error) {
	query, err := buildBaseQuery[T](queryConditions, dbName...)
	if err != nil {
		return nil, err
	}

	var result []T
	err = query.Find(&result).Error

	return result, err
}

// 批量插入
func BatchInsert[T any](data []map[string]any, dbName ...string) error {
	if len(data) == 0 {
		return fmt.Errorf("batch insert data cannot be empty")
	}
	// 默认批次大小
	size := 100
	// 预分配切片容量
	models := make([]T, 0, len(data))
	for i, item := range data {
		model, err := parseData[T](item)
		if err != nil {
			return fmt.Errorf("failed to parse data at index %d: %w", i+1, err)
		}
		models = append(models, model)
	}
	// 分批插入
	db := getDBClient(dbName...)

	return db.db.CreateInBatches(&models, size).Error
}

// 查询第一条
func First[T any](queryConditions map[string]any, dbName ...string) (T, error) {
	var result T
	query, err := buildBaseQuery[T](queryConditions, dbName...)
	if err != nil {
		return result, err
	}

	err = query.First(&result).Error

	return result, err
}

// 计数方法
func Count[T any](queryConditions map[string]any, dbName ...string) (int64, error) {
	var query *gorm.DB
	var err error

	if len(queryConditions) > 0 {
		query, err = buildBaseQuery[T](queryConditions, dbName...)
		if err != nil {
			return 0, err
		}
	} else {
		model, e := parseModelWithCache[T](dbName...)
		if e != nil {
			return 0, e
		}
		db := getDBClient(dbName...)
		query = db.db.Model(model)
	}

	var count int64
	err = query.Count(&count).Error

	return count, err
}

// 软删除查询
func QueryWithDeleted[T any](queryConditions map[string]any, includeDeleted bool, dbName ...string) ([]T, error) {
	query, err := buildBaseQuery[T](queryConditions, dbName...)
	if err != nil {
		return nil, err
	}

	if includeDeleted {
		query = query.Unscoped()
	}

	var result []T
	err = query.Find(&result).Error

	return result, err
}
