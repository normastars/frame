package frame

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Config project config
type Config struct {
	Project      string      `json:"project"`
	LogLevel     string      `json:"log_level" yaml:"log_level"  mapstructure:"log_level"`
	LogMode      string      `json:"log_mode" yaml:"log_mode"  mapstructure:"log_mode"`
	EnableMetric bool        `json:"enable_metric" yaml:"enable_metric" mapstructure:"enable_metric"`
	Env          string      `json:"env"`
	HTTPServer   HTTPServer  `json:"http_server" yaml:"http_server" mapstructure:"http_server"`
	Mysql        MySQLConfig `json:"mysql"`
	Redis        RedisConfig `json:"redis"`
}

// HTTPServer http config
type HTTPServer struct {
	Enable        bool               `json:"enable"`
	DisableReqLog bool               `json:"disable_req_log" yaml:"disable_req_log" mapstructure:"disable_req_log"` // default enable
	Configs       []HTTPServerConfig `json:"configs"`
}

// HTTPServerConfig http server config item
type HTTPServerConfig struct {
	Name string `json:"name"`
	Port string `json:"port"`
}

// MySQLConfig mysql config
type MySQLConfig struct {
	Enable        bool              `json:"enable"`
	DisableReqLog bool              `json:"disable_req_log" yaml:"disable_req_log" mapstructure:"disable_req_log"` // default enable
	Configs       []MySQLConfigItem `json:"configs"`
}

// MySQLConfigItem mysql config item
type MySQLConfigItem struct {
	Name              string `json:"name"`
	Enable            bool   `json:"enable"`
	EnableAutoMigrate bool   `json:"enable_auto_migrate" yaml:"enable_auto_migrate" mapstructure:"enable_auto_migrate"` // default disable
	Host              string `json:"host"`
	Database          string `json:"database"`
	User              string `json:"user"`
	Password          string `json:"password"`
	SlowThresholdSec  int    `json:"slow_threshold_sec" yaml:"slow_threshold_sec" mapstructure:"slow_threshold_sec"`
}

// RedisConfig redis config
type RedisConfig struct {
	Enable        bool              `json:"enable"`
	DisableReqLog bool              `json:"disable_req_log" yaml:"disable_req_log" mapstructure:"disable_req_log"` // default enable
	Configs       []RedisConfigItem `json:"configs"`
}

// RedisConfigItem redis config item
type RedisConfigItem struct {
	Name     string `json:"name"`
	Enable   bool   `json:"enable"`
	Host     string `json:"host"`
	PoolSize int    `json:"pool_size" mapstructure:"enable_auto_migrate"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

// LoadConfig read config
func LoadConfig() *Config {
	path := "./conf/default.yaml"
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
	jb, _ := json.Marshal(c)

	fmt.Println(string(jb))
	fmt.Println("=====")
	yb, _ := yaml.Marshal(c)
	fmt.Println(string(yb))
	return c
}
