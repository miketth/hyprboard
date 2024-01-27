package hyprland

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/miketth/hyprboard/pkg/hyprboard"
	"os/exec"
	"regexp"
	"strings"
)

type Hyprctl struct {
	Path string
}

var (
	ErrIndexOutOfRange = errors.New("index out of range")
	ErrDeviceNotFound  = errors.New("device not found")
)

var errorMapper = map[*regexp.Regexp]error{
	regexp.MustCompile(`ok`):                        nil,
	regexp.MustCompile(`layout idx out of range.*`): ErrIndexOutOfRange,
	regexp.MustCompile(`device not found`):          ErrDeviceNotFound,
}

func (h Hyprctl) runCommand(args ...string) (string, error) {
	var stdout bytes.Buffer

	path := h.Path
	if path == "" {
		path = "hyprctl"
	}

	cmd := exec.Command(path, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout

	err := cmd.Run()
	outStr := strings.TrimSpace(stdout.String())
	if err != nil {
		return "", fmt.Errorf("hyprctl: %w, stdout: %s", err, outStr)
	}

	return outStr, nil
}

func (h Hyprctl) SwitchToLayout(keyboard string, idx int) error {
	outStr, err := h.runCommand("switchxkblayout", "--", keyboard, fmt.Sprintf("%d", idx))
	if err != nil {
		return err
	}

	for re, mappedErr := range errorMapper {
		if re.MatchString(outStr) {
			return mappedErr
		}
	}

	return fmt.Errorf("unknown hyprctl error: %s", outStr)
}

type keyboard struct {
	Name         string `json:"name"`
	Layout       string `json:"layout"`
	Variant      string `json:"variant"`
	Options      string `json:"options"`
	ActiveKeymap string `json:"active_keymap"`
}

type devices struct {
	Keyboards []keyboard `json:"keyboards"`
}

func (h Hyprctl) GetKeyboards() ([]hyprboard.Keyboard, error) {
	outStr, err := h.runCommand("devices", "-j")
	if err != nil {
		return nil, err
	}

	var devs devices
	if err := json.Unmarshal([]byte(outStr), &devs); err != nil {
		return nil, fmt.Errorf("unmarshal: %w, (hyprctl: %s)", err, outStr)
	}

	keyboards := devs.Keyboards

	out := make([]hyprboard.Keyboard, 0, len(keyboards))
	for _, k := range keyboards {
		out = append(out, k.ToKeyboard())
	}

	return out, nil
}

func (k keyboard) ToKeyboard() hyprboard.Keyboard {
	return hyprboard.Keyboard{
		Name:     k.Name,
		Layouts:  strings.Split(k.Layout, ","),
		Variants: strings.Split(k.Variant, ","),
	}
}
