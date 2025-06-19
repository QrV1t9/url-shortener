package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go-urlshortner/internal/storage"
)

type Storage struct {
	db *pgx.Conn
}

func New(url string) (*Storage, error) {
	const op = "storage.postgres.new"

	db, err := pgx.Connect(context.Background(), url)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	stmt := `CREATE TABLE IF NOT EXISTS url(
	    id SERIAL PRIMARY KEY,
	    alias TEXT NOT NULL UNIQUE,
	    url TEXT NOT NULL);
	CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);
`

	_, err = db.Exec(context.Background(), stmt)

	if err != nil {
		db.Close(context.Background())
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Close(ctx context.Context) {
	s.db.Close(ctx)
}

func (s *Storage) SaveURL(urlToSave string, alias string) error {
	const op = "storage.postgres.SaveURL"

	_, err := s.db.Exec(context.Background(), "INSERT INTO url(url, alias) VALUES($1,$2)", urlToSave, alias)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return fmt.Errorf("%s: %w", op, storage.ErrUrlExists)
			}
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Storage) GetURL(alias string) (string, error) {
	const op = "storage.postgres.SaveURL"

	var url string
	err := s.db.QueryRow(context.Background(), "SELECT url FROM url WHERE alias = $1", alias).Scan(&url)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", fmt.Errorf("%s: %w", op, storage.ErrUrlNotFound)
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return url, nil
}

func (s *Storage) DeleteURL(alias string) error {
	const op = "storage.postgres.DeleteURL"

	cmdTag, err := s.db.Exec(context.Background(), "DELETE FROM url WHERE alias = $1", alias)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("%s: %w", op, storage.ErrUrlNotFound)
	}

	return nil
}