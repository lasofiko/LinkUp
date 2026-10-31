package config

import (
	"errors"
	"time"
)

var ErrNotImplemented = errors.New("sprint 3 configuration is not implemented")

type HTTPConfig struct {
	Address           string
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
}

type Config struct {
	LinkUp         HTTPConfig
	ProfileAPI     HTTPConfig
	ProfileBaseURL string
	ProfileTimeout time.Duration
	ProfileMaxBody int64
}

func Load() (Config, error) { return Config{}, ErrNotImplemented }
