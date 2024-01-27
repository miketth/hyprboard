package hyprland

import (
	"bufio"
	"errors"
	"fmt"
	"net"
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
	conn, reader, err := connect(Socket2)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	return &Client{conn: conn, reader: reader}, nil
}
