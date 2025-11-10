package sse

import (
	"bufio"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Represents an SSE event
type Event struct {
	ID        string
	Event     string
	Data      string
	Retry     string
	Timestamp time.Time
	OnChannel *manager
}

// Writes a SSE event to the writer according to the SSE standard
func (e *Event) Flush(w *bufio.Writer) error {
	if e.ID != "" {
		fmt.Fprintf(w, "id: %s\n", e.ID)
	}
	if e.Retry != "" {
		fmt.Fprintf(w, "retry: %s\n", e.Retry)
	}
	fmt.Fprintf(w, "event: %s\n", e.Event)
	fmt.Fprintf(w, "data: %s\n\n", e.Data)

	return w.Flush()
}

// Executes the handlers registered for this event
func (e *Event) FireEventHandlers(ctx *fiber.Ctx) {
	channel := e.OnChannel
	if handlers, ok := channel.EventHandlers[e.Event]; ok {
		for _, handler := range handlers {
			handler(ctx, channel.Name, e)
		}
	}
}
