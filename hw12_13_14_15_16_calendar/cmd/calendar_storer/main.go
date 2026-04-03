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
	defer consumer.Close()

	st := storer.New(consumer, store, *lgr)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := st.Run(ctx); err != nil {
		lgr.Error("Storer error: " + err.Error())
		os.Exit(1)
	}
}

func loadConfig(path string) (*Config, error) {
	var cfg Config
	_, err := toml.DecodeFile(path, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
