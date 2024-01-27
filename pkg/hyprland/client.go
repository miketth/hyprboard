package hyprland

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
)

var ErrNotRunning = errors.New("hyprland might not be running")

type Client struct {
	conn   net.Conn
	reader *bufio.Reader
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) ReadLine() (string, error) {
	str, err := c.reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("read from hypr socket: %w", err)
	}
	return strings.TrimSuffix(str, "\n"), nil
}

func Connect() (*Client, error) {
	socketPath, err := GetSocketPath()
	if err != nil {
		return nil, fmt.Errorf("get socket path: %w", err)
	}

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("dial: %w", err)
	}

	return &Client{conn: conn, reader: bufio.NewReader(conn)}, nil
}

func GetSocketPath() (string, error) {
	signature := os.Getenv("HYPRLAND_INSTANCE_SIGNATURE")
	if signature == "" {
		return "", fmt.Errorf("HYPRLAND_INSTANCE_SIGNATURE is not set, %w", ErrNotRunning)
	}

	return fmt.Sprintf("/tmp/hypr/%s/.socket2.sock", signature), nil
}
