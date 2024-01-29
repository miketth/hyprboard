package hyprboard

import (
	"codeberg.org/miketth/hyprboard/pkg/xkblayouts"
	"context"
	"fmt"
	"log"
	"strings"
)

type Switcher struct {
	activeLayouts  map[string]map[string]Layout
	layoutIdxCache map[string]map[Layout]int
	activeWindow   string

	listener        EventListener
	switcher        KeyboardLayoutSwitcher
	possibleLayouts *xkblayouts.XkbConfigRegistry
}

func NewSwitcher(
	listener EventListener,
	switcher KeyboardLayoutSwitcher,
	possibleLayouts *xkblayouts.XkbConfigRegistry,
) *Switcher {
	return &Switcher{
		activeLayouts:   make(map[string]map[string]Layout),
		layoutIdxCache:  make(map[string]map[Layout]int),
		activeWindow:    "",
		listener:        listener,
		switcher:        switcher,
		possibleLayouts: possibleLayouts,
	}
}

func (s *Switcher) ProcessLines(ctx context.Context) error {
	for {
		resultCh := make(chan string)
		errCh := make(chan error)
		go func() {
			line, err := s.listener.ReadLine()
			if err != nil {
				errCh <- err
				return
			}
			resultCh <- line
		}()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case line := <-resultCh:
			err := s.processLine(line)
			if err != nil {
				return fmt.Errorf("process line: %w", err)
			}
		case err := <-errCh:
			return fmt.Errorf("get line: %w", err)
		}
	}
}

func (s *Switcher) processLine(line string) error {
	fields := strings.Split(line, ">>")
	if len(fields) < 2 {
		return fmt.Errorf("invalid line: %q", line)
	}

	evType := fields[0]
	evData := fields[1]
	switch evType {
	case "activelayout":
		return s.processLayoutChange(evData)
	case "activewindow":
		return s.processWindowChange(evData)
	}

	return nil
}

func (s *Switcher) processLayoutChange(data string) error {
	dataParts := strings.Split(data, ",")

	if len(dataParts) < 2 {
		return fmt.Errorf("invalid layout change data: %q", data)
	}

	keyboardName := dataParts[0]
	layoutName := strings.Join(dataParts[1:], ",")

	// get layout code and variant code
	layoutCode, variantCode := s.possibleLayouts.GetLayoutAndVariantFromPrettyName(layoutName)
	if layoutCode == "" {
		return fmt.Errorf("layout %q not found", layoutName)
	}

	layout := Layout{Code: layoutCode, Variant: variantCode}

	if s.activeLayouts[s.activeWindow] == nil {
		s.activeLayouts[s.activeWindow] = make(map[string]Layout)
	}

	s.activeLayouts[s.activeWindow][keyboardName] = layout
	return nil
}

func (s *Switcher) getLayoutIndexForDevice(device string, layout Layout) (int, error) {
	// get it from cache if possible
	if idx, ok := s.layoutIdxCache[device][layout]; ok {
		return idx, nil
	}

	// get device keyboards
	keyboards, err := s.switcher.GetKeyboards()
	if err != nil {
		return -1, fmt.Errorf("get keyboards: %w", err)
	}

	// find the keyboard
	var keyboard Keyboard
	for _, k := range keyboards {
		if k.Name == device {
			keyboard = k
		}
	}
	if len(keyboard.Layouts) == 0 {
		return -1, fmt.Errorf("keyboard %q not found", device)
	}

	for i := range keyboard.Layouts {
		thisLayout, thisVariant := keyboard.Layouts[i], keyboard.Variants[i]
		if thisLayout == layout.Code && thisVariant == layout.Variant {
			if s.layoutIdxCache[device] == nil {
				s.layoutIdxCache[device] = make(map[Layout]int)
			}

			s.layoutIdxCache[device][layout] = i
			return i, nil
		}
	}

	return -1, fmt.Errorf("layout %q not found for keyboard %q", layout, keyboard.Name)
}

func (s *Switcher) processWindowChange(data string) error {
	s.activeWindow = strings.Split(data, ",")[0]
	newLayout, found := s.activeLayouts[s.activeWindow]
	if !found {
		return nil
	}

	for device, layout := range newLayout {
		idx, err := s.getLayoutIndexForDevice(device, layout)
		if err != nil {
			return fmt.Errorf("get layout index: %w", err)
		}

		if err := s.switcher.SwitchToLayout(device, idx); err != nil {
			log.Printf("switch layout: %v", err)
		}
	}

	return nil
}
