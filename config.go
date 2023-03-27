package frame

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config project config
type Config struct {
	Project      string      `json:"project"`
	Level        string      `json:"level"`
	EnableMetric bool        `json:"enable_metric" yaml:"enable_metric" mapstructure:"enable_metric"`
	Env          string      `json:"env"`
	HTTPServer   HTTPServer  `json:"http_server" yaml:"http_server" mapstructure:"http_server"`
	Mysql        MySQLConfig `json:"mysql"`
	Redis        RedisConfig `json:"redis"`
}

// HTTPServer http config
type HTTPServer struct {
	Enable  bool               `json:"enable"`
	Configs []HTTPServerConfig `json:"configs"`
}

// HTTPServerConfig http server config item
type HTTPServerConfig struct {
	Name string `json:"name"`
	Port string `json:"port"`
}

// MySQLConfig mysql config
type MySQLConfig struct {
	Enable  bool              `json:"enable"`
	Configs []MySQLConfigItem `json:"configs"`
}

// MySQLConfigItem mysql config item
type MySQLConfigItem struct {
	Name              string `json:"name"`
	Default           bool   `json:"default"`
	Enable            bool   `json:"enable"`
	EnableAutoMigrate bool   `json:"enable_auto_migrate"  mapstructure:"enable_auto_migrate"`
	Host              string `json:"host"`
	Database          string `json:"database"`
	User              string `json:"user"`
	Password          string `json:"password"`
}

// RedisConfig redis config
type RedisConfig struct {
	Enable  bool              `json:"enable"`
	Configs []RedisConfigItem `json:"configs"`
}

// RedisConfigItem redis config item
type RedisConfigItem struct {
	Name     string `json:"name"`
	Default  bool   `json:"default"`
	Enable   bool   `json:"enable"`
	Host     string `json:"host"`
	PoolSize int    `json:"pool_size" mapstructure:"enable_auto_migrate"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

// GetConfig read config
func GetConfig() Config {
	path := "../conf/default.json"
	viper.SetConfigFile(path)
	// 读取配置文件并检查错误
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("配置文件读取错误: %s", err))
	}
	c := &Config{}
	if err := viper.Unmarshal(c); err != nil {
		panic(err)
	}
	return *c
}
