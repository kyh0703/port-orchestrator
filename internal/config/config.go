package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPAddr             string
	APIGRPCAddr          string
	APIReportMaxAttempts int
	APIReportRetryDelay  time.Duration
	APIReportTimeout     time.Duration
	ShutdownTimeout      time.Duration
}

func Load() Config {
	return Config{
		HTTPAddr:             envString("ORCHESTRATOR_HTTP_ADDR", ":8080"),
		APIGRPCAddr:          envString("PORT_API_GRPC_ADDR", "localhost:50051"),
		APIReportMaxAttempts: envInt("PORT_API_REPORT_MAX_ATTEMPTS", 3),
		APIReportRetryDelay:  envDuration("PORT_API_REPORT_RETRY_DELAY", 200*time.Millisecond),
		APIReportTimeout:     envDuration("PORT_API_REPORT_TIMEOUT", 5*time.Second),
		ShutdownTimeout:      envDuration("ORCHESTRATOR_SHUTDOWN_TIMEOUT", 10*time.Second),
	}
}

func (c Config) Validate() error {
	return nil
}

func envString(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return fallback
	}
	return parsed
}

func envDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
