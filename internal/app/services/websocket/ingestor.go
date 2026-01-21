package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"ws_ingestor/internal/app/common/logger"
	"ws_ingestor/internal/app/constants"
	"ws_ingestor/internal/app/metrics"
	"ws_ingestor/internal/app/models"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type Ingestor struct {
	url     string
	apiKey  string
	out     chan<- models.MarketData
	symbols []string
	logger  *logrus.Logger
}

func New(url, apiKey string, out chan<- models.MarketData, symbols []string) *Ingestor {
	return &Ingestor{url: url, apiKey: apiKey, out: out, symbols: symbols, logger: logger.GetLogger()}
}

func (c *Ingestor) Start(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			c.logger.WithField("panic", r).Error("WebSocket Ingestor panicked")
		}
	}()
	backoff := time.Second

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		header := http.Header{}
		header.Set("x-api-key", c.apiKey)

		conn, _, err := websocket.DefaultDialer.Dial(c.url, header)
		if err != nil {
			c.logger.Error(fmt.Sprintf("WS connect failed: %v", err))
			time.Sleep(backoff)
			if backoff < 30*time.Second {
				backoff *= 2
			}
			continue
		}

		c.logger.Info("WebSocket connected")
		backoff = time.Second

		// Send subscription message after connecting
		subscriptionMsg := map[string]interface{}{
			"event":   "subscribe",
			"symbols": c.symbols,
		}
		msgBytes, err := json.Marshal(subscriptionMsg)
		if err != nil {
			c.logger.Error(fmt.Sprintf("Failed to marshal subscription message: %v", err))
			conn.Close()
			continue
		}
		if err := conn.WriteMessage(websocket.TextMessage, msgBytes); err != nil {
			c.logger.Error(fmt.Sprintf("Failed to send subscription message: %v", err))
			conn.Close()
			continue
		}
		c.logger.Info("Subscription message sent")

		c.readLoop(ctx, conn)
	}
}

func (c *Ingestor) readLoop(ctx context.Context, conn *websocket.Conn) {
	defer func() {
		if r := recover(); r != nil {
			c.logger.WithField("panic", r).Error("WebSocket readLoop panicked")
		}
		conn.Close()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		_, msg, err := conn.ReadMessage()
		if err != nil {
			c.logger.Error(fmt.Sprintf("WS read error: %v", err))
			return
		}

		var data models.MarketData
		if err := json.Unmarshal(msg, &data); err != nil {
			c.logger.Error(fmt.Sprintf("Failed to unmarshal message: %v", err))
			metrics.ErrorsTotal.WithLabelValues("unmarshal").Inc()
			continue
		}
		if err := data.Validate(); err != nil {
			c.logger.Error(fmt.Sprintf("Invalid market data: %v", err))
			metrics.ErrorsTotal.WithLabelValues("validation").Inc()
			continue
		}
		// Set exchange based on symbol
		allSymbols := constants.GetAllSymbols()
		if exch, ok := allSymbols[data.Name]; ok {
			data.Exchange = exch
		} else {
			data.Exchange = "unknown"
		}

		metrics.MessagesReceived.Inc()
		c.out <- data
	}
}
