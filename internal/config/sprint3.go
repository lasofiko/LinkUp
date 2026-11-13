package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

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
	CommandTimeout time.Duration
	ProfileBaseURL string
	ProfileTimeout time.Duration
	ProfileMaxBody int64
}

func Load() (Config, error) {
	linkUp, err := loadHTTP("LINKUP", ":8080")
	if err != nil {
		return Config{}, err
	}

	profileAPI, err := loadHTTP("PROFILE_API", "127.0.0.1:8081")
	if err != nil {
		return Config{}, err
	}

	baseURL := envString("PROFILE_BASE_URL", "http://"+profileAPI.Address)
	if _, err := parseEnvURL(baseURL); err != nil {
		return Config{}, fmt.Errorf("PROFILE_BASE_URL: %w", err)
	}

	profileTimeout, err := envDuration("PROFILE_TIMEOUT", 500*time.Millisecond)
	if err != nil {
		return Config{}, err
	}

	profileMaxBody, err := envInt64("PROFILE_MAX_BODY", 64<<10)
	if err != nil {
		return Config{}, err
	}

	commandTimeout, err := envDuration("LINKUP_COMMAND_TIMEOUT", 800*time.Millisecond)
	if err != nil {
		return Config{}, err
	}

	return Config{
		LinkUp:         linkUp,
		ProfileAPI:     profileAPI,
		CommandTimeout: commandTimeout,
		ProfileBaseURL: baseURL,
		ProfileTimeout: profileTimeout,
		ProfileMaxBody: profileMaxBody,
	}, nil
}

func loadHTTP(prefix, defaultAddress string) (HTTPConfig, error) {
	address := envString(prefix+"_ADDR", defaultAddress)
	if strings.TrimSpace(address) == "" {
		return HTTPConfig{}, fmt.Errorf("%s_ADDR: address must not be empty", prefix)
	}

	readHeaderTimeout, err := envDuration(prefix+"_READ_HEADER_TIMEOUT", 2*time.Second)
	if err != nil {
		return HTTPConfig{}, err
	}

	readTimeout, err := envDuration(prefix+"_READ_TIMEOUT", 5*time.Second)
	if err != nil {
		return HTTPConfig{}, err
	}

	writeTimeout, err := envDuration(prefix+"_WRITE_TIMEOUT", 5*time.Second)
	if err != nil {
		return HTTPConfig{}, err
	}

	idleTimeout, err := envDuration(prefix+"_IDLE_TIMEOUT", 30*time.Second)
	if err != nil {
		return HTTPConfig{}, err
	}

	shutdownTimeout, err := envDuration(prefix+"_SHUTDOWN_TIMEOUT", 10*time.Second)
	if err != nil {
		return HTTPConfig{}, err
	}

	return HTTPConfig{
		Address:           address,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		ShutdownTimeout:   shutdownTimeout,
	}, nil
}

func envString(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envDuration(key string, fallback time.Duration) (time.Duration, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", key, err)
	}
	if duration <= 0 {
		return 0, fmt.Errorf("%s: duration must be positive, got %s", key, value)
	}
	return duration, nil
}

func envInt64(key string, fallback int64) (int64, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", key, err)
	}
	if parsed <= 0 {
		return 0, fmt.Errorf("%s: value must be positive, got %s", key, value)
	}
	return parsed, nil
}

func parseEnvURL(raw string) (string, error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("invalid URL scheme %q: want http or https", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("invalid URL: host is empty")
	}
	return raw, nil
}
