package frame

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// LoadConfig read config
func LoadConfig(configPath ...string) *Config {
	var (
		ty   string
		path string
	)
	if len(configPath) > 0 && configPath[0] != "" {
		path = configPath[0]
		ty, path = getConfigFromPath(path)
	} else {
		ty, path = getConfigFromEnv()
	}
	viper.SetConfigFile(path)
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Printf("failed to read the configuration file, please check whether the %s file, err %v ", path, err)
		os.Exit(-1)
	}
	fmt.Printf("load configuration file path: %s\n", path)
	c := &Config{}
	if err := viper.Unmarshal(c); err != nil {
		fmt.Printf("load configuration file failed, err %v\n", err)
		os.Exit(-1)
	}
	if ers := c.validate(); ers != nil {
		fmt.Println("-----------------Load config failed------------------")
		for _, e := range ers {
			fmt.Println(e.Error())
		}
		fmt.Println("-------------Please fix the above errors-------------")
		os.Exit(-1)
	}
	// print loaded configuration content
	if c.PrintConf {
		fmt.Println("-------------Print loaded config content-------------")
		if ty == configTypeJSON {
			cbyts, _ := json.Marshal(c)
			fmt.Println(string(cbyts))
		} else {
			cbyts, _ := yaml.Marshal(c)
			fmt.Println(string(cbyts))
		}
		fmt.Println("----------------------Print end----------------------")
	}
	return c
}

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
	var ers []error
	if len(hs.Configs) <= 0 {
		ers = append(ers, errors.New("you enabled http server but didn’t declare the correct httpserver name/port, please set it again"))
	}
	for _, v := range hs.Configs {
		if err := v.Validate(); err != nil {
			ers = append(ers, err...)
		}
	}
	// port conflict
	for _, v := range hs.Configs {
		if isMetricPort(v.Port) && !(v.Name == defaultMetricName || v.Name == defaultMetricsName) {
			ers = append(ers, fmt.Errorf("%s http port can't be set to %s, %s is the default metric port, please reset %s http port", v.Name, defaultMetricPort2, defaultMetricPort, v.Name))
		}
	}
	if len(ers) <= 0 {
		return nil
	}
	return ers
}

// HTTPServerConfig http server config item
type HTTPServerConfig struct {
	Name string `json:"name"`
	Port string `json:"port"`
}

func (tsc HTTPServerConfig) isMetricConfig() bool {
	if tsc.Name == defaultMetricName || tsc.Name == defaultMetricsName {
		return true
	}
	return false
}

// Validate validate http server configs
func (tsc HTTPServerConfig) Validate() []error {
	var ers []error
	if tsc.Name == "" {
		ers = append(ers, errors.New("please fill in the correct http server name in the configuration file, name can't be empty, eg: server"))
	}
	if !strings.HasPrefix(tsc.Port, ":") {
		ers = append(ers, errors.New("please fill in the correct http server port in the configuration file, port can't be empty, eg :8080"))
	}
	port, _ := strconv.Atoi(strings.TrimPrefix(tsc.Port, ":"))
	if port <= 0 || port > 65535 {
		ers = append(ers, errors.New("please fill in the correct http server port in the configuration file, port range 1 ~ 65535, eg :8080"))
	}
	if len(ers) <= 0 {
		return nil
	}
	return ers
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
	var ers []error
	if len(mc.Configs) <= 0 {

		ers = append(ers, errors.New("you enabled mysql service but didn’t declare the correct mysql config, please set it again"))
	}
	for _, v := range mc.Configs {
		err := v.Validate()
		if err != nil {
			ers = append(ers, err...)
		}
	}
	if len(ers) <= 0 {
		return nil
	}
	return ers

}

