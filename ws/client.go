package ws

import (
	"github.com/gorilla/websocket"
)

type Client struct {
	ID   uint64
	Conn *websocket.Conn
	Send chan []byte
}

func StartWriter(c *Client) {
	for msg := range c.Send {
		if err := c.Conn.WriteMessage(websocket.BinaryMessage, msg); err != nil {
			break
		}
	}
}
