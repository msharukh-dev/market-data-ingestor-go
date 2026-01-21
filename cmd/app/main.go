package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"ws_ingestor/cmd/processor"
	"ws_ingestor/internal/app/common/logger"
	"ws_ingestor/internal/app/config"
	"ws_ingestor/internal/app/models"
	"ws_ingestor/internal/app/services/storage"

	ws "ws_ingestor/internal/app/services/websocket"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			logger := logger.GetLogger()
			logger.WithField("panic", r).Fatal("Application panicked")
		}
	}()

	logger := logger.GetLogger()

	if err := godotenv.Load(); err != nil {
		logger.WithError(err).Fatal("No .env file found")
	}

	cfg, err := config.Load()
	if err != nil {
		logger.WithError(err).Fatal("Failed to load configuration")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	// Health and metrics endpoints
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	http.Handle("/metrics", promhttp.Handler())

	go func() {
		logger.Info("Starting metrics server on :9090")
		if err := http.ListenAndServe(":9090", nil); err != nil {
			logger.Fatal("Failed to start metrics server: ", err)
		}
	}()

	dataChan := make(chan models.MarketData, 10000) // Increased buffer for backpressure

	store, err := storage.NewPostgres(cfg.DatabaseURL)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize database")
	}
	defer store.Close()

	cache, err := storage.NewRedis(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize cache")
	}
	defer cache.Close()

	proc := processor.New(store, cache, dataChan, cfg.BatchSize, cfg.NumWorkers, cfg.RedisTTL, cfg.FlushInterval)
	go proc.Start(ctx)

	client := ws.New(cfg.WebSocketURL, cfg.APIKey, dataChan, cfg.SubscriptionSymbols)
	go client.Start(ctx)

	server := ws.NewServer(cfg.WSServerAddr, cache, store)
	go server.Start(ctx)

	<-sig
	logger.Info("Shutting down...")
	cancel()
}
