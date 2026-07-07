package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	LogLevel     string `yaml:"log_level" env:"LOG_LEVEL" env-default:"DEBUG"`
	Address      string `yaml:"search_address" env:"SEARCH_ADDRESS" env-default:"localhost:83"`
	DBAddress    string `yaml:"db_address" env:"DB_ADDRESS" env-default:"localhost:82"`
	WordsAddress string `yaml:"words_address" env:"WORDS_ADDRESS" env-default:"localhost:81"`
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
