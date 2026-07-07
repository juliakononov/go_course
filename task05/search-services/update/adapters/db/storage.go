package db

import (
	"context"
	"log/slog"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"yadro.com/course/update/core"
)

type DB struct {
	log  *slog.Logger
	conn *sqlx.DB
}

type dbStats struct {
	WordsTotal    int `db:"words_total"`
	WordsUnique   int `db:"words_unique"`
	ComicsFetched int `db:"comics_fetched"`
}

func New(log *slog.Logger, address string) (*DB, error) {

	db, err := sqlx.Connect("pgx", address)
	if err != nil {
		log.Error("connection problem", "address", address, "error", err)
		return nil, err
	}

	return &DB{
		log:  log,
		conn: db,
	}, nil
}

func (db *DB) Add(ctx context.Context, comics core.Comics) error {
	q := `INSERT INTO comics (id, url, words) VALUES ($1, $2, $3)`
	_, err := db.conn.ExecContext(ctx, q, comics.ID, comics.URL, pq.Array(comics.Words))
	return err
}

func (db *DB) Stats(ctx context.Context) (core.DBStats, error) {
	var stats dbStats
	q := `
    WITH all_words AS (
        SELECT unnest(words) as word
        FROM comics
        WHERE words IS NOT NULL AND array_length(words, 1) > 0
    )
    SELECT 
        COALESCE((SELECT COUNT(*) FROM all_words), 0) as words_total,
        COALESCE((SELECT COUNT(DISTINCT word) FROM all_words), 0) as words_unique,
        COALESCE((SELECT COUNT(*) FROM comics), 0) as comics_fetched
	`
	err := db.conn.GetContext(ctx, &stats, q)
	if err != nil {
		return core.DBStats{}, err
	}
	return core.DBStats{
		WordsTotal:    stats.WordsTotal,
		WordsUnique:   stats.WordsUnique,
		ComicsFetched: stats.ComicsFetched,
	}, nil
}

func (db *DB) IDs(ctx context.Context) ([]int, error) {
	var ids []int
	err := db.conn.SelectContext(ctx, &ids, `SELECT id FROM comics`)
	if err != nil {
		return nil, err
	}
	return ids, nil
}

func (db *DB) Drop(ctx context.Context) error {
	_, err := db.conn.ExecContext(ctx, "DELETE FROM comics")
	return err
}

