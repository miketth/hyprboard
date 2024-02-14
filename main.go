package main

import (
	"codeberg.org/miketth/hyprboard/pkg/hyprboard"
	"codeberg.org/miketth/hyprboard/pkg/hyprland"
	"codeberg.org/miketth/hyprboard/pkg/layoutstore/json"
	"codeberg.org/miketth/hyprboard/pkg/xkblayouts"
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/coreos/go-systemd/v22/daemon"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	err := run()
	switch {
	case errors.Is(err, context.Canceled):
		return
	case err != nil:
		log.Fatalf("error: %+v", err)
	}
}

func run() error {
	evdevXmlPath := flag.String("evdev-xml-path", "/usr/share/X11/xkb/rules/evdev.xml", "path to evdev.xml")
	flag.Parse()

	ctx := context.Background()
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	registry, err := xkblayouts.ParseLayouts(*evdevXmlPath)
	if err != nil {
		return fmt.Errorf("parse layouts: %w", err)
	}

	client, err := hyprland.Connect()
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer client.Close()

	hyprctl, err := hyprland.NewHyprctl()
	if err != nil {
		return fmt.Errorf("connect hyprctl: %w", err)
	}

	errChan := make(chan error)

	configPath, err := getConfigDir()
	if err != nil {
		return fmt.Errorf("get config dir: %w", err)
	}
	layoutStore, err := json.NewLayoutStore(configPath + "/layouts.json")
	if err != nil {
		return fmt.Errorf("create layout store: %w", err)
	}
	go func() {
		err := layoutStore.SaveLooper(ctx)
		if err != nil {
			errChan <- fmt.Errorf("save looper: %w", err)
		}
	}()

	sw := hyprboard.NewSwitcher(client, hyprctl, registry, layoutStore)

	// notify systemd that we're ready
	// don't care about errors here; people might not be using systemd
	_, _ = daemon.SdNotify(false, daemon.SdNotifyReady)

	log.Println("started hyprboard")

	go func() {
		err := sw.ProcessLines(ctx)
		if err != nil {
			errChan <- fmt.Errorf("process lines: %w", err)
		}
	}()
	return <-errChan
}

func getConfigDir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("get user config dir: %w", err)
	}

	hyprboardConfigDir := dir + "/hyprboard"
	err = os.MkdirAll(hyprboardConfigDir, 0755)
	if err != nil {
		return "", fmt.Errorf("create hyprboard config dir: %w", err)
	}

	return hyprboardConfigDir, nil
}
