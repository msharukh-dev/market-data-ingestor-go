package config

import (
	"os"
	"strconv"
	common "ws_ingestor/internal/app/common/exception_handler"

	"github.com/sirupsen/logrus"
)

type Config struct {
	WebSocketURL  string
	APIKey        string
	DatabaseURL   string
	BatchSize     int
	NumWorkers    int
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	WSServerAddr  string
}

func Load(logger *logrus.Logger) (Config, error) {
	numWorkers, err := strconv.Atoi(os.Getenv("WORKER_COUNT"))
	if err != nil {
		numWorkers = 10 //default 10 workers
	}
	batchSize, err := strconv.Atoi(os.Getenv("BATCH_SIZE"))
	if err != nil {
		batchSize = 100 // default
	}

	cfg := Config{
		WebSocketURL:  os.Getenv("WS_URL"),
		APIKey:        os.Getenv("WS_API_KEY"),
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		BatchSize:     batchSize,
		NumWorkers:    numWorkers,
		RedisAddr:     os.Getenv("REDIS_ADDR"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		RedisDB:       0,
		WSServerAddr:  os.Getenv("WS_SERVER_ADDR"),
	}

	if cfg.WSServerAddr == "" {
		cfg.WSServerAddr = ":8080"
	}

	if cfg.WebSocketURL == "" || cfg.APIKey == "" || cfg.DatabaseURL == "" {
		return cfg, common.NewCustomError(common.ErrConfigLoad, "Missing required environment variables", nil)
	}
	return cfg, nil
}
