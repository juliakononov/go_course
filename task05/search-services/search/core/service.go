package core

import (
	"cmp"
	"context"
	"log/slog"
	"slices"
)

type Service struct {
	log   *slog.Logger
	db    DB
	words Words
}

type scoredComics struct {
	comics Comics
	score  int
}

func NewService(log *slog.Logger, db DB, words Words) (*Service, error) {
	return &Service{
		log:   log,
		db:    db,
		words: words,
	}, nil
}

func (s *Service) Search(ctx context.Context, phrase string, limit int) ([]ComicsInfo, error) {
	if limit < 1 {
		return nil, ErrBadArguments
	}

	queryWords, err := s.words.Norm(ctx, phrase)
	if err != nil {
		s.log.Error("failed to normalize phrase", "error", err)
		return nil, err
	}

	if len(queryWords) == 0 {
		s.log.Info("empty phrase after normalization", "phrase", phrase)
		return []ComicsInfo{}, nil
	}

	allComics, err := s.db.GetAll(ctx)
	if err != nil {
		s.log.Error("failed to load comics", "error", err)
		return nil, err
	}

	querySet := make(map[string]struct{}, len(queryWords))
	for _, w := range queryWords {
		querySet[w] = struct{}{}
	}

	var scored []scoredComics
	for _, c := range allComics {
		score := 0
		for _, cw := range c.Words {
			if _, ok := querySet[cw]; ok {
				score++
			}
		}
		if score > 0 {
			scored = append(scored, scoredComics{comics: c, score: score})
		}
	}

	slices.SortFunc(scored, func(a, b scoredComics) int {
		return cmp.Compare(b.score, a.score)
	})

	if len(scored) > limit {
		scored = scored[:limit]
	}

	result := make([]ComicsInfo, 0, len(scored))
	for _, sc := range scored {
		result = append(result, ComicsInfo{ID: sc.comics.ID, URL: sc.comics.URL})
	}

	s.log.Info("search finished", "query_words", len(queryWords), "found", len(result))
	return result, nil
}
