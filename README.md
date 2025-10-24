# GoStar

GoStar is a lightweight, feature-rich Go web framework designed to provide clean APIs and powerful functionality to help developers quickly build web applications.

## Features

- 🚀 **Simple & Easy** - Intuitive API design for quick start
- 🛣️ **Flexible Routing** - RESTful routing, route groups, and middleware support
- 🔌 **Powerful Middleware** - Built-in CORS, logging, error handling, rate limiting, and more
- 💾 **Multi-Database Support** - MySQL, PostgreSQL, SQLite, and MongoDB
- 📦 **Redis Integration** - Built-in Redis support for caching and session management
- 📝 **Advanced Logging** - Colorful console output, file saving, and auto-archiving
- 🔒 **Request Validation** - Built-in parameter parsing and validation
- 📤 **File Upload** - Simple file upload handling
- 🌐 **WebSocket Support** - Easy real-time communication
- 🎨 **Static File Serving** - Support for static files and SPA hosting
- ⚡ **High Performance** - Based on standard `net/http` library

## Installation

```bash
go get -u github.com/shi-yunsheng/gostar
```

## Quick Start

```go
package main

import (
    "github.com/shi-yunsheng/gostar"
)

func main() {
    // Create GoStar instance
    app := gostar.New()
    
    // Start server
    app.Run()
}
```

## Core Features

### Configuration Management

GoStar uses YAML format configuration file (`config.yaml`). A default config file will be generated automatically on first run.

Configuration options include:
- **debug** - Debug mode
- **bind** - Server bind address and port
- **allowed_origins** - CORS allowed origins
- **log** - Logging configuration (console output, file saving, auto-cleanup, etc.)
- **timezone** - Timezone setting
- **lang** - Language setting
- **database** - Database configuration (supports multiple database connections)
- **redis** - Redis configuration (supports multiple instance connections)

### Routing System

GoStar provides a flexible routing system with support for:

- RESTful style routes
- Path parameters (e.g., `/user/:id`)
- Query parameters
- Route groups
- Route-level middleware

### Middleware

Built-in middleware:
- **CORS Middleware** - Cross-origin resource sharing support
- **Logging Middleware** - Automatic request logging
- **Error Handling Middleware** - Unified error handling
- **Rate Limiting Middleware** - API access rate control

### Database ORM

Based on GORM and MongoDB driver, providing unified database operation interface:

- Support for MySQL, PostgreSQL, SQLite, MongoDB
- Auto database migration
- Query builder
- Transaction support
- Association queries
- Pagination support

### Redis

Built-in Redis support with convenient cache operations:

- Multiple Redis instance management
- Key-value operations
- Expiration time setting
- Key prefix support

### Logging System

Feature-rich logging system:

- Multiple log levels (Debug, Info, Warning, Error, Fatal)
- Colorful console output
- File saving
- Auto log archiving (by date)
- Auto cleanup of expired logs
- File size limitation

### Request Handling

Simplified request handling:

- Auto parameter parsing (path params, query params, form, JSON)
- Parameter validation
- File upload handling
- Cookie management
- Session support

### Response Handling

Convenient response methods:

- JSON response
- HTML response
- File download
- Redirect
- Error response

### WebSocket

Built-in WebSocket support:

- Simple connection management
- Message sending and receiving
- Connection pool management

### Static File Serving

- Static file hosting
- SPA application support
- File upload directory

### Utility Functions

Common utility functions provided:

- File operations
- IP address handling
- String processing
- Slice operations
- UUID generation
- Date/time processing

## Project Structure

```
gostar/
├── config.go           # Configuration management
├── gostar.go          # Core framework
├── date/              # Date/time handling
├── logger/            # Logging system
├── model/             # Database ORM
│   ├── db.go         # Database connection
│   ├── crud.go       # CRUD operations
│   ├── query_builder.go  # Query builder
│   ├── pagination.go  # Pagination
│   └── redis.go      # Redis support
├── router/            # Routing system
│   ├── router.go     # Router core
│   ├── route.go      # Route definitions
│   ├── handler/      # Request handlers
│   └── middleware/   # Middleware
└── utils/             # Utility functions
```

## Requirements

- Go 1.25.0 or higher

## Dependencies

Main dependencies:
- `gorm.io/gorm` - ORM framework
- `go.mongodb.org/mongo-driver` - MongoDB driver
- `github.com/go-redis/redis` - Redis client
- `github.com/gorilla/websocket` - WebSocket support
- `gopkg.in/yaml.v3` - YAML configuration parsing

## License

This project is licensed under the MIT License.

## Contributing

Issues and Pull Requests are welcome!

## Version

Current version: v1.0.5-beta

