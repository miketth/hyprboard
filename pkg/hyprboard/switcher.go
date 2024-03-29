package hyprboard

import (
	"codeberg.org/miketth/hyprboard/pkg/xkblayouts"
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"strings"
)

type Switcher struct {
	activeLayouts  ActiveLayoutStore
	layoutIdxCache map[string]map[Layout]int
	activeWindow   string

	listener        EventListener
	switcher        KeyboardLayoutSwitcher
	possibleLayouts *xkblayouts.XkbConfigRegistry
	log             *zap.SugaredLogger
}

func NewSwitcher(
	listener EventListener,
	switcher KeyboardLayoutSwitcher,
	possibleLayouts *xkblayouts.XkbConfigRegistry,
	activeLayoutStore ActiveLayoutStore,
	log *zap.SugaredLogger,
) *Switcher {
	return &Switcher{
		activeLayouts:   activeLayoutStore,
		layoutIdxCache:  make(map[string]map[Layout]int),
		activeWindow:    "",
		listener:        listener,
		switcher:        switcher,
		possibleLayouts: possibleLayouts,
		log:             log,
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

	err := s.activeLayouts.SetActiveLayout(s.activeWindow, keyboardName, layout)
	if err != nil {
		return fmt.Errorf("save active layout: %w", err)
	}

	return nil
}

var (
	errKeyboardNotFound = errors.New("keyboard not found")
	errLayoutNotFound   = errors.New("layout not found")
)

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
		return -1, fmt.Errorf("%w (%q)", errKeyboardNotFound, device)
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

	return -1, fmt.Errorf("%w (%q) for keyboard %q", errLayoutNotFound, layout, keyboard.Name)
}

func (s *Switcher) processWindowChange(data string) error {
	s.activeWindow = strings.Split(data, ",")[0]
	newLayout, err := s.activeLayouts.GetActiveLayout(s.activeWindow)
	if err != nil {
		return fmt.Errorf("get active layout: %w", err)
	}
	if newLayout == nil {
		return nil
	}

	for device, layout := range newLayout {
		idx, err := s.getLayoutIndexForDevice(device, layout)
		switch {
		case errors.Is(err, errKeyboardNotFound):
			continue
		case errors.Is(err, errLayoutNotFound):
			s.log.Warnf("get layout index: %v", err)
			continue
		case err != nil:
			return fmt.Errorf("get layout index: %w", err)
		}

		if err := s.switcher.SwitchToLayout(device, idx); err != nil {
			s.log.Warnf("switch layout: %v", err)
			continue
		}
	}

	return nil
}
