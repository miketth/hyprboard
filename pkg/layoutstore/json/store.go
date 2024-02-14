package json

import (
	"codeberg.org/miketth/hyprboard/pkg/hyprboard"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

type LayoutStore struct {
	layouts map[string]map[string]hyprboard.Layout
	file    *os.File
	lock    sync.Mutex
	dirty   bool
}

func NewLayoutStore(filename string) (*LayoutStore, error) {
	fileExists := true
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		fileExists = false
	}

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}

	store := &LayoutStore{
		layouts: make(map[string]map[string]hyprboard.Layout),
		file:    file,
		dirty:   true,
	}

	if fileExists {
		err = store.load()
		if err != nil {
			return nil, fmt.Errorf("load: %w", err)
		}

		store.dirty = false
	}

	return store, nil
}

func (s *LayoutStore) Close() error {
	return s.file.Close()
}

func (s *LayoutStore) load() error {
	s.lock.Lock()
	defer s.lock.Unlock()
	_, err := s.file.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("seek to start of file: %w", err)
	}

	dec := json.NewDecoder(s.file)
	err = dec.Decode(&s.layouts)
	if err != nil {
		return fmt.Errorf("decode json: %w", err)
	}

	return nil
}

func (s *LayoutStore) save() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if !s.dirty {
		return nil
	}

	_, err := s.file.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("seek to start of file: %w", err)
	}

	err = s.file.Truncate(0)
	if err != nil {
		return fmt.Errorf("truncate file: %w", err)
	}

	enc := json.NewEncoder(s.file)
	err = enc.Encode(s.layouts)
	if err != nil {
		return fmt.Errorf("encode json: %w", err)
	}

	s.dirty = false

	return nil
}

func (s *LayoutStore) SaveLooper(ctx context.Context) error {
	defer s.file.Close()

	for {
		select {
		case <-ctx.Done():
			err := s.save()
			if err != nil {
				return fmt.Errorf("save: %w", err)
			}

			return ctx.Err()
		case <-time.After(time.Minute):
			err := s.save()
			if err != nil {
				return fmt.Errorf("save: %w", err)
			}
		}
	}
}

func (s *LayoutStore) GetActiveLayout(window string) (map[string]hyprboard.Layout, error) {
	layouts, ok := s.layouts[window]
	if !ok {
		return nil, nil
	}
	return layouts, nil
}

func (s *LayoutStore) SetActiveLayout(window string, keyboard string, layout hyprboard.Layout) error {
	layouts, ok := s.layouts[window]
	if !ok {
		layouts = make(map[string]hyprboard.Layout)
		s.layouts[window] = layouts
	}
	layouts[keyboard] = layout
	s.dirty = true
	return nil
}
