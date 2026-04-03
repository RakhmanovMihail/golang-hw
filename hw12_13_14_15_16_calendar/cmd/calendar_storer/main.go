package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/BurntSushi/toml"
	"github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/kafka"
	"github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/storage/sql"
	"github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/storer"
)

type Config struct {
	Logger  LoggerConfig  `toml:"logger"`
	Storage StorageConfig `toml:"storage"`
	Kafka   KafkaConfig   `toml:"kafka"`
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
	GroupID string   `toml:"group_id"`
}

func main() {
	configPath := flag.String("config", "configs/storer_config.toml", "Path to config file")
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

	consumer, err := kafka.NewSaramaConsumer(cfg.Kafka.Brokers, cfg.Kafka.Topic, cfg.Kafka.GroupID)
	if err != nil {
		lgr.Error("Failed to create Kafka consumer: " + err.Error())
		os.Exit(1)
	}

	st := storer.New(consumer, store, *lgr)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Run storer in goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- st.Run(ctx)
	}()

	// Wait for context cancellation or error
	select {
	case <-ctx.Done():
		lgr.Info("shutting down storer...")
	case err := <-errCh:
		if err != nil {
			lgr.Error("Storer error: " + err.Error())
		}
	}

	// Graceful shutdown
	if err := consumer.Close(); err != nil {
		lgr.Error("Failed to close Kafka consumer: " + err.Error())
	}
	if err := store.Close(context.Background()); err != nil {
		lgr.Error("Failed to close storage: " + err.Error())
	}

	lgr.Info("storer stopped")
}

func loadConfig(path string) (*Config, error) {
	var cfg Config
	_, err := toml.DecodeFile(path, &cfg)
	if err != nil {
		return nil, err
	}

	// Override with environment variables if set
	if brokers := os.Getenv("KAFKA_BROKERS"); brokers != "" {
		cfg.Kafka.Brokers = []string{brokers}
	}
	if topic := os.Getenv("KAFKA_TOPIC"); topic != "" {
		cfg.Kafka.Topic = topic
	}
	if groupID := os.Getenv("KAFKA_GROUP_ID"); groupID != "" {
		cfg.Kafka.GroupID = groupID
	}
	if dsn := os.Getenv("DB_DSN"); dsn != "" {
		cfg.Storage.DSN = dsn
	}
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		cfg.Logger.Level = level
	}

	return &cfg, nil
}
