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
