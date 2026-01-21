package websocket

import (
	"sync"
	"ws_ingestor/internal/app/dto"

	"github.com/gorilla/websocket"
)

type Client struct {
	ID     string
	conns  map[*websocket.Conn]struct{}
	mu     sync.Mutex
	Config *dto.ClientConfig
}

func (c *Client) addConn(conn *websocket.Conn) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.conns[conn] = struct{}{}
}

func (c *Client) removeConn(conn *websocket.Conn) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.conns, conn)
}

func (c *Client) isEmpty() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.conns) == 0
}
