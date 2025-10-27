package sse

import (
	"bufio"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/valyala/fasthttp"
)

// Handles channel events (connect, disconnect)
type FiberSSEEventHandler func(ctx *fiber.Ctx, name string)

// Handles specific events
type FiberSSEOnEventHandler func(ctx *fiber.Ctx, name string, sseEvent *Event)

// Configuration options for Manager
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

// DefaultConfig returns default configuration
func DefaultConfig() ManagerConfig {
	return ManagerConfig{
		Name:              "",
		BufferSize:        10,
		HeartbeatInterval: 15 * time.Second,
		SendTimeout:       1 * time.Second,
		Debug:             false,
	}
}

// Client représente une connexion SSE individuelle
type Client struct {
	ID        string
	Events    chan *Event
	ConnectAt time.Time
}

type Manager interface {
	GetClientCount() int
	GetClients() []string
	SendEvent(event string, data interface{}) error
	Close() error
	Print()
	Handler(c *fiber.Ctx) error
	FireHandlers(c *fiber.Ctx, event string)
	OnConnect(handlers ...FiberSSEEventHandler) Manager
	OnDisconnect(handlers ...FiberSSEEventHandler) Manager
	OnEvent(eventName string, handlers ...FiberSSEOnEventHandler) Manager
}

// Manager représente un canal SSE avec plusieurs clients
type manager struct {
	Name          string
	Config        ManagerConfig
	clients       map[string]*Client
	clientsMux    sync.RWMutex
	Handlers      map[string][]FiberSSEEventHandler
	EventHandlers map[string][]FiberSSEOnEventHandler
	closed        bool
	closedMux     sync.RWMutex
}

// Create new SSE Manager with optional config
// Usage: NewManager(ManagerConfig{Name: "MyChannel"})
// or:    NewManager(ManagerConfig{Name: "MyChannel", BufferSize: 50})
func New(config ...ManagerConfig) Manager {
	// Start with default config
	cfg := DefaultConfig()

	// Merge with provided config if any
	if len(config) > 0 {
		userCfg := config[0]

		// Name is required
		if userCfg.Name != "" {
			cfg.Name = userCfg.Name
		}

		// Override defaults only if explicitly set (non-zero values)
		if userCfg.BufferSize > 0 {
			cfg.BufferSize = userCfg.BufferSize
		}
		if userCfg.HeartbeatInterval > 0 {
			cfg.HeartbeatInterval = userCfg.HeartbeatInterval
		}
		if userCfg.SendTimeout > 0 {
			cfg.SendTimeout = userCfg.SendTimeout
		}
		if userCfg.Debug == true {
			cfg.Debug = userCfg.Debug
		}
	}

	if cfg.Name == "" {
		log.Warn("[SSEManager] Manager created without a name")
	}

	return &manager{
		Name:          cfg.Name,
		Config:        cfg,
		clients:       make(map[string]*Client),
		Handlers:      make(map[string][]FiberSSEEventHandler),
		EventHandlers: make(map[string][]FiberSSEOnEventHandler),
		closed:        false,
	}
}

// GetClientCount returns the number of connected clients
func (m *manager) GetClientCount() int {
	m.clientsMux.RLock()
	defer m.clientsMux.RUnlock()
	return len(m.clients)
}

// GetClients returns a copy of all connected client IDs
func (m *manager) GetClients() []string {
	m.clientsMux.RLock()
	defer m.clientsMux.RUnlock()

	clients := make([]string, 0, len(m.clients))
	for id := range m.clients {
		clients = append(clients, id)
	}
	return clients
}

// IsClosed returns true if the manager is closed
func (m *manager) IsClosed() bool {
	m.closedMux.RLock()
	defer m.closedMux.RUnlock()
	return m.closed
}

// Ajoute un client
func (m *manager) addClient(client *Client) {
	m.clientsMux.Lock()
	defer m.clientsMux.Unlock()
	m.clients[client.ID] = client

	if m.Config.Debug {
		log.Infof("[SSEManager] Client %s added to channel %s (total: %d)", client.ID, m.Name, len(m.clients))
	}
}

// Supprime un client
func (m *manager) removeClient(clientID string) {
	m.clientsMux.Lock()
	defer m.clientsMux.Unlock()
	if client, ok := m.clients[clientID]; ok {
		close(client.Events)
		delete(m.clients, clientID)

		if m.Config.Debug {
			log.Infof("[SSEManager] Client %s removed from channel %s (remaining: %d)", clientID, m.Name, len(m.clients))
		}
	}
}

// SendEvent sends event to all connected clients (marshals data to JSON)
func (m *manager) SendEvent(event string, data interface{}) error {
	m.closedMux.RLock()
	defer m.closedMux.RUnlock()

	if m.closed {
		return fmt.Errorf("[SSEManager] channel %s is closed", m.Name)
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("[SSEManager] failed to marshal data: %w", err)
	}

	sseEvent := &Event{
		Timestamp: time.Now(),
		Event:     event,
		Data:      string(jsonData),
		OnChannel: m,
	}

	return m.broadcastEvent(sseEvent)
}

