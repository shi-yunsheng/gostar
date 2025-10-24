# GoStar

GoStar 是一个轻量级、功能丰富的 Go Web 框架，旨在提供简洁的 API 和强大的功能，帮助开发者快速构建 Web 应用。

## 特性

- 🚀 **简洁易用** - 简单直观的 API 设计，快速上手
- 🛣️ **灵活的路由系统** - 支持 RESTful 路由、路由分组、中间件
- 🔌 **强大的中间件** - 内置 CORS、日志、错误处理、速率限制等中间件
- 💾 **多数据库支持** - 支持 MySQL、PostgreSQL、SQLite、MongoDB
- 📦 **Redis 集成** - 内置 Redis 支持，便于缓存和会话管理
- 📝 **高级日志系统** - 支持彩色输出、文件保存、自动归档
- 🔒 **请求验证** - 内置参数解析和验证
- 📤 **文件上传** - 简单的文件上传处理
- 🌐 **WebSocket 支持** - 轻松实现实时通信
- 🎨 **静态文件服务** - 支持静态文件和 SPA 应用托管
- ⚡ **高性能** - 基于标准库 `net/http`，性能优异

## 安装

```bash
go get -u github.com/shi-yunsheng/gostar
```

## 快速开始

```go
package main

import (
    "github.com/shi-yunsheng/gostar"
)

func main() {
    // 创建 GoStar 实例
    app := gostar.New()
    
    // 启动服务器
    app.Run()
}
```

## 核心功能

### 配置管理

GoStar 使用 YAML 格式的配置文件 (`config.yaml`)。首次运行时会自动生成默认配置文件。

配置项包括：
- **debug** - 调试模式
- **bind** - 服务器绑定地址和端口
- **allowed_origins** - CORS 允许的来源
- **log** - 日志配置（控制台输出、文件保存、自动清理等）
- **timezone** - 时区设置
- **lang** - 语言设置
- **database** - 数据库配置（支持多数据库连接）
- **redis** - Redis 配置（支持多实例连接）

### 路由系统

GoStar 提供了灵活的路由系统，支持：

- RESTful 风格路由
- 路径参数（如 `/user/:id`）
- 查询参数
- 路由分组
- 路由级中间件

### 中间件

内置中间件：
- **CORS 中间件** - 跨域资源共享支持
- **日志中间件** - 自动记录请求日志
- **错误处理中间件** - 统一的错误处理
- **速率限制中间件** - API 访问速率控制

### 数据库 ORM

基于 GORM 和 MongoDB 驱动，提供统一的数据库操作接口：

- 支持 MySQL、PostgreSQL、SQLite、MongoDB
- 自动数据库迁移
- 查询构建器
- 事务支持
- 关联查询
- 分页支持

### Redis

内置 Redis 支持，提供便捷的缓存操作：

- 多 Redis 实例管理
- 键值操作
- 过期时间设置
- 键前缀支持

### 日志系统

功能丰富的日志系统：

- 多级别日志（Debug、Info、Warning、Error、Fatal）
- 彩色控制台输出
- 文件保存
- 自动日志归档（按日期）
- 自动清理过期日志
- 文件大小限制

### 请求处理

简化的请求处理：

- 自动参数解析（路径参数、查询参数、表单、JSON）
- 参数验证
- 文件上传处理
- Cookie 管理
- Session 支持

### 响应处理

便捷的响应方法：

- JSON 响应
- HTML 响应
- 文件下载
- 重定向
- 错误响应

### WebSocket

内置 WebSocket 支持：

- 简单的连接管理
- 消息发送和接收
- 连接池管理

### 静态文件服务

- 静态文件托管
- SPA 应用支持
- 文件上传目录

### 工具函数

提供常用工具函数：

- 文件操作
- IP 地址处理
- 字符串处理
- 切片操作
- UUID 生成
- 日期时间处理

## 项目结构

```
gostar/
├── config.go           # 配置管理
├── gostar.go          # 核心框架
├── date/              # 日期时间处理
├── logger/            # 日志系统
├── model/             # 数据库 ORM
│   ├── db.go         # 数据库连接
│   ├── crud.go       # CRUD 操作
│   ├── query_builder.go  # 查询构建器
│   ├── pagination.go  # 分页
│   └── redis.go      # Redis 支持
├── router/            # 路由系统
│   ├── router.go     # 路由核心
│   ├── route.go      # 路由定义
│   ├── handler/      # 请求处理器
│   └── middleware/   # 中间件
└── utils/             # 工具函数
```

## 系统要求

- Go 1.25.0 或更高版本

## 依赖

主要依赖：
- `gorm.io/gorm` - ORM 框架
- `go.mongodb.org/mongo-driver` - MongoDB 驱动
- `github.com/go-redis/redis` - Redis 客户端
- `github.com/gorilla/websocket` - WebSocket 支持
- `gopkg.in/yaml.v3` - YAML 配置解析

## 许可证

本项目采用 MIT 许可证。

## 贡献

欢迎提交 Issue 和 Pull Request！

## 版本

当前版本：v1.0.4-beta

