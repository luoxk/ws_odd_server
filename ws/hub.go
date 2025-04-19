package ws

import (
	"github.com/gorilla/websocket"
	"sync"
)

type ClientHub struct {
	mu      sync.RWMutex
	clients map[uint64]*Client
	nextID  uint64
}

func NewClientHub() *ClientHub {
	return &ClientHub{
		clients: make(map[uint64]*Client),
		nextID:  1,
	}
}

// 添加客户端
func (h *ClientHub) Add(conn *websocket.Conn) *Client {
	h.mu.Lock()
	defer h.mu.Unlock()
	id := h.nextID
	h.nextID++

	client := &Client{
		ID:   id,
		Conn: conn,
		Send: make(chan []byte, 64),
	}
	h.clients[id] = client
	return client
}

// 移除客户端
func (h *ClientHub) Remove(id uint64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, id)
}

// 遍历所有客户端
func (h *ClientHub) ForEach(fn func(id uint64, c *Client)) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for id, c := range h.clients {
		fn(id, c)
	}
}

// 广播消息给所有客户端
func (h *ClientHub) Broadcast(message []byte) {
	h.ForEach(func(_ uint64, c *Client) {
		select {
		case c.Send <- message:
		default:
			// 发送缓冲满，关闭连接（可选）
		}
	})
}
