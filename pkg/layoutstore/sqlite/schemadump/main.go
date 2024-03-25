package main

import (
	"codeberg.org/miketth/hyprboard/pkg/layoutstore/sqlite"
	"codeberg.org/miketth/hyprboard/pkg/layoutstore/sqlite/migrations"
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"os"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("error: %+v", err)
	}
}

func run() error {
	path := flag.String("path", "", "path to dump the schema to")
	debug := flag.Bool("debug", false, "use debug level logging")
	flag.Parse()

	if *path == "" {
		return errors.New("missing -path flag")
	}

	log, err := newLogger(*debug)
	if err != nil {
		return fmt.Errorf("create logger: %w", err)
	}

	log.Info("creating empty database")
	db, err := sql.Open("sqlite3", "file:/dev/null?cache=shared&mode=memory")
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	log.Info("applying migrations")
	if err := migrations.Migrate(db, log); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	file, err := os.Create(*path)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer file.Close()

	log.Info("dumping schema")
	if err := dumpSchema(sqlite.New(db), file); err != nil {
		return fmt.Errorf("dump schema: %w", err)
	}

	return nil
}

func dumpSchema(db *sqlite.Queries, file *os.File) error {
	ctx := context.Background()

	tables, err := db.DumpTables(ctx)
	if err != nil {
		return fmt.Errorf("dump tables: %w", err)
	}

	rest, err := db.DumpRest(ctx)
	if err != nil {
		return fmt.Errorf("dump non-statements content: %w", err)
	}

	schema := append(tables, rest...)

	for _, statement := range schema {
		if statement == nil {
			continue
		}
		line := fmt.Sprintf("%s;\n\n", *statement)
		if _, err := file.WriteString(line); err != nil {
			return fmt.Errorf("write file: %w", err)
		}
	}

	if _, err := file.WriteString(sqliteMasterSchema); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
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

const sqliteMasterSchema = `
create table sqlite_master (
    type     text,
    name     text,
    tbl_name text,
    rootpage int,
    sql      text
);
`
