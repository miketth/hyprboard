package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/miketth/hyprboard/pkg/hyprboard"
	"github.com/miketth/hyprboard/pkg/hyprland"
	"github.com/miketth/hyprboard/pkg/xkblayouts"
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

	sw := hyprboard.NewSwitcher(client, hyprland.Hyprctl{}, registry)

	for {
		err = sw.ProcessLines(ctx)
		if errors.Is(err, context.Canceled) {
			fmt.Println("exiting gracefully...")
			return nil
		}
		if err != nil {
			return fmt.Errorf("process lines: %w", err)
		}
	}
}
