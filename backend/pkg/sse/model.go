package sse

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

// OnStatusEventHandler Handles channel status events (connect, disconnect)
type OnStatusEventHandler func(ctx *fiber.Ctx, name string)

// OnEventHandler Handles specific events
type OnEventHandler func(ctx *fiber.Ctx, name string, sseEvent *Event)

// ManagerConfig Configuration options for Manager
type ManagerConfig struct {
	// Name of the SSE channel (required)
	Name string
	// Buffer size for client event channels (default: 10)
	BufferSize int
	// Heartbeat interval (default: 15s, 0 to disable)
	HeartbeatInterval time.Duration
	// Timeout for sending events to slow clients (default: 1s)
	SendTimeout time.Duration
	// Enable debug logs
	Debug bool
}

// Client represents an individual SSE connection
type Client struct {
	ID        string
	Events    chan *Event
	ConnectAt time.Time
}
