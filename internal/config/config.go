package config

import (
	"go-test/pkg/logging"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	IsDebug *bool `yaml:"is_debug"`
	Listen  struct {
		Type   string `yaml:"type"`
		BindIP string `yaml:"bind_ip"`
		Port   string `yaml:"port"`
	} `yaml:"listen"`
	MongoDB struct {
		Host string `json:"host"`
		Port string `json:"port"`
		Database string `json:"database"`
		Auth_db string `json:"auth_db"`
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"mongodb"`
	Jwt struct {
		Secret_key string `json:"secret_key"`
	}
}

var instance *Config
var once sync.Once

func GetConfig() *Config {
	once.Do(func() {
		logger := logging.GetLogger()
		logger.Info("Read application configuration")
		instance = &Config{}
		
		if err := cleanenv.ReadConfig("../../config.yml", instance); err != nil {
			help, _ := cleanenv.GetDescription(instance, nil)
			logger.Info(help)
			logger.Fatal(err)
		}
	})
	return instance
}