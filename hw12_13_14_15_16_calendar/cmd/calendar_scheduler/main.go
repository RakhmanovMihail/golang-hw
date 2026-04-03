package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/kafka"
	"github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/scheduler"
	"github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/storage/sql"
)

type Config struct {
	Logger   LoggerConfig   `toml:"logger"`
	Storage  StorageConfig  `toml:"storage"`
	Kafka    KafkaConfig    `toml:"kafka"`
	Schedule ScheduleConfig `toml:"schedule"`
}

type LoggerConfig struct {
	Level string `toml:"level"`
}

type StorageConfig struct {
	DSN string `toml:"dsn"`
}

type KafkaConfig struct {
	Brokers []string `toml:"brokers"`
	Topic   string   `toml:"topic"`
}

type ScheduleConfig struct {
	CheckInterval string `toml:"check_interval"`
	CleanupOlder  string `toml:"cleanup_older"`
}

func main() {
	configPath := flag.String("config", "configs/scheduler_config.toml", "Path to config file")
	flag.Parse()

	cfg, err := loadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	lgr := logger.New(logger.Level(cfg.Logger.Level))

	store, err := sql.New(cfg.Storage.DSN)
	if err != nil {
		lgr.Error("Failed to create SQL storage: " + err.Error())
		os.Exit(1)
	}
	defer store.Close(context.Background())

	producer, err := kafka.NewSaramaProducer(cfg.Kafka.Brokers)
	if err != nil {
		lgr.Error("Failed to create Kafka producer: " + err.Error())
		os.Exit(1)
	}

	checkInterval, err := time.ParseDuration(cfg.Schedule.CheckInterval)
	if err != nil {
		lgr.Error("Invalid check_interval: " + err.Error())
		os.Exit(1)
	}

	cleanupOlder, err := time.ParseDuration(cfg.Schedule.CleanupOlder)
	if err != nil {
		lgr.Error("Invalid cleanup_older: " + err.Error())
		os.Exit(1)
	}

	sched := scheduler.New(store, producer, *lgr, cfg.Kafka.Topic, checkInterval, cleanupOlder)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Run scheduler in goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- sched.Run(ctx)
	}()

	// Wait for context cancellation or error
	select {
	case <-ctx.Done():
		lgr.Info("shutting down scheduler...")
	case err := <-errCh:
		if err != nil {
			lgr.Error("Scheduler error: " + err.Error())
		}
	}

	// Graceful shutdown
	if err := producer.Close(); err != nil {
		lgr.Error("Failed to close Kafka producer: " + err.Error())
	}
	if err := store.Close(context.Background()); err != nil {
		lgr.Error("Failed to close storage: " + err.Error())
	}

	lgr.Info("scheduler stopped")
}

func loadConfig(path string) (*Config, error) {
	var cfg Config
	_, err := toml.DecodeFile(path, &cfg)
	if err != nil {
		return nil, err
	}

	// Override with environment variables if set
	if brokers := os.Getenv("KAFKA_BROKERS"); brokers != "" {
		cfg.Kafka.Brokers = strings.Split(brokers, ",")
	}
	if topic := os.Getenv("KAFKA_TOPIC"); topic != "" {
		cfg.Kafka.Topic = topic
	}
	if dsn := os.Getenv("DB_DSN"); dsn != "" {
		cfg.Storage.DSN = dsn
	}
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		cfg.Logger.Level = level
	}

	return &cfg, nil
}
