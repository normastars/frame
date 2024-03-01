package frame

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config project config
type Config struct {
	Project      string       `json:"project"`
	LogLevel     string       `json:"log_level" yaml:"log_level"  mapstructure:"log_level"`
	LogMode      string       `json:"log_mode" yaml:"log_mode"  mapstructure:"log_mode"`
	PrintConf    bool         `json:"print_conf" yaml:"print_conf"  mapstructure:"print_conf"`
	EnableMetric bool         `json:"enable_metric" yaml:"enable_metric" mapstructure:"enable_metric"`
	Env          string       `json:"env"`
	HTTPServer   HTTPServer   `json:"http_server" yaml:"http_server" mapstructure:"http_server"`
	HTTPClient   DoHTTPClient `json:"http_client" yaml:"http_client" mapstructure:"http_client"`
	Mysql        MySQLConfig  `json:"mysql"`
	Redis        RedisConfig  `json:"redis"`
}

// DoHTTPClient http client config
type DoHTTPClient struct {
	DisableReqLog bool `json:"disable_req_log" yaml:"disable_req_log" mapstructure:"disable_req_log"` // default enable
	EnableMetric  bool `json:"enable_metric" yaml:"enable_metric" mapstructure:"enable_metric"`
}

// HTTPServer http config
type HTTPServer struct {
	Enable        bool               `json:"enable"`
	EnableCors    bool               `json:"enable_cors" yaml:"enable_cors" mapstructure:"enable_cors"`
	DisableReqLog bool               `json:"disable_req_log" yaml:"disable_req_log" mapstructure:"disable_req_log"` // default enable
	Configs       []HTTPServerConfig `json:"configs"`
}

// Validate check http server
func (hs HTTPServer) Validate() []error {
	if !hs.Enable {
		return nil
	}
	var errs []error
	if len(hs.Configs) <= 0 {
		errs = append(errs, errors.New("you enabled http server but didn’t declare the correct httpserver name/port, please set it again"))
	}
	for _, v := range hs.Configs {
		if err := v.Validate(); err != nil {
			errs = append(errs, err...)
		}
	}
	// port conflict
	for _, v := range hs.Configs {
		if isMetricPort(v.Port) && !(v.Name == defaultMetricName || v.Name == defaultMetricsName) {
			errs = append(errs, fmt.Errorf("%s http port can't be set to %s, %s is the default metric port, please reset %s http port", v.Name, defaultMetricPort2, defaultMetricPort, v.Name))
		}
	}
	if len(errs) <= 0 {
		return nil
	}
	return errs
}

// HTTPServerConfig http server config item
type HTTPServerConfig struct {
	Name string `json:"name"`
	Port string `json:"port"`
}

// Validate validate http server configs
func (tsc HTTPServerConfig) Validate() []error {
	var errs []error
	if tsc.Name == "" {
		errs = append(errs, errors.New("please fill in the correct http server name in the configuration file, name can't be empty, eg: server"))
	}
	if !strings.HasPrefix(tsc.Port, ":") {
		errs = append(errs, errors.New("please fill in the correct http server port in the configuration file, port can't be empty, eg :8080"))
	}
	port, _ := strconv.Atoi(strings.TrimPrefix(tsc.Port, ":"))
	if port <= 0 || port > 65535 {
		errs = append(errs, errors.New("please fill in the correct http server port in the configuration file, port range 1 ~ 65535, eg :8080"))
	}
	if len(errs) <= 0 {
		return nil
	}
	return errs
}

// MySQLConfig mysql config
type MySQLConfig struct {
	Enable        bool              `json:"enable"`
	DisableReqLog bool              `json:"disable_req_log" yaml:"disable_req_log" mapstructure:"disable_req_log"` // default enable
	Configs       []MySQLConfigItem `json:"configs"`
}

// Validate validate mysql config
func (mc MySQLConfig) Validate() []error {
	if !mc.Enable {
		return nil
	}
	var errs []error
	if len(mc.Configs) <= 0 {

		errs = append(errs, errors.New("you enabled mysql service but didn’t declare the correct mysql config, please set it again"))
	}
	for _, v := range mc.Configs {
		err := v.Validate()
		if err != nil {
			errs = append(errs, err...)
		}
	}
	if len(errs) <= 0 {
		return nil
	}
	return errs

}