func isMetricPort(port string) bool {
	return port == defaultMetricPort || port == defaultMetricPort2
}
func isMetricName(name string) bool {
	return name == defaultMetricName || name == defaultMetricsName
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
	var ers []error
	if mci.Name == "" {
		ers = append(ers, errors.New("please fill in the correct mysql name in the configuration file, name can't be empty, eg: demo"))
	}
	if mci.Host == "" {
		ers = append(ers, errors.New("please fill in the correct mysql host in the configuration file, host can't be empty, eg: 127.0.0.1:3306"))
	}
	if mci.Database == "" {
		ers = append(ers, errors.New("please fill in the correct mysql database in the configuration file, database can't be empty, eg: demo"))
	}
	if mci.User == "" {
		ers = append(ers, errors.New("please fill in the correct mysql user in the configuration file, user can't be empty, eg: demo"))
	}
	if mci.Password == "" {
		ers = append(ers, errors.New("please fill in the correct mysql password in the configuration file, password can't be empty, eg: demo"))
	}
	if len(ers) <= 0 {
		return nil
	}
	return ers
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
	var ers []error
	if len(rc.Configs) <= 0 {
		ers = append(ers, errors.New("you enabled redis service but didn’t declare the correct redis config, please set it again"))
	}
	for _, v := range rc.Configs {
		err := v.Validate()
		if err != nil {
			ers = append(ers, err...)
		}
	}
	return ers
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
	var ers []error
	if !rci.Enable {
		return nil
	}
	if rci.Name == "" {
		ers = append(ers, errors.New("please fill in the correct redis name in the configuration file, name can't be empty, eg: demo"))
	}
	if rci.Host == "" {
		ers = append(ers, errors.New("please fill in the correct redis host in the configuration file, host can't be empty, eg: 127.0.0.1:6379"))
	}
	if len(ers) <= 0 {
		return nil
	}
	return ers
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

func (hs HTTPServer) isMergeMetricBusPort() bool {
	if !hs.Enable {
		return false
	}
	msc := hs.getMetricServerConfig()
	bsc := hs.getBusServerConfig()
	var (
		mPort = ""
		bPort = ""
	)
	if msc != nil && bsc != nil {
		mPort = strings.TrimPrefix(msc.Port, ":")
		bPort = strings.TrimPrefix(bsc.Port, ":")
	}
	return mPort != "" && bPort == mPort
}

func (c *Config) isMergeMetricBusPort() bool {
	if !c.EnableMetric {
		return false
	}
	return c.HTTPServer.isMergeMetricBusPort()
}

func (c *Config) setDefaultValue() {
	if c.LogLevel == "" {
		c.LogLevel = logLevelInfo // default info
	}
	if c.LogMode == "" {
		c.LogMode = ModeJSON // default json
	}
	// metric server
	if c.HTTPServer.Enable && c.EnableMetric {
		msc := c.HTTPServer.getMetricServerConfig()
		if msc == nil {
			c.setDefaultMetricServerConfig()
		}
	}
	if c.HTTPServer.Enable {
		c.setDefaultBusServerConfig()
	}
	//  mysql default
	if c.Mysql.Enable {
		for i := range c.Mysql.Configs {
			if c.Mysql.Configs[i].SlowThresholdSec <= 0 {
				c.Mysql.Configs[i].SlowThresholdSec = 5
			}
			if len(c.Mysql.Configs[i].Database) > 0 && c.Mysql.Configs[i].Name == "" {
				c.Mysql.Configs[i].Name = c.Mysql.Configs[i].Database
			}
		}
	}

}

func (c *Config) setDefaultMetricServerConfig() {
	if len(c.HTTPServer.Configs) <= 0 {
		c.HTTPServer.Configs = []HTTPServerConfig{defaultMetricHTTPConfig}
		return
	}
	msc := c.HTTPServer.getMetricServerConfig()
	// only metric port
	if len(c.HTTPServer.Configs) == 1 && msc != nil {
		c.HTTPServer.Configs = append(c.HTTPServer.Configs, defaultMetricHTTPConfig)
	}

}

func (c *Config) setDefaultBusServerConfig() {
	if len(c.HTTPServer.Configs) <= 0 {
		c.HTTPServer.Configs = []HTTPServerConfig{defaultMetricHTTPConfig}
		return
	}
	c.HTTPServer.Configs = append(c.HTTPServer.Configs, defaultBusHTTPConfig)
}

func (c *Config) isEnableMySQLAutoMigrate(dbName string) bool {
	if len(c.Mysql.Configs) < 0 {
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
	var ers []error
	if len(c.Project) <= 0 {
		ers = append(ers, errors.New("please fill in the correct project name in the configuration file, it can't be empty, eg: demo"))
	}
	c.LogLevel = strings.ToLower(c.LogLevel)
	if _, ok := logm[c.LogLevel]; !ok {
		ers = append(ers, errors.New("please fill in the correct log_level in the configuration file, choose one of: trace/debug/info/warn/error/fatal/panic"))
	}

	c.LogMode = strings.ToLower(c.LogMode)
	if !(c.LogMode == ModeJSON || c.LogMode == ModelText) {
		ers = append(ers, errors.New("please fill in the correct log_mode in the configuration file, choose one of: json/text"))
	}
	if c.Env == "" {
		ers = append(ers, errors.New("please fill in the correct env  in the configuration file, it can't be empty, eg: dev"))
	}
	if err := c.HTTPServer.Validate(); err != nil {
		ers = append(ers, err...)
	}
	if err := c.Mysql.Validate(); err != nil {
		ers = append(ers, err...)
	}
	if err := c.Redis.Validate(); err != nil {
		ers = append(ers, err...)
	}
	if len(ers) <= 0 {
		return nil
	}
	return ers

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
