package main

import (
	"codeberg.org/miketth/hyprboard/pkg/hyprboard"
	"codeberg.org/miketth/hyprboard/pkg/hyprland"
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
	if err := run(); err != nil {
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

	sw := hyprboard.NewSwitcher(client, hyprctl, registry)

	// notify systemd that we're ready
	// don't care about errors here; people might not be using systemd
	_, _ = daemon.SdNotify(false, daemon.SdNotifyReady)

	log.Println("started hyprboard")

	for {
		err = sw.ProcessLines(ctx)
		if errors.Is(err, context.Canceled) {
			log.Println("exiting gracefully...")
			return nil
		}
		if err != nil {
			return fmt.Errorf("process lines: %w", err)
		}
	}
}