func isMetricPort(port string) bool {
	return port == defaultMetricPort || port == defaultMetricPort2
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

// Validate mysql config item validate
func (mci MySQLConfigItem) Validate() []error {
	if !mci.Enable {
		return nil
	}
	var errs []error
	if mci.Name == "" {
		errs = append(errs, errors.New("please fill in the correct mysql name in the configuration file, name can't be empty, eg: demo"))
	}
	if mci.Host == "" {
		errs = append(errs, errors.New("please fill in the correct mysql host in the configuration file, host can't be empty, eg: 127.0.0.1:3306"))
	}
	if mci.Database == "" {
		errs = append(errs, errors.New("please fill in the correct mysql database in the configuration file, database can't be empty, eg: demo"))
	}
	if mci.User == "" {
		errs = append(errs, errors.New("please fill in the correct mysql user in the configuration file, user can't be empty, eg: demo"))
	}
	if mci.Password == "" {
		errs = append(errs, errors.New("please fill in the correct mysql password in the configuration file, password can't be empty, eg: demo"))
	}
	if len(errs) <= 0 {
		return nil
	}
	return errs
}

// RedisConfig redis config
type RedisConfig struct {
	Enable        bool              `json:"enable"`
	DisableReqLog bool              `json:"disable_req_log" yaml:"disable_req_log" mapstructure:"disable_req_log"` // default enable
	Configs       []RedisConfigItem `json:"configs"`
}

// Validate redis validate
func (rc RedisConfig) Validate() []error {
	if !rc.Enable {
		return nil
	}
	var errs []error
	if len(rc.Configs) <= 0 {
		errs = append(errs, errors.New("you enabled redis service but didn’t declare the correct redis config, please set it again"))
	}
	for _, v := range rc.Configs {
		err := v.Validate()
		if err != nil {
			errs = append(errs, err...)
		}
	}
	return errs
}

// RedisConfigItem redis config item
type RedisConfigItem struct {
	Name     string `json:"name"`
	Enable   bool   `json:"enable"`
	Host     string `json:"host"`
	PoolSize int    `json:"pool_size" mapstructure:"pool_size"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

// Validate redis config item validate
func (rci RedisConfigItem) Validate() []error {
	var errs []error
	if !rci.Enable {
		return nil
	}
	if rci.Name == "" {
		errs = append(errs, errors.New("please fill in the correct redis name in the configuration file, name can't be empty, eg: demo"))
	}
	if rci.Host == "" {
		errs = append(errs, errors.New("please fill in the correct redis host in the configuration file, host can't be empty, eg: 127.0.0.1:6379"))
	}
	if len(errs) <= 0 {
		return nil
	}
	return errs
}

// GetMetricServerConfig return http metric server
func (hs HTTPServer) getMetricServerConfig() *HTTPServerConfig {
	for _, v := range hs.Configs {
		if v.Name == defaultMetricName || v.Name == defaultMetricsName {
			return &v
		}
	}
	return nil
}

func (hs HTTPServer) getBusServerConfig() *HTTPServerConfig {
	for _, v := range hs.Configs {
		if !(v.Name == defaultMetricName || v.Name == defaultMetricsName) {
			return &v
		}
	}
	return nil
}

func (c *Config) isEnableMySQLAutoMigrate(dbName string) bool {
	if len(c.Mysql.Configs) <= 0 {
		return false
	}
	if !c.Mysql.Enable {
		return false
	}
	for _, v := range c.Mysql.Configs {
		if v.Name == dbName && v.EnableAutoMigrate {
			return true
		}
	}
	return false
}

// Validate validate config
func (c *Config) validate() []error {
	var errs []error
	if len(c.Project) <= 0 {
		errs = append(errs, errors.New("please fill in the correct project name in the configuration file, it can't be empty, eg: demo"))
	}
	c.LogLevel = strings.ToLower(c.LogLevel)
	if _, ok := logm[c.LogLevel]; !ok {
		errs = append(errs, errors.New("please fill in the correct log_level in the configuration file, choose one of: trace/debug/info/warn/error/fatal/panic"))
	}

	c.LogMode = strings.ToLower(c.LogMode)
	if !(c.LogMode == ModeJSON || c.LogMode == ModelText) {
		errs = append(errs, errors.New("please fill in the correct log_mode in the configuration file, choose one of: json/text"))
	}
	if c.Env == "" {
		errs = append(errs, errors.New("please fill in the correct env  in the configuration file, it can't be empty, eg: dev"))
	}
	if err := c.HTTPServer.Validate(); err != nil {
		errs = append(errs, err...)
	}
	if err := c.Mysql.Validate(); err != nil {
		errs = append(errs, err...)
	}
	if err := c.Redis.Validate(); err != nil {
		errs = append(errs, err...)
	}
	if len(errs) <= 0 {
		return nil
	}
	return errs

}

func (c *Config) getMetricPort() string {
	return c.HTTPServer.getMetricServerConfig().Port
}

func (c *Config) getServerPort() string {
	return c.HTTPServer.getBusServerConfig().Port
}

func getConfigFromEnv() (t, path string) {
	path = os.Getenv(configPath)
	if len(path) <= 0 {
		path = configDefaultPath
	}
	if strings.HasSuffix(path, configTypeYal) || strings.HasSuffix(path, configTypeYaml) {
		t = configTypeYaml
	} else {
		t = configTypeJSON
	}
	return
}
func getConfigFromPath(pa string) (t string, path string) {
	path = pa
	if strings.HasSuffix(path, configTypeYal) || strings.HasSuffix(path, configTypeYaml) {
		t = configTypeYaml
	} else {
		t = configTypeJSON
	}
	return t, path
}
