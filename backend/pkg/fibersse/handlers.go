package fibersse

import (
	"github.com/gofiber/fiber/v2"
)

// Executes handlers for a channel SSE event
func (channel *FiberSSEChannel) FireHandlers(fiberCtx *fiber.Ctx, event string) {
	if handlers, ok := channel.Handlers[event]; ok {
		for _, handler := range handlers {
			handler(fiberCtx, channel)
		}
	}
}

// Registers handlers for connection
func (channel *FiberSSEChannel) OnConnect(handlers ...FiberSSEEventHandler) *FiberSSEChannel {
	if _, ok := channel.Handlers["connect"]; !ok {
		channel.Handlers["connect"] = []FiberSSEEventHandler{}
	}
	channel.Handlers["connect"] = append(channel.Handlers["connect"], handlers...)
	return channel
}

// Registers handlers for disconnection
func (channel *FiberSSEChannel) OnDisconnect(handlers ...FiberSSEEventHandler) *FiberSSEChannel {
	if _, ok := channel.Handlers["disconnect"]; !ok {
		channel.Handlers["disconnect"] = []FiberSSEEventHandler{}
	}
	channel.Handlers["disconnect"] = append(channel.Handlers["disconnect"], handlers...)
	return channel
}

// Registers handlers for a specific event
func (channel *FiberSSEChannel) OnEvent(eventName string, handlers ...FiberSSEOnEventHandler) *FiberSSEChannel {
	if _, ok := channel.EventHandlers[eventName]; !ok {
		channel.EventHandlers[eventName] = []FiberSSEOnEventHandler{}
	}
	channel.EventHandlers[eventName] = append(channel.EventHandlers[eventName], handlers...)
	return channel
}
