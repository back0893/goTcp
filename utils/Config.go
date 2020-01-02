package utils

import (
	"github.com/spf13/viper"
	"log"
)

var GlobalConfig *Config

type Config struct {
	*viper.Viper
	hook func(config *Config)
}

func NewConfig() *Config {
	return &Config{
		Viper: viper.New(),
	}
}
func init() {
	GlobalConfig = NewConfig()
}

/**
自定义的在配置加载时的动作
*/
func (config *Config) Hook(hook func(config *Config)) {
	config.hook = hook
}
func (config *Config) Load(configType string, configFile string) {
	config.SetConfigType(configType)
	config.SetConfigFile(configFile)
	if err := config.ReadInConfig(); err != nil {
		switch err.(type) {
		case viper.ConfigFileNotFoundError:
			log.Println("not found")
		default:
			panic(err)
		}
	}
	config.SetDefault("PacketReceiveChanLimit", 128)
	config.SetDefault("PacketSendChanLimit", 128)
	config.SetDefault("Ip", "0.0.0.0")
	config.SetDefault("Port", 8001)
	if config.hook != nil {
		config.hook(config)
	}
}
