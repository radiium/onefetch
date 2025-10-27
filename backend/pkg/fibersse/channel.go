package fibersse

import (
	"bufio"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/valyala/fasthttp"
)

// Send SSE event
func (channel *FiberSSEChannel) SendEvent(event, data string) error {
	if channel.closed {
		return fmt.Errorf("[FiberSSE] channel %s is closed", channel.Name)
	}

	sseEvent := &FiberSSEEvent{
		Timestamp: time.Now(),
		Event:     event,
		Data:      data,
		OnChannel: channel,
	}

	select {
	case channel.Events <- sseEvent:
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("[FiberSSE] timeout sending event to channel %s", channel.Name)
	}
}

// Send SSE event with custom ID
func (channel *FiberSSEChannel) SendEventWithID(id, event, data string) error {
	if channel.closed {
		return fmt.Errorf("[FiberSSE] channel %s is closed", channel.Name)
	}

	sseEvent := &FiberSSEEvent{
		Timestamp: time.Now(),
		ID:        id,
		Event:     event,
		Data:      data,
		OnChannel: channel,
	}

	select {
	case channel.Events <- sseEvent:
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("[FiberSSE] timeout sending event to channel %s", channel.Name)
	}
}

// Close SSE channel securely
func (channel *FiberSSEChannel) Close() error {
	if channel.closed {
		return fmt.Errorf("[FiberSSE] channel %s already closed", channel.Name)
	}
	channel.closed = true
	close(channel.Events)
	return nil
}

// Print SSE channel infos
func (c *FiberSSEChannel) Print() {
	log.Info("[FiberSSE] ==CHANNEL CREATED==\nName: %s\nRoute Endpoint: %s\n===================\n",
		c.Name, c.ParentSSEApp.Base+c.Base)
}

// ServeHTTP gère la connexion SSE pour le canal
func (fChan *FiberSSEChannel) ServeHTTP1(c *fiber.Ctx) error {

	log.Info("[FiberSSE] ServeHTTP start")
	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		// Capturer les références nécessaires AVANT de lancer les goroutines
		ctx := c.Context()

		// Fire OnConnect Event Handlers
		fChan.fireHandlersSync(c, "connect")

		for {
			select {
			case event, more := <-fChan.Events:
				if !more {
					log.Info("[FiberSSE] disconnect !more ")
					fChan.fireHandlersSync(c, "disconnect")
					return
				}

				// Fire event handlers dans la même goroutine
				if handlers, ok := fChan.EventHandlers[event.Event]; ok {
					for _, handler := range handlers {
						log.Info("[FiberSSE] handler event ")
						handler(c, fChan, event)
					}
				}

				if err := event.Flush(w); err != nil {
					log.Info("[FiberSSE] disconnect err := event.Flush(w)")
					fChan.fireHandlersSync(c, "disconnect")
					return
				}
			case <-ctx.Done():
				log.Info("[FiberSSE] disconnect ctx.Done()")
				fChan.fireHandlersSync(c, "disconnect")
				return
			}
		}
	})
	return nil
}

// Manages the SSE connection for the channel
func (fChan *FiberSSEChannel) Handler(c *fiber.Ctx) error {
	c.Status(fiber.StatusOK).Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		// Fire OnConnect Event Handlers
		fChan.fireHandlersSync(c, "connect")
		defer fChan.fireHandlersSync(c, "disconnect")

		for event := range fChan.Events {
			// Fire event handlers
			if handlers, ok := fChan.EventHandlers[event.Event]; ok {
				for _, handler := range handlers {
					handler(c, fChan, event)
				}
			}

			if err := event.Flush(w); err != nil {
				log.Warnf("[FiberSSE] Error while flushing on channel %s: %v. Closing connection.\n", fChan.Name, err)
				return
			}
		}
	}))

	return nil
}

// Executes handlers synchronously with the valid context
func (channel *FiberSSEChannel) fireHandlersSync(c *fiber.Ctx, event string) {
	if handlers, ok := channel.Handlers[event]; ok {
		for _, handler := range handlers {
			handler(c, channel)
		}
	}
}
