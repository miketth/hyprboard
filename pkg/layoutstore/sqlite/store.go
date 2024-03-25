package sqlite

import (
	"codeberg.org/miketth/hyprboard/pkg/hyprboard"
	"context"
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
)

type LayoutStore struct {
	db      *sql.DB
	querier *Queries
}

func NewLayoutStore(filename string) (*LayoutStore, error) {
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if err := createSchema(db); err != nil {
		return nil, fmt.Errorf("create schema: %w", err)
	}

	querier := New(db)

	return &LayoutStore{
		db:      db,
		querier: querier,
	}, nil
}

func (s *LayoutStore) Close() error {
	return s.db.Close()
}

func (s *LayoutStore) GetActiveLayout(window string) (map[string]hyprboard.Layout, error) {
	layouts, err := s.querier.GetLayoutsForApp(context.Background(), window)
	if err != nil {
		return nil, fmt.Errorf("sqlite select: %w", err)
	}

	ret := make(map[string]hyprboard.Layout)
	for _, layout := range layouts {
		ret[layout.App] = hyprboard.Layout{
			Code:    layout.Code,
			Variant: layout.Variant,
		}
	}

	return ret, nil
}

func (s *LayoutStore) SetActiveLayout(window string, keyboard string, layout hyprboard.Layout) error {
	if err := s.querier.SetLayout(context.Background(), SetLayoutParams{
		App:     window,
		Device:  keyboard,
		Code:    layout.Code,
		Variant: layout.Variant,
	}); err != nil {
		return fmt.Errorf("sqlite update: %w", err)
	}

	return nil
}
