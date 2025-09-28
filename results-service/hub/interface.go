package hub

// Client defines the interface for a WebSocket client connection.
// This abstraction allows the hub to manage clients without knowing the
// underlying WebSocket implementation.
type Client interface {
	WriteMessage(messageType int, data []byte) error
	ReadMessage() (messageType int, p []byte, err error)
	Close() error
}

// Runnable defines an interface for a component that can be started.
type Runnable interface {
	Run()
}

// ClientManager defines the interface for managing client connections.
type ClientManager interface {
	Register(client Client)
	Unregister(client Client)
}

// MessageBroadcaster defines the interface for broadcasting messages.
type MessageBroadcaster interface {
	Broadcast(message []byte)
}

// Hub defines the complete interface for the WebSocket message hub.
// It combines running, client management, and message broadcasting.
type Hub interface {
	Runnable
	ClientManager
	MessageBroadcaster
}
