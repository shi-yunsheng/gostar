package model

import (
	"gorm.io/gorm"
)

// 事务支持
func WithTransaction[T any](fn func(*gorm.DB) (T, error), txConfig *TxConfig, dbName ...string) (T, error) {
	var result T

	dbClient := getDBClient(dbName...)
	db := dbClient.db

	// 设置事务隔离级别
	if txConfig != nil && txConfig.IsolationLevel != "" {
		db = db.Set("gorm:isolation_level", txConfig.IsolationLevel)
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		var err error
		result, err = fn(tx)
		return err
	})

	return result, err
}
