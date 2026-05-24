<div align="center">
  <h1>Norma Frame</h1>
  <p>基于 Gin 的生产级 Go Web 框架 · 开箱即用 · 零配置启动</p>
  <p>
    <a href="https://github.com/normastars/frame/releases/latest"><img alt="Release" src="https://img.shields.io/github/v/release/normastars/frame?logo=github&style=flat-square"></a>
    <a href="https://github.com/normastars/frame/actions?query=workflow%3A%22Lint+Test+Build%22"><img alt="CI" src="https://github.com/normastars/frame/workflows/Lint%20Test%20Build/badge.svg"></a>
  </p>
</div>

---

## 简介

Norma Frame 是一个**开箱即用**的 Go Web 框架，在 Gin 的基础上集成了企业级项目所需的基础设施：

- **MySQL** 连接管理（多库、延迟初始化、自动建表）
- **Redis** 连接管理（多实例）
- **日志**：自动携带 `trace_id`，链路追踪开箱即用
- **监控**：Prometheus 指标自动采集（请求耗时、业务错误码）
- **HTTP 客户端**：基于 `req/v3` 的内置 HTTP 请求能力
- **配置管理**：支持 JSON / YAML / 环境变量，热加载

### 一句话概括

> 写一个 Web 服务，只需要 `frame.New()` → 定义路由 → `app.Run()`，数据库、Redis、日志、监控全部就绪。

---

## 快速开始

```go
package main

import "github.com/normastars/frame"

func main() {
    app := frame.New()
    app.GET("/hello", Hello)
    app.Run()
}

func Hello(c *frame.Context) {
    db := c.GetDB("default")     // 获取 MySQL 连接（延迟初始化）
    c.Success(map[string]string{"msg": "hello, world"})
}
```

把配置文件放到 `./conf/default.json`（或通过环境变量 `CONFPATH` 指定路径），启动即可。

> 如果使用 Go 1.26+ 在 macOS 上运行，需要加上 `CGO_ENABLED=0`（框架不依赖 cgo）：
> ```bash
> CGO_ENABLED=0 go run main.go
> ```
> 或直接 `make dev`。

[配置参考 →](docs/1-config.md)

---

## 使用场景

### 1. 常规 Web 服务

```go
app := frame.New()
app.GET("/user/:id", GetUser)
app.POST("/user", CreateUser)
app.Run()
```

### 2. 定时任务 / 脚本

```go
ctx := frame.NewContextNoGin()
db := ctx.GetDB("default")
// ... 执行数据库操作
```

### 3. 单元测试（注入 Mock）

```go
func TestHandler(t *testing.T) {
    sqliteDB, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    app := frame.New(frame.WithMockDB(sqliteDB))
    // ... 测试逻辑
}
```

---

## 路由 & 中间件

```go
app := frame.New()
app.Use(MyMiddleware())              // 全局中间件
app.GET("/users", ListUsers)
app.POST("/users", CreateUser)
app.PUT("/users/:id", UpdateUser)
app.DELETE("/users/:id", DeleteUser)
app.PATCH("/users/:id", PatchUser)
app.HEAD("/health", HealthCheck)
```

Handler 签名同 Gin，但参数为 `*frame.Context`，提供了更丰富的响应方法：

```go
func MyHandler(c *frame.Context) {
    c.Success(data)                          // 成功响应
    c.Error("E0001", "something wrong")      // 业务错误
    c.HTTPError(http.StatusBadRequest, "E0002", "invalid params", err)
    c.HTTPListSuccess(data, total)           // 分页成功响应
}
```

---

## 核心功能

| 功能 | 说明 |
|------|------|
| `ctx.GetDB(name)` | 获取 MySQL 连接（支持多数据库实例） |
| `ctx.GetRedis(name)` | 获取 Redis 连接（支持多实例） |
| `ctx.DoHTTP()` | 发送 HTTP 请求（基于 `req/v3`，自动携带 trace_id） |
| `ctx.GetLogger()` | 获取带 trace_id 的日志器 |
| `ctx.GetTraceID()` | 获取当前请求的链路 ID |
| `ctx.GetConfig()` | 获取应用配置 |
| `RegisterTable()` + `TablesInit()` | 自动建表、数据初始化 |

### 链路追踪

每个请求自动生成 `trace_id`，贯穿所有日志和 HTTP 调用，故障排查一目了然：

```go
// 日志自动携带 trace_id
c.Infof("processing user %d", id)
c.Errorf("failed to query: %v", err)
```

### Prometheus 指标

默认采集以下指标：

- `http_request_duration_ms`：按 URL × 状态码 × 方法统计的请求耗时
- `http_business_code`：按 URL × 业务错误码统计的请求数

暴露于 `{metric_port}/metrics` 端点。

---

## 配置

配置文件默认路径 `./conf/default.json`，支持 JSON / YAML 格式。

```json5
{
  "httpserver": {
    "server_port": "8080",
    "metric_port": "8081"
  },
  "mysql": {
    "enable": true,
    "configs": [{
      "name": "default",
      "dsn": "user:pass@tcp(127.0.0.1:3306)/db?charset=utf8mb4&parseTime=True&loc=Local"
    }]
  },
  "redis": {
    "enable": true,
    "configs": [{
      "name": "default",
      "addr": "127.0.0.1:6379"
    }]
  }
}
```

[完整配置说明 →](docs/1-config.md)

---

## 安装

```bash
go get github.com/normastars/frame
```

要求 Go 1.20+。

---

## 项目结构

```
.
├── core/          # 核心抽象层（内部包，对外不可见）
├── cache/         # Redis 连接管理器
├── database/      # MySQL 连接管理器
├── logger/        # 日志封装
├── docs/          # 文档
└── example/       # 使用示例
```

> 使用者只需要 `import "github.com/normastars/frame"`，无需了解内部包。

---

## License

[MIT](LICENSE)
