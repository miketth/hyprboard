package hyprboard

type EventListener interface {
	ReadLine() (string, error)
}

type KeyboardLayoutSwitcher interface {
	GetKeyboards() ([]Keyboard, error)
	SwitchToLayout(keyboard string, idx int) error
}

type Keyboard struct {
	Name     string
	Layouts  []string
	Variants []string
}

type Layout struct {
	Code    string
	Variant string
}

type ActiveLayoutStore interface {
	GetActiveLayout(window string) (map[string]Layout, error)
	SetActiveLayout(window string, keyboard string, layout Layout) error
}
