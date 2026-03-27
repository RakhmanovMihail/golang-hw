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
	"github.com/pressly/goose/v3"

	app "github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/app"
	loggerpkg "github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/logger"
	internalhttp "github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/server/http"
	storage "github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/storage"
	memorystorage "github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/storage/memory"
	migrations "github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/migrations"
	psqlstorage "github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/storage/sql"
	_ "github.com/lib/pq"
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

	cfg := &Config{}
	_, err := toml.DecodeFile(configPath, cfg)
	if err != nil {
		log.Fatalf("config %s: %v", configPath, err)
	}

	logg := loggerpkg.New(cfg.LoggerLevel())

	var store storage.Storage
	switch cfg.Storage.Mode {
	case "memory":
		store = memorystorage.New()
	case "postgres":
		// Миграции ПЕРЕД созданием storage!
		if err := runMigrations(logg, cfg.Storage.DSN); err != nil {
			logg.Error(fmt.Sprintf("migrations: %v", err))
			os.Exit(1)
		}

		storeSQL, err := psqlstorage.New(cfg.Storage.DSN)
		if err != nil {
			logg.Error(fmt.Sprintf("postgres: %v", err))
			os.Exit(1)
		}
		store = storeSQL
	default:
		logg.Error(fmt.Sprintf("unknown storage: %s", cfg.Storage.Mode))
		os.Exit(1)
	}

	calendar := app.New(*logg, store)
	server := internalhttp.NewServer(logg, calendar, fmt.Sprintf("%s:%s", cfg.Api.Host, cfg.Api.Port))

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
		logg.Error(fmt.Sprintf("server start: %v", err))
		os.Exit(1)
	}
}

func runMigrations(logger *loggerpkg.Logger, dsn string) error {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

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
