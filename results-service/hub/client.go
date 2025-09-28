package hub

import (
	"github.com/gorilla/websocket"
)

// websocketClient is a concrete implementation of the Client interface that wraps
// a gorilla/websocket connection. This is kept private to the hub package.
type websocketClient struct {
	conn *websocket.Conn
}

// NewWebsocketClient creates a new client that wraps the given connection.
func NewWebsocketClient(conn *websocket.Conn) Client {
	return &websocketClient{conn: conn}
}

func (c *websocketClient) WriteMessage(messageType int, data []byte) error {
	return c.conn.WriteMessage(messageType, data)
}

func (c *websocketClient) ReadMessage() (int, []byte, error) {
	return c.conn.ReadMessage()
}

func (c *websocketClient) Close() error {
	return c.conn.Close()
}
