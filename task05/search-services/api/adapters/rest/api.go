package rest

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"yadro.com/course/api/core"
)

type PingResponse struct {
	Replies map[string]string `json:"replies"`
}

type WordsResponse struct {
	Words []string `json:"words"`
	Total int      `json:"total"`
}
type Comics struct {
	ID  int    `json:"id"`
	URL string `json:"url"`
}

type SearchResponse struct {
	Comics []Comics `json:"comics"`
	Total  int      `json:"total"`
}

type StatusResponse struct {
	Status string `json:"status"`
}

type StatsResponse struct {
	WordsTotal    int `json:"words_total"`
	WordsUnique   int `json:"words_unique"`
	ComicsFetched int `json:"comics_fetched"`
	ComicsTotal   int `json:"comics_total"`
}

const searchLimit = 10

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
			switch {
			case errors.Is(err, core.ErrBadArguments):
				http.Error(w, "phrase too large or invalid", http.StatusBadRequest)
			case errors.Is(err, core.ErrServiceUnavailable):
				http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			default:
				http.Error(w, "internal error", http.StatusInternalServerError)
			}
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

func NewUpdateHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := updater.Update(r.Context()); err != nil {
			if errors.Is(err, core.ErrAlreadyExists) {
				w.WriteHeader(http.StatusAccepted)
				return
			}
			log.Error("failed to update db", "error", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func NewUpdateStatsHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats, err := updater.Stats(r.Context())
		if err != nil {
			log.Error("failed to get stats", "error", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(StatsResponse{
			WordsTotal:    stats.WordsTotal,
			WordsUnique:   stats.WordsUnique,
			ComicsFetched: stats.ComicsFetched,
			ComicsTotal:   stats.ComicsTotal,
		}); err != nil {
			log.Error("failed to encode response", "error", err)
		}
	}
}

func NewUpdateStatusHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status, err := updater.Status(r.Context())
		if err != nil {
			log.Error("failed to get status", "error", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(StatusResponse{
			Status: string(status),
		}); err != nil {
			log.Error("failed to encode response", "error", err)
		}
	}
}

func NewDropHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := updater.Drop(r.Context()); err != nil {
			log.Error("failed to drop db", "error", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func NewSearchHandler(log *slog.Logger, searcher core.Searcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		phrase := r.URL.Query().Get("phrase")
		if phrase == "" {
			http.Error(w, "empty phrase", http.StatusBadRequest)
			return
		}

		strLimit := r.URL.Query().Get("limit")
		if strLimit == "" {
			strLimit = strconv.Itoa(searchLimit)
		}

		limit, err := strconv.Atoi(strLimit)
		if err != nil {
			http.Error(w, "limit must be int", http.StatusBadRequest)
			return
		}

		if limit < 1 {
			http.Error(w, "limit must be greater than 0", http.StatusBadRequest)
			return
		}

		reply, err := searcher.Search(r.Context(), phrase, limit)
		if err != nil {
			log.Error("search failed", "phrase_len", len(phrase), "limit", limit, "error", err)
			switch {
			case errors.Is(err, core.ErrBadArguments):
				http.Error(w, "phrase too large or invalid", http.StatusBadRequest)
			case errors.Is(err, core.ErrServiceUnavailable):
				http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			default:
				http.Error(w, "internal error", http.StatusInternalServerError)
			}
			return
		}

		comics := make([]Comics, 0, len(reply))
		for _, c := range reply {
			comics = append(comics, Comics{ID: int(c.ID), URL: c.URL})
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(SearchResponse{
			Comics: comics,
			Total:  len(comics),
		}); err != nil {
			log.Error("failed to encode response", "error", err)
		}
	}
}
