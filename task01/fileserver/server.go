package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Port    int    `yaml:"port" env:"FILESERVER_PORT" env-default:"8080"`
	Storage string `yaml:"storage" env:"FILESERVER_STORAGE" env-default:"/tmp/storage"`
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.Parse()

	var cfg Config

	if configPath != "" {
		if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
			slog.Error("fileserver: cannot read config", "error", err)
		}
	} else {
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			slog.Error("fileserver: cannot read env", "error", err)
		}
	}

	if err := os.MkdirAll(cfg.Storage, 0755); err != nil {
		slog.Error("fileserver: cannot create storage dir", "error", err)
	}

	root, err := os.OpenRoot(cfg.Storage)
	if err != nil {
		slog.Error("fileserver: cannot open storage root", "error", err)
		return
	}
	defer func() {
		if err := root.Close(); err != nil {
			slog.Error("fileserver: failed to close root", "error", err)
		}
	}()

	mux := http.NewServeMux()

	mux.HandleFunc("POST /files", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			http.Error(w, "cannot parse form", http.StatusBadRequest)
			return
		}

		formFile, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "cannot get file", http.StatusBadRequest)
			return
		}
		defer func() {
			if err := formFile.Close(); err != nil {
				slog.Error("fileserver: failed to close form file", "error", err)
			}
		}()

		if _, err := root.Stat(header.Filename); err == nil {
			http.Error(w, "file already exists", http.StatusConflict)
			return
		}

		dest, err := root.Create(header.Filename)
		if err != nil {
			http.Error(w, "cannot create file", http.StatusInternalServerError)
			return
		}
		defer func() {
			if err := dest.Close(); err != nil {
				slog.Error("fileserver: failed to close dest file", "error", err)
			}
		}()

		if _, err := io.Copy(dest, formFile); err != nil {
			http.Error(w, "cannot write file", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		fmt.Fprintln(w, header.Filename)
	})

	mux.HandleFunc("GET /files", func(w http.ResponseWriter, r *http.Request) {
		files, err := os.ReadDir(cfg.Storage)
		if err != nil {
			http.Error(w, "cannot read storage", http.StatusInternalServerError)
			return
		}

		names := make([]string, 0, len(files))
		for _, e := range files {
			if !e.IsDir() {
				names = append(names, e.Name())
			}
		}

		sort.Strings(names)

		fmt.Fprint(w, strings.Join(names, "\n"))
		if len(names) > 0 {
			fmt.Fprintln(w)
		}
	})

	mux.HandleFunc("GET /files/{filename}", func(w http.ResponseWriter, r *http.Request) {
		filename := r.PathValue("filename")

		if _, err := root.Stat(filename); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			http.Error(w, "cannot stat file", http.StatusInternalServerError)
			return
		}

		http.ServeFile(w, r, filepath.Join(cfg.Storage, filename))
	})

	mux.HandleFunc("PUT /files/{filename}", func(w http.ResponseWriter, r *http.Request) {
		filename := r.PathValue("filename")

		if _, err := root.Stat(filename); errors.Is(err, os.ErrNotExist) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		if err := r.ParseMultipartForm(1 << 20); err != nil {
			http.Error(w, "cannot parse form", http.StatusBadRequest)
			return
		}

		formFile, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "cannot get file", http.StatusBadRequest)
			return
		}
		defer func() {
			if err := formFile.Close(); err != nil {
				slog.Error("fileserver: failed to close form file", "error", err)
			}
		}()

		dest, err := root.Create(filename)
		if err != nil {
			http.Error(w, "cannot create file", http.StatusInternalServerError)
			return
		}
		defer func() {
			if err := dest.Close(); err != nil {
				slog.Error("fileserver: failed to close dest file", "error", err)
			}
		}()

		if _, err := io.Copy(dest, formFile); err != nil {
			http.Error(w, "cannot write file", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("DELETE /files/{filename}", func(w http.ResponseWriter, r *http.Request) {
		filename := r.PathValue("filename")

		if err := root.Remove(filename); err != nil && !errors.Is(err, os.ErrNotExist) {
			slog.Error("fileserver: cannot remove file", "filename", filename, "error", err)
		}

		w.WriteHeader(http.StatusOK)
	})

	addr := fmt.Sprintf(":%d", cfg.Port)
	slog.Info("fileserver: starting fileserver", "addr", addr, "storage", cfg.Storage)

	if err := http.ListenAndServe(addr, mux); !errors.Is(err, http.ErrServerClosed) {
		slog.Error("fileserver:", "error", err)
	}
}
