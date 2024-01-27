package hyprland

import (
	"bytes"
	"codeberg.org/miketth/hyprboard/pkg/hyprboard"
	"encoding/json"
	"fmt"
	"io"
	"net"
)

type Hyprctl struct{}

func NewHyprctl() (*Hyprctl, error) {
	return &Hyprctl{}, nil
}

func (c *Hyprctl) SwitchToLayout(keyboard string, idx int) error {
	conn, err := c.makeRequest(fmt.Sprintf("switchxkblayout %s %d", keyboard, idx), "")
	if err != nil {
		return err
	}
	defer conn.Close()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, conn)
	if err != nil {
		return fmt.Errorf("read response from hyprctl socket: %w", err)
	}

	if buf.String() != "ok" {
		return fmt.Errorf("hyprctl: %s", buf.String())
	}

	return nil
}

func (c *Hyprctl) GetKeyboards() ([]hyprboard.Keyboard, error) {
	conn, err := c.makeRequest("devices", "j")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	dec := json.NewDecoder(conn)

	var devs devices
	if err := dec.Decode(&devs); err != nil {
		return nil, fmt.Errorf("unmarshal devices: %w", err)
	}

	keyboards := devs.Keyboards
	out := make([]hyprboard.Keyboard, 0, len(keyboards))
	for _, k := range keyboards {
		out = append(out, k.ToKeyboard())
	}

	return out, nil
}

func (c *Hyprctl) makeRequest(request string, args string) (net.Conn, error) {
	conn, err := connect(Hyperctl)
	_, err = conn.Write([]byte(fmt.Sprintf("%s/%s", args, request)))
	if err != nil {
		return nil, fmt.Errorf("write to hyprctl socket: %w", err)
	}

	return conn, nil
}

const readBufferSize = 8192

func readResponse(reader io.Reader) ([]byte, error) {
	buf := make([]byte, readBufferSize)
	n, err := reader.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("read from hyprctl socket: %w", err)
	}

	for n == readBufferSize {
		tmpBuf := make([]byte, readBufferSize)
		tmpN, err := reader.Read(tmpBuf)
		if err != nil {
			return nil, fmt.Errorf("chunked read from hyprctl socket: %w", err)
		}

		buf = append(buf, tmpBuf[:tmpN]...)
		n += tmpN
	}

	return buf[:n], nil
}
