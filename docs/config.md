
### JSON 配置示例

```
{
	"project": "normastars",
	"log_level": "info",
	"log_mode": "json",
	"print_conf": true,
	"enable_metric": true,
	"env": "dev",
	"http_server": {
		"enable": true,
		"enable_cors": false,
		"disable_req_log": false,
		"configs": [{
			"name": "server",
			"port": ":8080"
		}, {
			"name": "metrics",
			"port": ":9090"
		}]
	},
	"http_client": {
		"disable_req_log": false,
		"enable_metric": true
	},
	"mysql": {
		"enable": true,
		"disable_req_log": true,
		"configs": [{
			"name": "user",
			"enable": true,
			"enable_auto_migrate": true,
			"host": "127.0.0.1:3309",
			"database": "user",
			"user": "root",
			"password": "root",
			"slow_threshold_sec": 3
		}]
	},
	"redis": {
		"enable": true,
		"disable_req_log": true,
		"configs": [{
			"name": "user",
			"enable": true,
			"host": "127.0.0.1:6380",
			"pool_size": 0,
			"password": "",
			"db": 0
		}]
	}
}
```

### 配置解释

| 字段名 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| project | string | normastars | 项目名称,必填 |
| log_level | string | info | 日志级别,支持:debug,info,warn,error|
| log_mode | string | json | 日志模式,json/text,默认:json |
| print_conf | bool | false | 是否控制台打印加载后的日志内容 |
| enable_metric | bool | false | 是否启用指标采集,默认:不采集 |
| env | string | dev | 环境名称,必填项 |
| http_server.enable | bool | false | 是否启动HTTP服务,默认不启动 |
| http_server.enable_cors | bool | false | 是否运行cors跨域, 默认不允许 |
| http_server.disable_req_log | bool | false | 是否禁用HTTP请求日志,默认启用 |
| http_server.configs | array | nil | HTTP服务配置项列表, 如果 http_server.enable 为true,此处不能为空 |
| http_client.disable_req_log | bool | false | 是否禁用请求HTTP请求日志,默认启用 |
| http_client.enable_metric | bool | false | 是否启用请求HTTP请求指标,默认禁用 |
| mysql.enable | bool | false | 是否启用MySQL数据库,默认不启用 |
| mysql.disable_req_log | bool | false | 是否禁用MySQL请求日志,默认打印 |
| mysql.configs | array | nil | MySQL数据库配置项列表 |
| redis.enable | bool | false | 是否启用Redis数据库,默认禁用 |
| redis.disable_req_log | bool | false | 是否禁用Redis请求日志,默认打印 |
| redis.configs | array | nil | Redis数据库配置项列表 |

### http_server.configs 字段

| 字段名 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| name* | string | 无 | HTTP服务名称, 当值是metric或metrics 是用于prometheus 指标暴露 |
| port* | string |  | HTTP服务监听端口, 指标格式 :8080 |

### mysql.configs 字段

| 字段名 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| name* | string |  | MySQL数据库名称 |
| enable | bool | false | 是否启用MySQL数据库 |
| enable_auto_migrate | bool | false | 启用后会自动创建库,默认不启用 |
| host | string | 127.0.0.1:3309 | MySQL数据库主机地址 |
| database | string | user | MySQL数据库名称 |
| user | string | root | MySQL数据库用户名 |
| password | string | root | MySQL数据库密码 |
| slow_threshold_sec | int | 3 | 慢查询阈值（秒） |

### redis.configs 字段

| 字段名 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| name | string |  | Redis数据库名称 |
| enable | bool | true | 是否启用Redis数据库 |
| host | string | 127.0.0.1:6379 | Redis数据库主机地址 |
| pool_size | int | 0 | Redis数据库连接池大小 |
| password | string |  | Redis数据库密码 |
| db | int | 0 | Redis数据库编号 |
