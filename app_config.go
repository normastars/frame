package frame

import (
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// ConfigManager
type ConfigManager struct {
	*viper.Viper
}

func setConfigFilePath(path string) {
	fileSyncOnce.Do(func() {
		configFilePath = path
	})
}

// NewConfigManager new config manager
func NewConfigManager(configPath string) (*ConfigManager, error) {
	v := viper.New()
	v.SetConfigFile(configPath)
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %s", err)
	}
	return &ConfigManager{Viper: v}, nil
}

// ReadConfigObject read config
// add `mapstructure:"key"` tag
func (cm *ConfigManager) ReadConfigObject(obj interface{}) error {
	if !(cm.Viper != nil) {
		return fmt.Errorf("config manager is nil")
	}
	if err := cm.Unmarshal(obj); err != nil {
		return fmt.Errorf("failed to unmarshal config: %s", err)
	}
	return nil
}

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
	setConfigFilePath(path)

	cm, err := NewConfigManager(path)
	if err != nil {
		logrus.Fatalf("failed to read the configuration file, please check whether the %s file, err %v ", path, err)
	}
	logrus.Infof("load configuration file path: %s\n", path)
	c := &Config{}
	if err := cm.ReadConfigObject(c); err != nil {
		logrus.Fatalf("load configuration file failed, err %v\n", err)
	}
	if errs := c.validate(); errs != nil {
		logrus.Infoln("loading config")
		for _, e := range errs {
			logrus.Errorln(e)
		}
		logrus.Fatalln("please fix the above errors")
	}
	// print loaded configuration content
	if c.PrintConf {
		var cbyts []byte
		if ty == configTypeJSON {
			cbyts, _ = json.Marshal(c)
		} else {
			cbyts, _ = yaml.Marshal(c)
		}
		logrus.Infoln("loading config content: ", string(cbyts))
	}
	return c
}

// ReadAppConfigManager
func ReadAppConfigManager() (*ConfigManager, error) {
	return NewConfigManager(configFilePath)
}
