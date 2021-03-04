package config

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/fwhezfwhez/errorx"
	"github.com/spf13/viper"
	"os"
	"sync"
)

var config *viper.Viper
var m sync.Mutex

// Init 初始化配置
func init() {
	var env string
	if env = os.Getenv("ENV"); env == "" {
		env = "config"
	}
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigName(env)
	v.AddConfigPath("../config/")
	v.AddConfigPath("config/")
	ReadConfig(v)
	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("Config file changed:", e.Name)
	})

	config = v
}

// GetConfig 获取配置
func GetConfig() *viper.Viper {
	return config
}

func ReadConfig(v *viper.Viper) error {
	m.Lock()
	defer m.Unlock()
	err := v.ReadInConfig()
	if err != nil {
		return errorx.NewFromString("Error on parsing config file!")
	}
	return nil
}
