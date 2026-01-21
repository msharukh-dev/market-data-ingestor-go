package processor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"ws_ingestor/internal/app/common/logger"
	"ws_ingestor/internal/app/metrics"
	"ws_ingestor/internal/app/models"
	"ws_ingestor/internal/app/services/storage"

	"github.com/sirupsen/logrus"
)

type Processor struct {
	store         *storage.Store
	cache         *storage.CacheService
	in            <-chan models.MarketData
	batchSize     int
	numWorkers    int
	ttl           time.Duration
	flushInterval time.Duration
	logger        *logrus.Logger
}

func New(store *storage.Store, cache *storage.CacheService, in <-chan models.MarketData, batchSize int, numWorkers int, ttl time.Duration, flushInterval time.Duration) *Processor {
	return &Processor{
		store:         store,
		cache:         cache,
		in:            in,
		batchSize:     batchSize,
		numWorkers:    numWorkers,
		ttl:           ttl,
		flushInterval: flushInterval,
		logger:        logger.GetLogger(),
	}
}

func (p *Processor) Start(ctx context.Context) {
	wg := &sync.WaitGroup{}
	for i := 0; i < p.numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			p.worker(ctx)
		}()
	}
	wg.Wait()
}

func (p *Processor) worker(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			p.logger.WithField("panic", r).Error("Worker panicked")
		}
	}()
	batch := make([]models.MarketData, 0, p.batchSize)
	ticker := time.NewTicker(p.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			if len(batch) > 0 {
				p.flush(ctx, batch)
			}
			return
		case d := <-p.in:
			batch = append(batch, d)
			if len(batch) >= p.batchSize {
				p.flush(ctx, batch)
				batch = batch[:0]
			}
		case <-ticker.C:
			if len(batch) > 0 {
				p.flush(ctx, batch)
				batch = batch[:0]
			}
		}
	}
}

func (p *Processor) flush(ctx context.Context, batch []models.MarketData) {
	start := time.Now()
	const maxRetries = 3
	var err error

	// Retry store insert
	for i := 0; i < maxRetries; i++ {
		if err = p.store.InsertBatch(ctx, batch); err == nil {
			break
		}
		p.logger.Warn(fmt.Sprintf("Store insert failed (attempt %d/%d): %v", i+1, maxRetries, err))
		metrics.ErrorsTotal.WithLabelValues("store_insert").Inc()
		time.Sleep(time.Duration(i+1) * time.Second) // Exponential backoff
	}
	if err != nil {
		p.logger.Error(fmt.Sprintf("Store insert failed after retries: %v", err))
	}

	// Retry cache insert
	for i := 0; i < maxRetries; i++ {
		if err = p.cache.InsertBatch(ctx, batch, p.ttl); err == nil {
			break
		}
		p.logger.Warn(fmt.Sprintf("Cache insert failed (attempt %d/%d): %v", i+1, maxRetries, err))
		metrics.ErrorsTotal.WithLabelValues("cache_insert").Inc()
		time.Sleep(time.Duration(i+1) * time.Second)
	}
	if err != nil {
		p.logger.Error(fmt.Sprintf("Cache insert failed after retries: %v", err))
	}

	metrics.BatchInserts.Inc()
	metrics.MessagesProcessed.Add(float64(len(batch)))
	metrics.ProcessingLatency.Observe(time.Since(start).Seconds())
}
