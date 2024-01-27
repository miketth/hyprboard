package hyprland

import (
	"fmt"
	"net"
	"os"
)

func connect(sock socketType) (net.Conn, error) {
	socketPath, err := getSocketPath(sock)
	if err != nil {
		return nil, fmt.Errorf("get socket path: %w", err)
	}

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("dial: %w", err)
	}

	return conn, nil
}

type socketType int

const (
	Hyperctl socketType = iota
	Socket2
)

func getSocketPath(sock socketType) (string, error) {
	signature := os.Getenv("HYPRLAND_INSTANCE_SIGNATURE")
	if signature == "" {
		return "", fmt.Errorf("HYPRLAND_INSTANCE_SIGNATURE is not set, %w", ErrNotRunning)
	}

	switch sock {
	case Hyperctl:
		return fmt.Sprintf("/tmp/hypr/%s/.socket.sock", signature), nil
	case Socket2:
		return fmt.Sprintf("/tmp/hypr/%s/.socket2.sock", signature), nil
	}

	return "", fmt.Errorf("unknown socket type: %d", sock)
}