// broadcastEvent broadcasts an event to all clients
func (m *manager) broadcastEvent(sseEvent *Event) error {
	m.clientsMux.RLock()
	defer m.clientsMux.RUnlock()

	if m.Config.Debug {
		log.Infof("[SSEManager] Broadcasting event '%s' to %d clients", sseEvent.Event, len(m.clients))
	}

	var failedClients []string
	for clientID, client := range m.clients {
		select {
		case client.Events <- sseEvent:
			if m.Config.Debug {
				log.Debugf("[SSEManager] Event sent to client %s", clientID)
			}
		case <-time.After(m.Config.SendTimeout):
			failedClients = append(failedClients, clientID)
			if m.Config.Debug {
				log.Warnf("[SSEManager] timeout sending event to client %s", clientID)
			}
		}
	}

	if len(failedClients) > 0 && m.Config.Debug {
		return fmt.Errorf("[SSEManager] failed to send to %d clients: %v", len(failedClients), failedClients)
	}

	return nil
}

// Close closes the manager and all client connections
func (m *manager) Close() error {
	m.closedMux.Lock()
	defer m.closedMux.Unlock()

	if m.closed {
		return fmt.Errorf("[SSEManager] channel %s already closed", m.Name)
	}
	m.closed = true

	m.clientsMux.Lock()
	defer m.clientsMux.Unlock()

	// Fermer tous les clients
	for _, client := range m.clients {
		close(client.Events)
	}
	m.clients = make(map[string]*Client)

	if m.Config.Debug {
		log.Infof("[SSEManager] Manager %s closed", m.Name)
	}

	return nil
}

// Print SSE channel infos
func (m *manager) Print() {
	log.Info("[SSEManager] ==CHANNEL CREATED==\nName: %s\n", m.Name)
}

// Handler pour gérer une connexion SSE
func (m *manager) Handler(c *fiber.Ctx) error {
	// Vérifier si le manager est fermé
	if m.IsClosed() {
		return c.Status(fiber.StatusServiceUnavailable).SendString("SSE channel is closed")
	}

	c.Set("Cache-Control", "no-cache")
	c.Set("Content-Type", "text/event-stream")
	c.Set("Connection", "keep-alive")
	c.Set("Access-Control-Allow-Origin", "*")
	c.Set("Transfer-Encoding", "chunked")
	c.Set("X-Accel-Buffering", "no") // Disable nginx buffering

	// Créer un nouveau client pour cette connexion
	client := &Client{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Events:    make(chan *Event, m.Config.BufferSize),
		ConnectAt: time.Now(),
	}
	m.addClient(client)

	c.Status(fiber.StatusOK).Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		// Fire OnConnect Event Handlers
		m.FireHandlers(c, "connect")

		// Assurer le nettoyage à la fin
		defer func() {
			duration := time.Since(client.ConnectAt)
			if m.Config.Debug {
				log.Infof("[SSEManager] Client %s disconnected after %v", client.ID, duration)
			}
			m.FireHandlers(c, "disconnect")
			m.removeClient(client.ID)
		}()

		// Ticker pour envoyer des heartbeats si activé
		var ticker *time.Ticker
		var tickerChan <-chan time.Time

		if m.Config.HeartbeatInterval > 0 {
			ticker = time.NewTicker(m.Config.HeartbeatInterval)
			tickerChan = ticker.C
			defer ticker.Stop()
		}

		// Boucle de lecture des événements
		for {
			select {
			case <-tickerChan:
				// Envoyer un commentaire SSE comme heartbeat
				if _, err := w.WriteString(": heartbeat\n\n"); err != nil {
					if m.Config.Debug {
						log.Infof("[SSEManager] Client %s disconnected (heartbeat write failed)", client.ID)
					}
					return
				}
				if err := w.Flush(); err != nil {
					if m.Config.Debug {
						log.Infof("[SSEManager] Client %s disconnected (heartbeat flush failed)", client.ID)
					}
					return
				}

			case event, ok := <-client.Events:
				if !ok {
					if m.Config.Debug {
						log.Infof("[SSEManager] Event channel closed for client %s", client.ID)
					}
					return
				}

				// Fire event handlers
				if handlers, ok := m.EventHandlers[event.Event]; ok {
					for _, handler := range handlers {
						handler(c, m.Name, event)
					}
				}

				if err := event.Flush(w); err != nil {
					if m.Config.Debug {
						log.Warnf("[SSEManager] Error while flushing on channel %s: %v. Closing connection.\n", m.Name, err)
					}
					return
				}
			}
		}
	}))

	return nil
}

// Executes handlers synchronously with the valid context
func (m *manager) FireHandlers(c *fiber.Ctx, event string) {
	if handlers, ok := m.Handlers[event]; ok {
		for _, handler := range handlers {
			handler(c, m.Name)
		}
	}
}

// Registers handlers for connection
func (m *manager) OnConnect(handlers ...FiberSSEEventHandler) Manager {
	if _, ok := m.Handlers["connect"]; !ok {
		m.Handlers["connect"] = []FiberSSEEventHandler{}
	}
	m.Handlers["connect"] = append(m.Handlers["connect"], handlers...)
	return m
}

// Registers handlers for disconnection
func (m *manager) OnDisconnect(handlers ...FiberSSEEventHandler) Manager {
	if _, ok := m.Handlers["disconnect"]; !ok {
		m.Handlers["disconnect"] = []FiberSSEEventHandler{}
	}
	m.Handlers["disconnect"] = append(m.Handlers["disconnect"], handlers...)
	return m
}

// Registers handlers for a specific event
func (m *manager) OnEvent(eventName string, handlers ...FiberSSEOnEventHandler) Manager {
	if _, ok := m.EventHandlers[eventName]; !ok {
		m.EventHandlers[eventName] = []FiberSSEOnEventHandler{}
	}
	m.EventHandlers[eventName] = append(m.EventHandlers[eventName], handlers...)
	return m
}

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
