package websocket

import (
	"context"
	"fmt"
	"maps"
	"net/http"
	"sync"
	"time"

	"ws_ingestor/internal/app/common/logger"
	"ws_ingestor/internal/app/dto"
	"ws_ingestor/internal/app/models"
	"ws_ingestor/internal/app/services/storage"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type Server struct {
	addr     string
	store    *storage.Store
	cache    *storage.CacheService
	logger   *logrus.Logger
	upgrader websocket.Upgrader
	clients  sync.Map // map[*websocket.Conn]bool
}

func NewServer(addr string, cache *storage.CacheService, store *storage.Store) *Server {
	return &Server{
		addr:   addr,
		cache:  cache,
		store:  store,
		logger: logger.GetLogger(),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for demo
			},
		},
	}
}

func (s *Server) Start(ctx context.Context) {
	// Start broadcaster with stream reading
	go s.broadcaster(ctx)

	http.HandleFunc("/ws", s.handleConnection)
	s.logger.Info("Starting WebSocket server on " + s.addr)
	if err := http.ListenAndServe(s.addr, nil); err != nil {
		s.logger.Fatal("Failed to start WebSocket server: ", err)
	}
}

func (s *Server) handleConnection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	apiKey := r.Header.Get("X-API-Key")
	if apiKey == "" {
		http.Error(w, "missing api key", http.StatusUnauthorized)
		return
	}

	clientID, err := s.store.ValidateApiKey(ctx, apiKey)
	if err != nil {
		http.Error(w, "invalid api key", http.StatusUnauthorized)
		return
	}

	clientConfig, err := s.store.GetClientConfig(ctx, clientID)
	if err != nil {
		s.logger.Error("Failed to get client config: ", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	fmt.Println("client config retrienved ============= ", clientConfig)

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := s.getOrCreateClient(clientID, clientConfig)
	client.addConn(conn)

	go s.readPump(clientID, conn)
}

func (s *Server) broadcaster(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second) // Send data every second
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Disconnected from the websocket server")
			return
		case <-ticker.C:
			// Fetch all current data from Redis
			allData, err := s.cache.GetAllData(ctx)
			if err != nil {
				s.logger.Error("Failed to get all data: ", err)
				continue
			}

			// Send all data to connected clients
			s.clients.Range(func(_, value interface{}) bool {
				client := value.(*Client)
				client.mu.Lock()
				for conn := range client.conns {
					for _, item := range allData {
						flat := normalizeMarketData(item)
						if client.Config != nil && client.Config.Symbols != nil {
							if cfg, ok := client.Config.Symbols[item.Name]; ok {
								flat = transformFlat(flat, &cfg)
							}
						}
						if err := conn.WriteJSON(flat); err != nil {
							conn.Close()
							delete(client.conns, conn)
							break
						}
					}
				}
				client.mu.Unlock()

				return true
			})
		}
	}
}

func (s *Server) getOrCreateClient(clientID string, clientConfig *dto.ClientConfig) *Client {
	val, ok := s.clients.Load(clientID)
	if ok {
		return val.(*Client)
	}

	client := &Client{
		ID:     clientID,
		conns:  make(map[*websocket.Conn]struct{}),
		Config: clientConfig,
	}

	actual, _ := s.clients.LoadOrStore(clientID, client)
	return actual.(*Client)
}

func (s *Server) readPump(clientID string, conn *websocket.Conn) {
	defer func() {
		conn.Close()

		if val, ok := s.clients.Load(clientID); ok {
			client := val.(*Client)
			client.removeConn(conn)

			if client.isEmpty() {
				s.clients.Delete(clientID)
			}
		}
	}()

	conn.SetReadLimit(512)
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}

func normalizeMarketData(item models.MarketData) dto.FlatMarketData {
	out := dto.FlatMarketData{}

	// Flatten data block
	if inner, ok := item.Data["data"].(map[string]interface{}); ok {
		maps.Copy(out, inner)
	}
	out["symbol"] = item.Name
	out["timestamp"] = item.Timestamp
	out["exchange"] = item.Exchange

	return out
}

func transformFlat(data dto.FlatMarketData, cfg *dto.SymbolConfig) dto.FlatMarketData {

	// Value transforms
	for field, rule := range cfg.ValueRules {
		if v, ok := data[field].(float64); ok {
			data[field] = applyValueRule(v, rule)
		}
	}

	// Rename fields
	for oldKey, newKey := range cfg.RenameFields {
		if v, ok := data[oldKey]; ok {
			data[newKey] = v
			delete(data, oldKey)
		}
	}

	// Remove fields
	for _, field := range cfg.RemoveFields {
		delete(data, field)
	}

	// Hard overrides
	for k, v := range cfg.OverrideFields {
		if k == "timestamp" && v == "current" {
			data[k] = time.Now().UnixMilli()
		} else {
			data[k] = v
		}
	}

	return data
}

func applyValueRule(num float64, rule dto.ValueRule) float64 {
	switch rule.Op {
	case "add":
		return num + rule.Value
	case "subtract":
		return num - rule.Value
	case "multiply":
		return num * rule.Value
	case "divide":
		if rule.Value != 0 {
			return num / rule.Value
		}
	}
	return num
}
