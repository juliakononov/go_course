package main

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Port int `yaml:"port" env:"HELLO_PORT" env-default:"8080"`
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.Parse()

	var cfg Config

	if configPath != "" {
		//reads YAML file, then overrides with env
		if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
			slog.Error("helloserver: cannot read config", "error", err)
		}
	} else {
		//reads only env
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			slog.Error("helloserver: cannot read env", "error", err)
		}
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "pong")
	})

	mux.HandleFunc("GET /hello", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")

		if name == "" {
			http.Error(w, "empty name", http.StatusBadRequest)
			return
		}
		fmt.Fprintf(w, "Hello, %s!\n", name)
	})

	addr := fmt.Sprintf(":%d", cfg.Port)
	slog.Info("helloserver: starting hello server", "addr", addr)

	if err := http.ListenAndServe(addr, mux); !errors.Is(err, http.ErrServerClosed) {
		slog.Error("helloserver", "error", err)
	}
}
