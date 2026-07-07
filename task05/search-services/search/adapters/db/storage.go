package db

import (
	"context"
	"log/slog"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"yadro.com/course/search/core"
)

type DB struct {
	log  *slog.Logger
	conn *sqlx.DB
}

type dbComics struct {
	ID    int            `db:"id"`
	URL   string         `db:"url"`
	Words pq.StringArray `db:"words"`
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

func (db *DB) GetAll(ctx context.Context) ([]core.Comics, error) {
	var comics []dbComics
	q := `SELECT id, url, words FROM comics`
	err := db.conn.SelectContext(ctx, &comics, q)
	if err != nil {
		db.log.ErrorContext(ctx, "failed to get all comics", "error", err)
		return nil, err
	}

	res := make([]core.Comics, len(comics))
	for i, c := range comics {
		res[i] = core.Comics{
			ID:    c.ID,
			URL:   c.URL,
			Words: []string(c.Words),
		}
	}

	return res, nil
}
