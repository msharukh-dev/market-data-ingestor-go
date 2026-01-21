package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	MessagesReceived = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ws_ingestor_messages_received_total",
		Help: "Total number of messages received from websocket",
	})

	MessagesProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ws_ingestor_messages_processed_total",
		Help: "Total number of messages processed",
	})

	BatchInserts = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ws_ingestor_batch_inserts_total",
		Help: "Total number of batch inserts",
	})

	ErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ws_ingestor_errors_total",
		Help: "Total number of errors",
	}, []string{"type"})

	ProcessingLatency = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "ws_ingestor_processing_latency_seconds",
		Help:    "Latency of processing batches",
		Buckets: prometheus.DefBuckets,
	})
)
