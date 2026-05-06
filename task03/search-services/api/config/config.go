package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Server struct {
		Address string        `yaml:"address" env:"HTTP_SERVER_ADDRESS" env-default:"localhost:8080"`
		Timeout time.Duration `yaml:"timeout" env:"HTTP_SERVER_TIMEOUT" env-default:"5s"`
	} `yaml:"http_server"`
	WordsAddress string `yaml:"words_address" env:"WORDS_ADDRESS" env-default:"localhost:8081"`
	LogLevel     string `yaml:"log_level" env:"LOG_LEVEL" env-default:"INFO"`
}

func MustLoad(configPath string) Config {
	var cfg Config

	if configPath != "" {
		if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
			log.Fatalf("cannot read config %q: %s", configPath, err)
		}
	} else {
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			log.Fatalf("cannot read env: %s", err)
		}
	}
	return cfg
}
