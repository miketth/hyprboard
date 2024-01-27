package hyprland

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/miketth/hyprboard/pkg/hyprboard"
	"net"
	"syscall"
)

type Hyprctl struct {
	conn   net.Conn
	reader *bufio.Reader
}

func (c *Hyprctl) Close() error {
	return c.conn.Close()
}

func ConnectHyprctl() (*Hyprctl, error) {
	conn, reader, err := connect(Hyperctl)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	return &Hyprctl{conn: conn, reader: reader}, nil
}

func (c *Hyprctl) SwitchToLayout(keyboard string, idx int) error {
	err := c.makeRequest(fmt.Sprintf("switchxkblayout %s %d", keyboard, idx), "")
	if err != nil {
		return err
	}

	resp, err := c.readResponse()
	if err != nil {
		return fmt.Errorf("read response from hyprctl socket: %w", err)
	}

	str := string(resp)
	if str != "ok" {
		return fmt.Errorf("hyprctl: %s", str)
	}

	return nil
}

func (c *Hyprctl) GetKeyboards() ([]hyprboard.Keyboard, error) {
	err := c.makeRequest("devices", "j")
	if err != nil {
		return nil, err
	}

	resp, err := c.readResponse()
	if err != nil {
		return nil, fmt.Errorf("read response from hyprctl socket: %w", err)
	}

	var devs devices
	if err := json.Unmarshal(resp, &devs); err != nil {
		return nil, fmt.Errorf("unmarshal devices: %w, (hyprctl: %s)", err, string(resp))
	}

	keyboards := devs.Keyboards
	out := make([]hyprboard.Keyboard, 0, len(keyboards))
	for _, k := range keyboards {
		out = append(out, k.ToKeyboard())
	}

	return out, nil
}

func (c *Hyprctl) makeRequest(request string, args string) error {
	_, err := c.conn.Write([]byte(fmt.Sprintf("%s/%s", args, request)))
	if errors.Is(err, syscall.EPIPE) { // broken pipe
		err = c.reconnect()
		if err != nil {
			return fmt.Errorf("reconnect: %w", err)
		}

		_, err = c.conn.Write([]byte(fmt.Sprintf("%s/%s", args, request)))
	}
	if err != nil {
		return fmt.Errorf("write to hyprctl socket: %w", err)
	}

	return nil
}

func (c *Hyprctl) reconnect() error {
	conn, reader, err := connect(Hyperctl)
	if err != nil {
		return err
	}

	c.conn = conn
	c.reader = reader

	return nil
}

const readBufferSize = 8192

func (c *Hyprctl) readResponse() ([]byte, error) {
	buf := make([]byte, readBufferSize)
	n, err := c.reader.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("read from hyprctl socket: %w", err)
	}

	for n == readBufferSize {
		tmpBuf := make([]byte, readBufferSize)
		tmpN, err := c.reader.Read(tmpBuf)
		if err != nil {
			return nil, fmt.Errorf("chunked read from hyprctl socket: %w", err)
		}

		buf = append(buf, tmpBuf[:tmpN]...)
		n += tmpN
	}

	return buf[:n], nil
}
