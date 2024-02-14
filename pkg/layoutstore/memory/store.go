package memory

import "codeberg.org/miketth/hyprboard/pkg/hyprboard"

type LayoutStore struct {
	layouts map[string]map[string]hyprboard.Layout
}

func NewLayoutStore() *LayoutStore {
	return &LayoutStore{
		layouts: make(map[string]map[string]hyprboard.Layout),
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
	return nil
}
