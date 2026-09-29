package main

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct{ DB *pgxpool.Pool }

type Wallpaper struct {
	ID          string     `json:"id"`
	FileName    string     `json:"fileName"`
	URL         string     `json:"url"`
	OriginalURL string     `json:"originalUrl"`
	IsMobile    int16      `json:"isMobile"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   *time.Time `json:"updatedAt"`
}

func (s *Store) ExistingLinks(ctx context.Context, links []string) (map[string]bool, error) {
	known := map[string]bool{}
	if len(links) == 0 {
		return known, nil
	}
	rows, err := s.DB.Query(ctx, `SELECT original_url FROM wallpapers WHERE original_url = ANY($1)`, links)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var link string
		if err := rows.Scan(&link); err != nil {
			return nil, err
		}
		known[link] = true
	}
	return known, rows.Err()
}

func (s *Store) All(ctx context.Context) ([]Wallpaper, error) {
	rows, err := s.DB.Query(ctx, `SELECT id::text, file_name, url, original_url, is_mobile, created_at, updated_at FROM wallpapers ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []Wallpaper{}
	for rows.Next() {
		var w Wallpaper
		if err := rows.Scan(&w.ID, &w.FileName, &w.URL, &w.OriginalURL, &w.IsMobile, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, w)
	}
	return result, rows.Err()
}

func (s *Store) Insert(ctx context.Context, w Wallpaper) error {
	_, err := s.DB.Exec(ctx, `INSERT INTO wallpapers (file_name, url, original_url, is_mobile, created_at) VALUES ($1,$2,$3,$4,$5) ON CONFLICT DO NOTHING`, w.FileName, w.URL, w.OriginalURL, w.IsMobile, w.CreatedAt)
	return err
}
