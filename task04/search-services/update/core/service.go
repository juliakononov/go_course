package core

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
)

type Service struct {
	log         *slog.Logger
	db          DB
	xkcd        XKCD
	words       Words
	concurrency int
	isRunning   atomic.Bool
}

const maxPhraseLen = 4096

func NewService(
	log *slog.Logger, db DB, xkcd XKCD, words Words, concurrency int,
) (*Service, error) {
	if concurrency < 1 {
		return nil, fmt.Errorf("wrong concurrency specified: %d", concurrency)
	}
	return &Service{
		log:         log,
		db:          db,
		xkcd:        xkcd,
		words:       words,
		concurrency: concurrency,
	}, nil
}

func (s *Service) Update(ctx context.Context) (err error) {
	if !s.isRunning.CompareAndSwap(false, true) {
		return ErrAlreadyExists
	}
	defer s.isRunning.Store(false)

	s.log.Info("update started")

	missing, err := s.missingIDs(ctx)
	if err != nil {
		return err
	}
	s.log.Info("comics to fetch", "count", len(missing))

	s.fetchAll(ctx, missing)
	s.log.Info("update finished", "fetched", len(missing))
	return nil
}

func (s *Service) missingIDs(ctx context.Context) ([]int, error) {
	lastID, err := s.xkcd.LastID(ctx)
	if err != nil {
		return nil, err
	}

	existing, err := s.db.IDs(ctx)
	if err != nil {
		return nil, err
	}

	existingSet := make(map[int]struct{}, len(existing))
	for _, id := range existing {
		existingSet[id] = struct{}{}
	}

	var missing []int
	for id := 1; id <= lastID; id++ {
		if id == 404 {
			continue
		}

		if _, ok := existingSet[id]; !ok {
			missing = append(missing, id)
		}
	}

	return missing, nil
}

func (s *Service) fetchAll(ctx context.Context, missing []int) {
	var wg sync.WaitGroup
	tasks := make(chan int)

	for i := 0; i < s.concurrency; i++ {
		wg.Go(func() {
			for id := range tasks {
				if err := s.addComics(ctx, id); err != nil {
					s.log.Error("failed to add comics", "id", id, "error", err)
				}
			}
		})
	}

	for _, id := range missing {
		tasks <- id
	}
	close(tasks)

	wg.Wait()
}

func (s *Service) addComics(ctx context.Context, id int) error {
	comics, err := s.xkcd.Get(ctx, id)
	if err != nil {
		return err
	}

	phrase := fmt.Sprint(comics.Title, " ", comics.Description)
	if len(phrase) > maxPhraseLen {
		phrase = phrase[:maxPhraseLen]
	}

	words, err := s.words.Norm(ctx, phrase)
	if err != nil {
		return err
	}

	err = s.db.Add(ctx,
		Comics{
			ID:    comics.ID,
			URL:   comics.URL,
			Words: words,
		})
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Stats(ctx context.Context) (ServiceStats, error) {
	dbStats, err := s.db.Stats(ctx)
	if err != nil {
		return ServiceStats{}, err

	}

	lastID, err := s.xkcd.LastID(ctx)
	if err != nil {
		return ServiceStats{}, err
	}

	return ServiceStats{
		DBStats:     dbStats,
		ComicsTotal: lastID - 1,
	}, nil

}

func (s *Service) Status(ctx context.Context) ServiceStatus {
	if s.isRunning.Load() {
		return StatusRunning
	}
	return StatusIdle
}

func (s *Service) Drop(ctx context.Context) error {
	if !s.isRunning.CompareAndSwap(false, true) {
		return ErrAlreadyExists
	}
	defer s.isRunning.Store(false)

	s.log.Info("dropping database")
	return s.db.Drop(ctx)
}
