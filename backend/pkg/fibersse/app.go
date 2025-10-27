package fibersse

import (
	"fmt"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

// Create new Fiber SSE app on `base` path
func New(app *fiber.App, base string) (*FiberSSEApp, error) {
	fiberRouter := app.Group(base, func(c *fiber.Ctx) error {
		c.Set("Cache-Control", "no-cache")
		c.Set("Content-Type", "text/event-stream")
		c.Set("Connection", "keep-alive")
		c.Set("Access-Control-Allow-Origin", "*")
		c.Set("Transfer-Encoding", "chunked")
		return c.Next()
	})

	if err := ValidateBasePath(base); err != nil {
		return nil, err
	}

	fiberApp := &FiberSSEApp{
		Base:     base,
		Router:   &fiberRouter,
		FiberApp: app,
		Channels: make(map[string]*FiberSSEChannel),
	}

	return fiberApp, nil
}

// Create new SSE channel
func (app *FiberSSEApp) CreateChannel(name, base string) (*FiberSSEChannel, error) {
	if err := ValidateBasePath(base); err != nil {
		return nil, err
	}

	if _, exists := app.Channels[name]; exists {
		return nil, fmt.Errorf("[FiberSSE] channel %s already exists", name)
	}

	newChannel := &FiberSSEChannel{
		Name:          name,
		Base:          base,
		Events:        make(chan *FiberSSEEvent, 10),
		ParentSSEApp:  app,
		Handlers:      make(map[string][]FiberSSEEventHandler),
		EventHandlers: make(map[string][]FiberSSEOnEventHandler),
		closed:        false,
	}

	app.Channels[name] = newChannel
	(*app.Router).Get(newChannel.Base, newChannel.Handler)

	return newChannel, nil
}

// List all channels
func (app *FiberSSEApp) ListChannels() map[string]*FiberSSEChannel {
	log.Info("[FiberSSE] Listing Channels...")
	for _, channel := range app.Channels {
		channel.Print()
	}
	return app.Channels
}

// Get channel by name
func (app *FiberSSEApp) GetChannel(name string) *FiberSSEChannel {
	return app.Channels[name]
}

// Closes all SSE channels securely
func (sseApp *FiberSSEApp) Cleanup() error {
	var wg sync.WaitGroup
	var errs []error
	var mu sync.Mutex

	for _, channel := range sseApp.Channels {
		wg.Add(1)
		go func(ch *FiberSSEChannel) {
			defer wg.Done()
			if err := ch.Close(); err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			}
		}(channel)
	}

	wg.Wait()

	if len(errs) > 0 {
		return fmt.Errorf("[FiberSSE] Error during cleanup: %v", errs)
	}

	log.Info("[FiberSSE] All Channels Closed - Cleanup Successful")
	return nil
}
