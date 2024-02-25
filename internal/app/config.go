package app

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

type (
	Config struct {
		Server        ServerConfig        `yaml:"server"`
		Publisher     PublisherConfig     `yaml:"publisher"`
		MemoryStorage MemoryStorageConfig `yaml:"memory_storage"`
		API           APIConfig           `yaml:"api"`
	}

	PublisherConfig struct {
		Protocol string `yaml:"protocol"`
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	}

	ServerConfig struct {
		Debug *bool  `yaml:"debug"`
		Host  string `yaml:"host"`
		Port  string `yaml:"port"`
	}

	MemoryStorageConfig struct {
		Host      string `yaml:"host"`
		Port      string `yaml:"port"`
		KeyPrefix string `yaml:"key_prefix"`
	}

	APIConfig struct {
		Protocol string `yaml:"protocol"`
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
	}
)

func (a APIConfig) URL() string {
	return fmt.Sprintf("%s://%s:%s", a.Protocol, a.Host, a.Port)
}

func (m MemoryStorageConfig) URL() string {
	return fmt.Sprintf("%s:%s", m.Host, m.Port)
}

func (p PublisherConfig) URL() string {
	return fmt.Sprintf("%s://%s:%s@%s:%s",
		p.Protocol,
		p.Username,
		p.Password,
		p.Host,
		p.Port,
	)
}

func (s ServerConfig) URL() string {
	return fmt.Sprintf("%s:%s", s.Host, s.Port)
}

var instance *Config
var once sync.Once
var devConfigPath string = "../../config/dev_env.yaml"
var prodConfigPath string = "../../config/prod_env.yaml"

func GetConfig() *Config {
	once.Do(func() {
		instance = &Config{}
		configPath := prodConfigPath
		if _, ok := os.LookupEnv("DEVENV"); ok {
			configPath = devConfigPath
		}
		log.Println("read application configuration from", configPath)
		if err := cleanenv.ReadConfig(configPath, instance); err != nil {
			help, _ := cleanenv.GetDescription(instance, nil)
			log.Println(help)
			log.Fatal(err)
		}
	})
	return instance
}
