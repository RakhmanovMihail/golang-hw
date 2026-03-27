package main

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
	logger "github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/logger"
	"log"
)

// Локальный тип для TOML десериализации
type tomlLevel logger.Level

func (l *tomlLevel) UnmarshalTOML(data []byte) error {
	level := strings.Trim(string(data), `"`)
	switch level {
	case string(logger.LevelDebug), string(logger.LevelInfo),
		string(logger.LevelWarn), string(logger.LevelError):
		*l = tomlLevel(level)
		return nil
	default:
		return fmt.Errorf("invalid log level: %s", level)
	}
}

type Config struct {
	Logger  LoggerConf    `toml:"logger"`
	Api     ApiConfig     `toml:"api"`
	Storage StorageConfig `toml:"storage"`
}

// Конвертеры уровней
func (c *Config) LoggerLevel() logger.Level {
	return logger.Level(c.Logger.Level)
}

func (c *Config) ApiLevel() logger.Level {
	return logger.Level(c.Api.Level)
}

type ApiConfig struct {
	Level tomlLevel `toml:"level"`
	Host  string    `toml:"host"`
	Port  string    `toml:"port"`
}

type LoggerConf struct {
	Level tomlLevel `toml:"level"`
}

type StorageMode string

const (
	StorageModeMemory StorageMode = "memory"
	StorageModeSQL    StorageMode = "sql"
)

type StorageConfig struct {
	Mode StorageMode `toml:"mode"`
	DSN  string      `toml:"dsn,omitempty"` // для SQL
}

func NewConfig(path string) *Config {
	var config Config
	if _, err := toml.DecodeFile(path, &config); err != nil {
		log.Fatal(err)
	}
	return &config
}
