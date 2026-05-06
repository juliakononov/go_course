package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	"yadro.com/course/api/adapters/rest"
	"yadro.com/course/api/adapters/words"
	"yadro.com/course/api/config"
	"yadro.com/course/api/core"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "server configuration file")
	flag.Parse()

	cfg := config.MustLoad(configPath)

	log := mustMakeLogger(cfg.LogLevel)

	log.Info("starting server", "addr", cfg.Server.Address)
	log.Debug("debug messages are enabled")

	wordsClient, err := words.NewClient(cfg.WordsAddress, log)
	if err != nil {
		log.Error("cannot init words adapter", "error", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.Handle("GET /api/words", rest.NewWordsHandler(log, wordsClient))
	mux.Handle("GET /ping", rest.NewPingHandler(log, map[string]core.Pinger{"words": wordsClient}))

	server := http.Server{
		Addr:        cfg.Server.Address,
		ReadTimeout: cfg.Server.Timeout,
		Handler:     mux,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		<-ctx.Done()
		log.Debug("shutting down server")
		if err := server.Shutdown(context.Background()); err != nil {
			log.Error("erroneous shutdown", "error", err)
		}
	}()

	if err := server.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			log.Error("server closed unexpectedly", "error", err)
			return
		}
	}
}

func mustMakeLogger(logLevel string) *slog.Logger {
	var level slog.Level
	if err := level.UnmarshalText([]byte(logLevel)); err != nil {
		slog.Error("unknown log level", "level", logLevel, "error", err)
		os.Exit(1)
	}
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})
	return slog.New(handler)
}
