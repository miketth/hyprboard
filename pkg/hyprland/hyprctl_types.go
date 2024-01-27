package hyprland

import (
	"github.com/miketth/hyprboard/pkg/hyprboard"
	"strings"
)

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

func (k keyboard) ToKeyboard() hyprboard.Keyboard {
	return hyprboard.Keyboard{
		Name:     k.Name,
		Layouts:  strings.Split(k.Layout, ","),
		Variants: strings.Split(k.Variant, ","),
	}
}
