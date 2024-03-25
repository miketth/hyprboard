package main

import (
	"codeberg.org/miketth/hyprboard/pkg/hyprboard"
	"codeberg.org/miketth/hyprboard/pkg/hyprland"
	"codeberg.org/miketth/hyprboard/pkg/layoutstore/json"
	"codeberg.org/miketth/hyprboard/pkg/layoutstore/memory"
	"codeberg.org/miketth/hyprboard/pkg/layoutstore/sqlite"
	"codeberg.org/miketth/hyprboard/pkg/xkblayouts"
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/adrg/xdg"
	"github.com/coreos/go-systemd/v22/daemon"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"log"
	"os"
	"os/signal"
	"path"
	"sync"
	"syscall"
	"time"
)

func main() {
	err := run()
	if err != nil {
		log.Fatalf("error: %+v", err)
	}
}

func run() error {
	stateFileDefault, err := getStateFile()
	if err != nil {
		return fmt.Errorf("get state file: %w", err)
	}

	evdevXmlPath := flag.String("evdev-xml-path", "/usr/share/X11/xkb/rules/evdev.xml", "path to evdev.xml")
	stateFile := flag.String("state-file", stateFileDefault, "path to persist state, can be *.json (JSON), *.db (sqlite), or - (memory)")
	debug := flag.Bool("debug", false, "enable debug logging")
	flag.Parse()

	log, err := newLogger(*debug)
	if err != nil {
		return fmt.Errorf("create logger: %w", err)
	}

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

	layoutStore, err := createLayoutStore(ctx, *stateFile)
	if err != nil {
		return fmt.Errorf("create layout store: %w", err)
	}
	if closer, ok := layoutStore.(io.Closer); ok {
		defer closer.Close()
	}

	sw := hyprboard.NewSwitcher(client, hyprctl, registry, layoutStore, log)

	log.Info("started hyprboard")

	errChan := make(chan error, 3)
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		err := sw.ProcessLines(ctx)
		if err != nil {
			errChan <- fmt.Errorf("process lines: %w", err)
		}
	}()

	go func() {
		defer wg.Done()
		err := systemdNotifyLoop(ctx)
		if err != nil {
			errChan <- fmt.Errorf("systemd notify: %w", err)
		}
	}()

	err = <-errChan
	switch {
	case errors.Is(err, context.Canceled):
		log.Info("shutting down")
		wg.Wait()
		return nil
	case err != nil:
		return err
	}

	return nil
}

func createLayoutStore(ctx context.Context, filename string) (hyprboard.ActiveLayoutStore, error) {
	extension := path.Ext(filename)

	var layoutStore hyprboard.ActiveLayoutStore
	var err error
	switch {
	case extension == ".db":
		layoutStore, err = sqlite.NewLayoutStore(filename)
	case extension == ".json":
		layoutStore, err = json.NewLayoutStore(ctx, filename)
	case filename == "-":
		layoutStore = memory.NewLayoutStore()
	default:
		err = fmt.Errorf("unknown file extension for state storage: %q", extension)
	}

	if err != nil {
		return nil, fmt.Errorf("create layout store: %w", err)
	}

	return layoutStore, nil
}

func systemdNotifyLoop(ctx context.Context) error {
	// tell systemd that we're ready
	supported, err := daemon.SdNotify(false, daemon.SdNotifyReady)
	if err != nil {
		return fmt.Errorf("notify systemd: %w", err)
	}
	if !supported {
		return nil
	}

	// set funky message
	_, _ = daemon.SdNotify(false, "STATUS=Wildly switching keyboard layouts! 🤖")

	// notify watchdog
	t, err := daemon.SdWatchdogEnabled(false)
	if err != nil {
		return fmt.Errorf("check watchdog: %w", err)
	}
	// if watchdog is not enabled, we don't need to notify it
	if t == 0 {
		return nil
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-time.After(t / 2):
			_, err := daemon.SdNotify(false, daemon.SdNotifyWatchdog)
			if err != nil {
				return fmt.Errorf("notify watchdog: %w", err)
			}
		}
	}
}

func getStateFile() (string, error) {
	file, err := xdg.DataFile("hyprboard/data.db")
	if err != nil {
		return "", fmt.Errorf("get data dir: %w", err)
	}

	return file, nil
}

func newLogger(debug bool) (*zap.SugaredLogger, error) {
	loggerConfig := zap.NewDevelopmentConfig()

	loggerConfig.OutputPaths = []string{"stdout"}
	loggerConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	if debug {
		loggerConfig.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	} else {
		loggerConfig.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}

	logger, err := loggerConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("build logger: %w", err)
	}

	return logger.Sugar(), nil
}
