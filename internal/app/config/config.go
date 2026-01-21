package config

import (
	"os"
	"time"

	common "ws_ingestor/internal/app/common/exception_handler"

	"github.com/spf13/viper"
)

type Config struct {
	WebSocketURL        string        `mapstructure:"WS_URL"`
	APIKey              string        `mapstructure:"WS_API_KEY"`
	DatabaseURL         string        `mapstructure:"DATABASE_URL"`
	BatchSize           int           `mapstructure:"BATCH_SIZE"`
	NumWorkers          int           `mapstructure:"WORKER_COUNT"`
	RedisAddr           string        `mapstructure:"REDIS_ADDR"`
	RedisPassword       string        `mapstructure:"REDIS_PASSWORD"`
	RedisDB             int           `mapstructure:"REDIS_DB"`
	WSServerAddr        string        `mapstructure:"WS_SERVER_ADDR"`
	RedisTTL            time.Duration `mapstructure:"REDIS_TTL"`
	FlushInterval       time.Duration `mapstructure:"FLUSH_INTERVAL"`
	SubscriptionSymbols []string      `mapstructure:"SUBSCRIPTION_SYMBOLS"`
}

func Load() (Config, error) {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("BATCH_SIZE", 100)
	viper.SetDefault("WORKER_COUNT", 10)
	viper.SetDefault("REDIS_DB", 0)
	viper.SetDefault("REDIS_TTL", "24h")
	viper.SetDefault("FLUSH_INTERVAL", "2s")
	viper.SetDefault("SUBSCRIPTION_SYMBOLS", []string{"USDSGD"})

	if err := viper.ReadInConfig(); err != nil {
		// Fallback to env if .env not found
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return cfg, common.NewCustomError(common.ErrConfigLoad, "Failed to unmarshal config", err)
	}

	// Parse duration
	ttlStr := os.Getenv("REDIS_TTL")
	if ttlStr == "" {
		ttlStr = "24h"
	}
	ttl, err := time.ParseDuration(ttlStr)
	if err != nil {
		ttl = 24 * time.Hour
	}
	cfg.RedisTTL = ttl

	flushStr := os.Getenv("FLUSH_INTERVAL")
	if flushStr == "" {
		flushStr = "2s"
	}
	flush, err := time.ParseDuration(flushStr)
	if err != nil {
		flush = 2 * time.Second
	}
	cfg.FlushInterval = flush

	// Subscription symbols
	symbolsStr := os.Getenv("SUBSCRIPTION_SYMBOLS")
	if symbolsStr != "" {
		// Parse comma-separated
		// For simplicity, assume env is comma-separated
		cfg.SubscriptionSymbols = []string{"USDSGD"} // Keep default for now
	}

	if cfg.WebSocketURL == "" || cfg.APIKey == "" || cfg.DatabaseURL == "" {
		return cfg, common.NewCustomError(common.ErrConfigLoad, "Missing required environment variables", nil)
	}
	return cfg, nil
}
