package fibersse

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

// Represents an SSE event
type FiberSSEEvent struct {
	Timestamp time.Time
	ID        string
	Event     string
	Data      string
	Retry     string
	OnChannel *FiberSSEChannel
}

// Handles channel events (connect, disconnect)
type FiberSSEEventHandler func(ctx *fiber.Ctx, sseChannel *FiberSSEChannel)

// Handles specific events
type FiberSSEOnEventHandler func(ctx *fiber.Ctx, sseChannel *FiberSSEChannel, sseEvent *FiberSSEEvent)

// Defines the interface for events
type FiberSSEEvents interface {
	OnConnect(handlers ...FiberSSEEventHandler)
	OnDisconnect(handlers ...FiberSSEEventHandler)
	OnEvent(eventName string, handlers ...FiberSSEOnEventHandler)
	FireOnEventHandlers(fiberCtx *fiber.Ctx, event string)
}

// Represents an SSE channel
type FiberSSEChannel struct {
	FiberSSEEvents
	Name          string
	Base          string
	Events        chan *FiberSSEEvent
	ParentSSEApp  *FiberSSEApp
	Handlers      map[string][]FiberSSEEventHandler
	EventHandlers map[string][]FiberSSEOnEventHandler
	closed        bool
}

// Interface for SSE application
type IFiberSSEApp interface {
	ServeHTTP(ctx *fiber.Ctx) error
	CreateChannel(name, base string) (*FiberSSEChannel, error)
	ListChannels() map[string]*FiberSSEChannel
	GetChannel(name string) *FiberSSEChannel
}

// Represents the main SSE application
type FiberSSEApp struct {
	IFiberSSEApp
	Base     string
	Router   *fiber.Router
	Channels map[string]*FiberSSEChannel
	FiberApp *fiber.App
}
