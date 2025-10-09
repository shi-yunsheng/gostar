package gostar

import (
	"strings"

	"github.com/shi-yunsheng/gostar/logger"
)

// @en Log config
//
// @zh 日志配置
type logConfig struct {
	EnablePrint      bool   `yaml:"enable_print"`
	EnableSave       bool   `yaml:"enable_save"`
	SavePath         string `yaml:"save_path"`
	EnableAutoDelete bool   `yaml:"enable_auto_delete"`
	MaxSaveDays      string `yaml:"max_save_days"`
	MaxFileSize      string `yaml:"max_file_size"`
}

// @en init log
//
// @zh 初始化日志
func (g *goStar) initLog() {
	config := g.config.Log

	if !config.EnablePrint {
		logger.DisablePrint()
	}
	if strings.TrimSpace(config.MaxFileSize) != "" {
		logger.SetMaxSingleLogFileSize(config.MaxFileSize)
	}
	if config.EnableSave {
		logger.EnableSave()
	}
	if strings.TrimSpace(config.SavePath) != "" {
		logger.SetSavePath(config.SavePath)
	}
	if strings.TrimSpace(config.MaxSaveDays) != "" {
		logger.SetMaxSaveDays(config.MaxSaveDays)
	}
	if config.EnableAutoDelete {
		logger.EnableAutoDelete()
	}
}
