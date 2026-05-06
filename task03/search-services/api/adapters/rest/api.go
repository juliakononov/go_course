package rest

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"yadro.com/course/api/core"
)

type PingResponse struct {
	Replies map[string]string `json:"replies"`
}

type WordsResponse struct {
	Words []string `json:"words"`
	Total int      `json:"total"`
}

func NewPingHandler(log *slog.Logger, pingers map[string]core.Pinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		replies := make(map[string]string)
		for n, p := range pingers {
			if err := p.Ping(r.Context()); err != nil {
				log.Error("service unavailable", "service", n, "error", err)
				replies[n] = "unavailable"
			} else {
				replies[n] = "ok"
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(PingResponse{Replies: replies}); err != nil {
			log.Error("failed to encode response", "error", err)
		}
	}
}

func NewWordsHandler(log *slog.Logger, norm core.Normalizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		phrase := r.URL.Query().Get("phrase")

		if phrase == "" {
			http.Error(w, "empty phrase", http.StatusBadRequest)
			return
		}

		words, err := norm.Norm(r.Context(), phrase)
		if err != nil {
			log.Error("failed to normalize", "phrase_len", len(phrase), "error", err)
			if errors.Is(err, core.ErrBadArguments) {
				http.Error(w, "phrase too large", http.StatusBadRequest)
				return
			}
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		if words == nil {
			words = []string{}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(WordsResponse{
			Words: words,
			Total: len(words),
		}); err != nil {
			log.Error("failed to encode response", "error", err)
		}
	}
}
