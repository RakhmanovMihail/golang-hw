package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
	app "github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/app"
	loggerpkg "github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/logger"
	internalhttp "github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/server/http"
	storage "github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/storage"
	memorystorage "github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/storage/memory"
	psqlstorage "github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/storage/sql"
	migrations "github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/migrations"
	"github.com/pressly/goose/v3"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "configs/config.toml", "Path to configuration file")
}

func main() {
	flag.Parse()
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Println("calendar version develop")
		return
	}

	if err := run(); err != nil {
		log.Fatalf("fatal error: %v", err)
	}
}

func run() error {
	cfg := &Config{}
	_, err := toml.DecodeFile(configPath, cfg)
	if err != nil {
		return fmt.Errorf("config %s: %w", configPath, err)
	}

	logg := loggerpkg.New(cfg.LoggerLevel())

	var store storage.Storage
	switch cfg.Storage.Mode {
	case StorageModeMemory:
		store = memorystorage.New()
	case StorageModeSQL:
		// Миграции ПЕРЕД созданием storage!
		if err := runMigrations(logg, cfg.Storage.DSN); err != nil {
			return fmt.Errorf("migrations: %w", err)
		}

		storeSQL, err := psqlstorage.New(cfg.Storage.DSN)
		if err != nil {
			return fmt.Errorf("postgres: %w", err)
		}
		store = storeSQL
	default:
		return fmt.Errorf("unknown storage: %s", cfg.Storage.Mode)
	}

	calendar := app.New(*logg, store)
	server := internalhttp.NewServer(logg, calendar, fmt.Sprintf("%s:%s", cfg.API.Host, cfg.API.Port))

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	go func() {
		<-ctx.Done()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer shutdownCancel()
		if err := server.Stop(shutdownCtx); err != nil {
			logg.Error(fmt.Sprintf("server stop: %v", err))
		}
	}()

	logg.Info("calendar is running...")
	if err := server.Start(ctx); err != nil {
		return fmt.Errorf("server start: %w", err)
	}

	return nil
}

func runMigrations(logger *loggerpkg.Logger, dsn string) error {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			logger.Error(fmt.Sprintf("db close: %v", closeErr))
		}
	}()

	goose.SetBaseFS(migrations.FS)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}

	if err := goose.Up(db, "."); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}

	logger.Info("migrations applied successfully")
	return nil
}
